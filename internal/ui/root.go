package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/umekikazuya/gh-grass/internal/infrastructure"
	"github.com/umekikazuya/gh-grass/internal/usecase"
)

var rootCmd = &cobra.Command{
	Use:   "gh-grass",
	Short: "Check GitHub contributions from the terminal",
	Long:  `gh-grass is a CLI extension for GitHub that allows you to check contribution counts for yourself, other users, or organization members using an interactive TUI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Get Token
		token, err := infrastructure.GetGHToken()
		if err != nil {
			return err
		}

		// 2. Setup Dependencies
		repo := infrastructure.NewGitHubRepository(token)
		uc := usecase.NewGrassUsecase(repo)

		// 3. Start TUI
		p := tea.NewProgram(NewInitialModel(uc))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run tui: %w", err)
		}
		return nil
	},
}

// Execute はルートの Cobra コマンドを実行します。
// 実行中にエラーが発生した場合はエラーを標準出力に表示してプロセスをステータス 1 で終了します.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// TODO: Bubble Teaがシグナルを処理するため、現時点ではcontext.Background()を使用。
// 将来的にコンテキストキャンセレーションが必要な場合は再検討。
