package app

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

var update = flag.Bool("update", false, "update golden files")

// assertGolden は View の出力（色の制御コードを除いたもの）を testdata/<name>.golden と比較する。
func assertGolden(t *testing.T, name string, m Model) {
	t.Helper()
	got := ansi.Strip(View(m))
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file (run `go test ./internal/app -update` to create): %v", err)
	}
	if got != string(want) {
		t.Errorf("View() mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

func TestView(t *testing.T) {
	t.Parallel()

	loading, _ := Init(Flags{Today: today})
	viewerFailed, _ := Update(loading, GotViewer{Err: ErrUnauthorized})

	fetchingOlder := started(t)
	fetchingOlder.selected = NewDate(2025, 10, 10)
	fetchingOlder, _ = press(t, fetchingOlder, "up")

	notFound, _ := press(t, started(t), "u", "g", "h", "o", "s", "t", "enter")
	notFound, _ = Update(notFound, GotCalendar{Login: "ghost", Range: fetchRangeEnding(today), Err: ErrNotFound})

	picker, _ := press(t, started(t), "o", "a", "c", "m", "e", "enter")
	picker, _ = Update(picker, GotOrgMembers{Org: "acme", Members: []string{"alice", "Bob", "carol", "dave"}})
	picker, _ = press(t, picker, "a", "down")

	tests := map[string]Model{
		"loading_account": loading,
		"viewer_failed":   viewerFailed,
		"today":           started(t),
		"previous_week":   must(press(t, started(t), "up", "left")),
		"fetching_older":  fetchingOlder,
		"not_found":       notFound,
		"user_input":      must(press(t, started(t), "u", "b", "o")),
		"member_picker":   picker,
		"help":            must(press(t, started(t), "?")),
	}
	for name, m := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertGolden(t, name, m)
		})
	}
}

func must(m Model, _ []Effect) Model { return m }
