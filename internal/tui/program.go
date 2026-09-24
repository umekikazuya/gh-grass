// Package tui は app（Elm アーキテクチャの本体）を Bubble Tea 上で動かす外殻。
// キー入力などを app の Msg に変換し、app が返す Cmd と Sub を実際に実行する。
package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/umekikazuya/gh-grass/internal/app"
)

// Client は Cmd を実行するために app が必要とする外部接続（ポート）。
type Client interface {
	Viewer(ctx context.Context) (string, error)
	Calendar(ctx context.Context, login string, r app.DateRange) (app.Calendar, error)
	OrgMembers(ctx context.Context, org string) ([]string, error)
}

// Run は app を起動し、終了するまでブロックする。
func Run(flags app.Flags, client Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, cmds := app.Init(flags)
	p := &program{ctx: ctx, cancel: cancel, client: client, model: m, initial: cmds}
	if _, err := tea.NewProgram(p).Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}
	return nil
}

// program は tea.Model の実装。状態は app.Model だけが持ち、ここは中継に徹する。
type program struct {
	ctx     context.Context // 終了時に実行中の取得をまとめて取り消すために保持する
	cancel  context.CancelFunc
	client  Client
	model   app.Model
	initial []app.Cmd
	ticking bool // Tick を予約済みか
}

func (p *program) Init() tea.Cmd {
	return tea.Batch(p.perform(p.initial), p.subscribe())
}

func (p *program) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var in app.Msg
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		in = app.KeyPressed{Key: msg.String()}
	case tea.WindowSizeMsg:
		in = app.WindowResized{Width: msg.Width, Height: msg.Height}
	case tickMsg:
		p.ticking = false
		in = app.Ticked{}
	case app.Msg:
		// 副作用の結果（GotViewer など）はそのまま渡す。
		in = msg
	default:
		return p, nil
	}

	var cmds []app.Cmd
	p.model, cmds = app.Update(p.model, in)
	return p, tea.Batch(p.perform(cmds), p.subscribe())
}

func (p *program) View() tea.View {
	return tea.NewView(app.View(p.model))
}
