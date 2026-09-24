package github

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/umekikazuya/gh-grass/internal/app"
)

// request はクライアントが送った GraphQL リクエスト。
type request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// fakeGraphQL は受け取ったリクエストを記録し、用意したレスポンスを順に返す RoundTripper。
type fakeGraphQL struct {
	status    int
	responses []string
	requests  []request
}

func (f *fakeGraphQL) RoundTrip(r *http.Request) (*http.Response, error) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	f.requests = append(f.requests, req)

	body := f.responses[0]
	f.responses = f.responses[1:]
	status := f.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}, nil
}

func newTestClient(t *testing.T, f *fakeGraphQL) *Client {
	t.Helper()
	c, err := NewWithOptions(api.ClientOptions{Host: "github.com", AuthToken: "test-token", Transport: f})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestViewer(t *testing.T) {
	t.Parallel()

	f := &fakeGraphQL{responses: []string{`{"data":{"viewer":{"login":"octocat"}}}`}}
	got, err := newTestClient(t, f).Viewer(context.Background())
	if err != nil || got != "octocat" {
		t.Errorf("Viewer() = %q, %v", got, err)
	}
}

func TestCalendar(t *testing.T) {
	t.Parallel()

	f := &fakeGraphQL{responses: []string{`{"data":{"user":{"contributionsCollection":{"contributionCalendar":{"weeks":[
		{"contributionDays":[{"date":"2026-09-20","contributionCount":0},{"date":"2026-09-21","contributionCount":3}]},
		{"contributionDays":[{"date":"2026-09-27","contributionCount":12}]}
	]}}}}}`}}
	r := app.DateRange{From: app.NewDate(2026, time.September, 20), To: app.NewDate(2026, time.September, 27)}

	cal, err := newTestClient(t, f).Calendar(context.Background(), "octocat", r)
	if err != nil {
		t.Fatal(err)
	}
	if cal.Range != r || cal.Count(app.NewDate(2026, time.September, 21)) != 3 || cal.Count(app.NewDate(2026, time.September, 27)) != 12 {
		t.Errorf("Calendar() = %+v", cal)
	}

	req := f.requests[0]
	if !strings.Contains(req.Query, "$from:DateTime!") || !strings.Contains(req.Query, "$login:String!") {
		t.Errorf("query = %s", req.Query)
	}
	// To の日を丸ごと含める。
	if req.Variables["from"] != "2026-09-20T00:00:00Z" || req.Variables["to"] != "2026-09-27T23:59:59Z" {
		t.Errorf("variables = %v", req.Variables)
	}
}

func TestOrgMembersPaginates(t *testing.T) {
	t.Parallel()

	f := &fakeGraphQL{responses: []string{
		`{"data":{"organization":{"membersWithRole":{"nodes":[{"login":"alice"},{"login":"bob"}],"pageInfo":{"hasNextPage":true,"endCursor":"c1"}}}}}`,
		`{"data":{"organization":{"membersWithRole":{"nodes":[{"login":"carol"}],"pageInfo":{"hasNextPage":false,"endCursor":"c2"}}}}}`,
	}}

	got, err := newTestClient(t, f).OrgMembers(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "alice,bob,carol" {
		t.Errorf("OrgMembers() = %v", got)
	}
	if len(f.requests) != 2 {
		t.Fatalf("requests = %d, want 2", len(f.requests))
	}
	if f.requests[0].Variables["cursor"] != nil || f.requests[1].Variables["cursor"] != "c1" {
		t.Errorf("cursors = %v, %v", f.requests[0].Variables["cursor"], f.requests[1].Variables["cursor"])
	}
	if !strings.Contains(f.requests[0].Query, "$cursor:String") || strings.Contains(f.requests[0].Query, "$cursor:String!") {
		t.Errorf("cursor should be nullable: %s", f.requests[0].Query)
	}
}

func TestErrorClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{
			name: "存在しないユーザー",
			body: `{"data":{"user":null},"errors":[{"type":"NOT_FOUND","path":["user"],"message":"Could not resolve to a User with the login of 'ghost'."}]}`,
			want: app.ErrNotFound,
		},
		{
			name:   "認証エラー",
			status: http.StatusUnauthorized,
			body:   `{"message":"Bad credentials"}`,
			want:   app.ErrUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &fakeGraphQL{status: tt.status, responses: []string{tt.body}}
			r := app.DateRange{From: app.NewDate(2026, time.September, 1), To: app.NewDate(2026, time.September, 2)}
			_, err := newTestClient(t, f).Calendar(context.Background(), "ghost", r)
			if !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}

	t.Run("タイムアウト", func(t *testing.T) {
		t.Parallel()
		if err := classify(context.DeadlineExceeded); !errors.Is(err, app.ErrTimeout) {
			t.Errorf("err = %v", err)
		}
	})
}
