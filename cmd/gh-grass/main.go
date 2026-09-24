package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/umekikazuya/gh-grass/internal/app"
	"github.com/umekikazuya/gh-grass/internal/github"
	"github.com/umekikazuya/gh-grass/internal/tui"
)

// version はリリース時に ldflags で埋め込まれる。
var version = "dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// newRootCmd is to build command.
func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "gh-grass",
		Short:        "Check GitHub contributions from the terminal",
		Long:         "gh-grass is a GitHub CLI extension that shows the contribution graph of yourself, other users, or organization members in an interactive TUI.",
		Version:      version,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(*cobra.Command, []string) error {
			client, err := github.New()
			if err != nil {
				return err //nolint:wrapcheck
			}
			return tui.Run(app.Flags{Today: app.DateOf(time.Now())}, client)
		},
	}
}
