package app

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// grassPalette は GitHub のコントリビューショングラフ相当の緑の濃淡（Intensity の 0〜4 に対応）。
var grassPalette = []color.Color{
	lipgloss.Color("#d1d5db"),
	lipgloss.Color("#aceebb"),
	lipgloss.Color("#4ac26b"),
	lipgloss.Color("#2da44e"),
	lipgloss.Color("#116329"),
}

var (
	docStyle    = lipgloss.NewStyle().Margin(1, 2)
	titleStyle  = lipgloss.NewStyle().Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#dc2626"))
	cursorStyle = lipgloss.NewStyle().Bold(true)
	grassStyles = buildGrassStyles()
)

func buildGrassStyles() []lipgloss.Style {
	styles := make([]lipgloss.Style, len(grassPalette))
	for i, c := range grassPalette {
		styles[i] = lipgloss.NewStyle().Foreground(c)
	}
	return styles
}

// keyHelp はキー操作の説明 1 行分。
type keyHelp struct {
	keys, desc string
}

// mainKeys は結果画面のキー操作の一覧。ヘルプとフッターの表示に使う。
var mainKeys = []keyHelp{
	{"←/→ h/l", "previous / next day"},
	{"↑/↓ k/j", "previous / next week"},
	{"t", "today"},
	{"u", "show another user"},
	{"o", "pick a member of an organization"},
	{"m", "show yourself"},
	{"r", "reload"},
	{"?", "toggle help"},
	{"q ctrl+c", "quit"},
}
