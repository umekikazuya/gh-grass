package tui

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/umekikazuya/gh-grass/internal/app"
)

var today = app.NewDate(2026, time.September, 24)

type fakeClient struct{}

func (fakeClient) Viewer(context.Context) (string, error) { return "octocat", nil }

func (fakeClient) Calendar(_ context.Context, _ string, r app.DateRange) (app.Calendar, error) {
	return app.NewCalendar(r, nil), nil
}

func (fakeClient) OrgMembers(context.Context, string) ([]string, error) {
	return []string{"alice"}, nil
}

func newTestProgram() *program {
	ctx, cancel := context.WithCancel(context.Background())
	m, cmds := app.Init(app.Flags{Today: today})
	return &program{ctx: ctx, cancel: cancel, client: fakeClient{}, model: m, initial: cmds}
}

func TestPerformReturnsResultMsg(t *testing.T) {
	t.Parallel()

	p := newTestProgram()
	r := app.DateRange{From: today, To: today}
	tests := []struct {
		cmd  app.Cmd
		want tea.Msg
	}{
		{app.FetchViewer{}, app.GotViewer{Login: "octocat"}},
		{app.FetchCalendar{Login: "bob", Range: r}, app.GotCalendar{Login: "bob", Range: r, Calendar: app.NewCalendar(r, nil)}},
		{app.FetchOrgMembers{Org: "acme"}, app.GotOrgMembers{Org: "acme", Members: []string{"alice"}}},
	}
	for _, tt := range tests {
		if got := p.performOne(tt.cmd)(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("performOne(%#v)() = %#v, want %#v", tt.cmd, got, tt.want)
		}
	}
}

func TestQuitCancelsRequests(t *testing.T) {
	t.Parallel()

	p := newTestProgram()
	if cmd := p.performOne(app.Quit{}); cmd == nil {
		t.Fatal("Quit should return tea.Quit")
	}
	if p.ctx.Err() == nil {
		t.Error("Quit should cancel in-flight requests")
	}
}

func TestSubscribeSchedulesSingleTick(t *testing.T) {
	t.Parallel()

	p := newTestProgram() // 起動直後は読み込み中なので Tick を購読する
	if p.subscribe() == nil {
		t.Fatal("should schedule a tick while loading")
	}
	if p.subscribe() != nil {
		t.Error("should not schedule another tick while one is pending")
	}

	// Tick が届いたら次を予約し直す。
	_, cmd := p.Update(tickMsg{})
	if cmd == nil || !p.ticking {
		t.Error("should reschedule a tick after receiving one while loading")
	}
}

func TestUpdateTranslatesKeys(t *testing.T) {
	t.Parallel()

	p := newTestProgram()
	p.Update(app.GotViewer{Login: "octocat"})
	p.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if got := app.View(p.model); !strings.Contains(got, "octocat") || !strings.Contains(got, "2026-09-23") {
		t.Errorf("left key should move to the previous day:\n%s", got)
	}
}
