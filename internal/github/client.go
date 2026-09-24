// Package github は GitHub GraphQL API への接続を担うアダプタ。取得結果は app の型に変換して返す。
package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/umekikazuya/gh-grass/internal/app"
)

// GraphQL の変数の型名は Go の型名から決まるため、スキーマと同じ名前の型を用意する。
type (
	String   string
	Int      int
	DateTime struct{ time.Time }
)

// membersPageSize は Organization のメンバーを 1 回に取得する件数（API の上限）。
const membersPageSize = 100

// Client は GitHub GraphQL API のクライアント。
type Client struct {
	gql *api.GraphQLClient
}

// New は gh CLI の認証情報とホスト設定を使うクライアントを返す。
func New() (*Client, error) {
	c, err := NewWithOptions(api.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w (run `gh auth login` if you are not logged in)", err)
	}
	return c, nil
}

// NewWithOptions は opts で設定したクライアントを返す。テストで通信先を差し替えるために使う。
func NewWithOptions(opts api.ClientOptions) (*Client, error) {
	base := opts.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	opts.Transport = unauthorizedAsError{base: base}

	gql, err := api.NewGraphQLClient(opts)
	if err != nil {
		return nil, fmt.Errorf("create GraphQL client: %w", err)
	}
	return &Client{gql: gql}, nil
}

// unauthorizedAsError は 401 のレスポンスを app.ErrUnauthorized に変換する。
// GraphQL クライアントは 200 以外を文字列だけのエラーにしてしまうため、その手前で種別を付ける。
type unauthorizedAsError struct {
	base http.RoundTripper
}

func (t unauthorizedAsError) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(r)
	if err != nil {
		return nil, err //nolint:wrapcheck // RoundTripper はエラーを加工せずに返す
	}
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		return nil, app.ErrUnauthorized
	}
	return resp, nil
}

// Viewer は認証中のユーザーのログイン名を返す。
func (c *Client) Viewer(ctx context.Context) (string, error) {
	var q struct {
		Viewer struct {
			Login string
		}
	}
	if err := c.gql.QueryWithContext(ctx, "Viewer", &q, nil); err != nil {
		return "", classify(fmt.Errorf("query viewer: %w", err))
	}
	return q.Viewer.Login, nil
}

// Calendar は login の r 期間のコントリビューションを返す。r は 1 年以内である必要がある。
func (c *Client) Calendar(ctx context.Context, login string, r app.DateRange) (app.Calendar, error) {
	var q struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					Weeks []struct {
						ContributionDays []struct {
							Date              string
							ContributionCount int
						}
					}
				}
			} `graphql:"contributionsCollection(from: $from, to: $to)"`
		} `graphql:"user(login: $login)"`
	}
	vars := map[string]any{
		"login": String(login),
		"from":  DateTime{r.From.Time()},
		// To の日を丸ごと含めるため、その日の終わりまでを指定する。
		"to": DateTime{r.To.AddDays(1).Time().Add(-time.Second)},
	}
	if err := c.gql.QueryWithContext(ctx, "Calendar", &q, vars); err != nil {
		return app.Calendar{}, classify(fmt.Errorf("query contributions of %q (%s to %s): %w", login, r.From, r.To, err))
	}

	counts := map[app.Date]int{}
	for _, week := range q.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			d, err := app.ParseDate(day.Date)
			if err != nil {
				return app.Calendar{}, fmt.Errorf("parse contribution date: %w", err)
			}
			counts[d] = day.ContributionCount
		}
	}
	return app.NewCalendar(r, counts), nil
}

// OrgMembers は Organization の全メンバーのログイン名を返す。
func (c *Client) OrgMembers(ctx context.Context, org string) ([]string, error) {
	var members []string
	var cursor *String
	for {
		var q struct {
			Organization struct {
				MembersWithRole struct {
					Nodes []struct {
						Login string
					}
					PageInfo struct {
						HasNextPage bool
						EndCursor   String
					}
				} `graphql:"membersWithRole(first: $first, after: $cursor)"`
			} `graphql:"organization(login: $org)"`
		}
		vars := map[string]any{
			"org":    String(org),
			"first":  Int(membersPageSize),
			"cursor": cursor,
		}
		if err := c.gql.QueryWithContext(ctx, "OrgMembers", &q, vars); err != nil {
			return nil, classify(fmt.Errorf("query members of %q: %w", org, err))
		}

		page := q.Organization.MembersWithRole
		for _, n := range page.Nodes {
			members = append(members, n.Login)
		}
		if !page.PageInfo.HasNextPage {
			return members, nil
		}
		next := page.PageInfo.EndCursor
		cursor = &next
	}
}

// classify は API のエラーを app のエラー種別で包む。元のエラーも errors.Is / As で辿れる。
func classify(err error) error {
	var gqlErr *api.GraphQLError
	if errors.As(err, &gqlErr) {
		for _, e := range gqlErr.Errors {
			if e.Type == "NOT_FOUND" {
				return fmt.Errorf("%w: %w", app.ErrNotFound, err)
			}
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %w", app.ErrTimeout, err)
	}
	return err
}
