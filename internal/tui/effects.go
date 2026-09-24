package tui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/umekikazuya/gh-grass/internal/app"
)

// requestTimeout は 1 回の API 呼び出しの制限時間。
const requestTimeout = 10 * time.Second

// perform は app の Effect を tea.Cmd に変換する。各 Effect の結果は app が決めた Msg で返す。
func (p *program) perform(effs []app.Effect) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(effs))
	for _, e := range effs {
		cmds = append(cmds, p.performOne(e))
	}
	return tea.Batch(cmds...)
}

func (p *program) performOne(e app.Effect) tea.Cmd {
	switch e := e.(type) {
	case app.FetchViewer:
		return p.request(func(ctx context.Context) app.Msg {
			login, err := p.client.Viewer(ctx)
			return app.GotViewer{Login: login, Err: err}
		})
	case app.FetchCalendar:
		return p.request(func(ctx context.Context) app.Msg {
			cal, err := p.client.Calendar(ctx, e.Login, e.Range)
			return app.GotCalendar{Login: e.Login, Range: e.Range, Calendar: cal, Err: err}
		})
	case app.FetchOrgMembers:
		return p.request(func(ctx context.Context) app.Msg {
			members, err := p.client.OrgMembers(ctx, e.Org)
			return app.GotOrgMembers{Org: e.Org, Members: members, Err: err}
		})
	case app.Quit:
		p.cancel()
		return tea.Quit
	}
	return nil
}

// request は制限時間付きで f を実行する tea.Cmd を返す。終了時には実行中のものも取り消される。
func (p *program) request(f func(ctx context.Context) app.Msg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(p.ctx, requestTimeout)
		defer cancel()
		return f(ctx)
	}
}
