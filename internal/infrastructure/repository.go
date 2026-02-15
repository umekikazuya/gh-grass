package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/umekikazuya/gh-grass/internal/domain"
	"golang.org/x/oauth2"
)

type GitHubRepository struct {
	client *githubv4.Client
}

// NewGitHubRepository は指定した GitHub アクセストークンを使用して認証済みの GitHub GraphQL クライアントを初期化し、GitHubRepository を返します。
// token は GitHub API にアクセスするためのアクセストークン（例: Personal Access Token）です。
func NewGitHubRepository(token string) *GitHubRepository {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	return &GitHubRepository{client: client}
}

func (r *GitHubRepository) GetUser(ctx context.Context, login string) (*domain.User, error) {
	if login == "" {
		var q struct {
			Viewer struct {
				Login githubv4.String
			}
		}
		err := r.client.Query(ctx, &q, nil)
		if err != nil {
			return nil, err
		}
		return &domain.User{Login: string(q.Viewer.Login)}, nil
	}

	var q struct {
		User struct {
			Login githubv4.String
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]interface{}{
		"login": githubv4.String(login),
	}
	err := r.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}
	return &domain.User{Login: string(q.User.Login)}, nil
}

func (r *GitHubRepository) ListOrgMembers(ctx context.Context, orgName string) ([]domain.User, error) {
	var q struct {
		Organization struct {
			MembersWithRole struct {
				Nodes []struct {
					Login githubv4.String
				}
			} `graphql:"membersWithRole(first: 100)"`
		} `graphql:"organization(login: $org)"`
	}
	variables := map[string]interface{}{
		"org": githubv4.String(orgName),
	}
	err := r.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}

	var users []domain.User
	for _, node := range q.Organization.MembersWithRole.Nodes {
		users = append(users, domain.User{Login: string(node.Login)})
	}
	return users, nil
}

func (r *GitHubRepository) GetContributions(ctx context.Context, username string, date time.Time) (*domain.Contribution, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	// API requires generic DateTime, so we use the start and end of the day roughly,
	// or just the same day with time 00:00:00 to 23:59:59.
	// Actually, contributionCalendar aggregates by day.
	// We can request a range covering the specific date.

	// Ensure date is in UTC or appropriate logic.
	// The API treats dates based on UTC usually or viewer timezone.
	// Let's set the range to cover the target date.
	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour).Add(-1 * time.Second)

	var q struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					Weeks []struct {
						ContributionDays []struct {
							Date              githubv4.String // "YYYY-MM-DD"
							ContributionCount githubv4.Int
						}
					}
				}
			} `graphql:"contributionsCollection(from: $from, to: $to)"`
		} `graphql:"user(login: $user)"`
	}

	variables := map[string]interface{}{
		"user": githubv4.String(username),
		"from": githubv4.DateTime{Time: from},
		"to":   githubv4.DateTime{Time: to},
	}

	err := r.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}

	targetDateStr := date.Format("2006-01-02")

	for _, week := range q.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			if string(day.Date) == targetDateStr {
				return &domain.Contribution{
					Date:  date,
					Count: int(day.ContributionCount),
				}, nil
			}
		}
	}

	// If not found in the returned calendar (should not happen if date is valid)
	return &domain.Contribution{Date: date, Count: 0}, nil
}
