package ui

import "github.com/charmbracelet/lipgloss"

// GitHub のコントリビューショングラフ相当の緑濃淡（強度 0〜4）。
// 0: 非コントリビュート、1〜4: 濃くなるほど多い。
var grassPalette = []lipgloss.Color{
	lipgloss.Color("#d1d5db"),
	lipgloss.Color("#aceebb"),
	lipgloss.Color("#4ac26b"),
	lipgloss.Color("#2da44e"),
	lipgloss.Color("#116329"),
}

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	// grassIntensity は強度（0〜4）ごとの描画スタイル。グリフを Foreground でパレット色に塗る。
	grassIntensity = buildGrassStyles()
)

func buildGrassStyles() []lipgloss.Style {
	styles := make([]lipgloss.Style, len(grassPalette))
	for i, c := range grassPalette {
		styles[i] = lipgloss.NewStyle().Foreground(c)
	}
	return styles
}

// intensityIndex は count から grassIntensity のインデックス（0〜4）を返す。
func intensityIndex(count int) int {
	switch {
	case count <= 0:
		return 0
	case count <= 2:
		return 1
	case count <= 5:
		return 2
	case count <= 9:
		return 3
	default:
		return 4
	}
}
