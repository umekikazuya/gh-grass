package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/umekikazuya/gh-grass/internal/app"
)

// tickMsg は Every の購読で予約した Tick が来たことを表す。app には app.Ticked として渡す。
type tickMsg struct{}

// subscribe は app.Subscriptions に従って Tick を予約する。
// Bubble Tea には購読の仕組みがないため、必要な間だけ 1 回ずつ予約し直す。
func (p *program) subscribe() tea.Cmd {
	if p.ticking {
		return nil
	}
	for _, s := range app.Subscriptions(p.model) {
		if every, ok := s.(app.Every); ok {
			p.ticking = true
			return tea.Tick(every.Interval, func(time.Time) tea.Msg { return tickMsg{} })
		}
	}
	return nil
}
