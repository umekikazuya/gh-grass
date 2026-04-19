package ui

import "github.com/charmbracelet/lipgloss"

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	// grassIntensity は強度（0〜4）ごとの描画文字。PR2 で色付きスタイルに差し替える前提。
	grassIntensity = []string{"·", "░", "▒", "▓", "█"}

	grassTargetStyle = lipgloss.NewStyle().Bold(true).Underline(true)
)

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
