package app

import (
	"errors"
	"fmt"
	"strings"
)

const (
	glyph         = "█"
	labelWidth    = "       "
	defaultRows   = 10
	reservedLines = 16 // メンバー一覧以外に使う行数の目安
)

var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// View は Model を画面の文字列に変換する。
func View(m Model) string {
	sections := []string{m.viewBody()}
	if o := m.viewOverlay(); o != "" {
		sections = append(sections, o)
	}
	sections = append(sections, dimStyle.Render(m.footer()))
	return docStyle.Render(strings.Join(sections, "\n\n"))
}

// viewBody は結果画面の本体。表示するユーザーが決まるまでは、自分の取得状況を出す。
func (m Model) viewBody() string {
	if m.login == "" {
		if err := m.viewer.Err(); err != nil {
			return errorStyle.Render("Error: " + errorText(err))
		}
		return m.spinner() + " Loading your account..."
	}

	lines := []string{
		titleStyle.Render(m.login) + dimStyle.Render("  "+m.selected.String()+" ("+m.selected.Weekday().String()[:3]+")"),
		"",
		m.viewGraph(),
	}
	if summary := m.viewSummary(); summary != "" {
		lines = append(lines, "", summary)
	}
	if status := m.viewStatus(); status != "" {
		lines = append(lines, "", status)
	}
	return strings.Join(lines, "\n")
}

// viewGraph は表示期間のカレンダーを、週を行・曜日を列にしたヒートマップにする。
func (m Model) viewGraph() string {
	cal, loaded := m.calendars[m.login]
	w := m.window()

	var b strings.Builder
	b.WriteString(labelWidth + "Sun  Mon  Tue  Wed  Thu  Fri  Sat")
	for week := w.From; !week.After(w.To); week = week.AddDays(7) {
		// 行は日曜始まり、ISO 週番号は月曜始まりなので、その行の月曜日の週番号を出す。
		fmt.Fprintf(&b, "\n  W%02d  ", week.AddDays(1).ISOWeek())
		for d := week; d.Before(week.AddDays(7)); d = d.AddDays(1) {
			b.WriteString(m.cell(cal, loaded, d))
		}
	}
	return b.String()
}

// cell はグラフの 1 日分（幅 5）を返す。
func (m Model) cell(cal Calendar, loaded bool, d Date) string {
	switch {
	case d.After(m.today):
		return " -   "
	case !loaded || !cal.Range.Contains(d):
		return dimStyle.Render(" ·   ")
	}
	g := grassStyles[Intensity(cal.Count(d))].Render(glyph)
	if d == m.selected {
		return "[" + g + "]  "
	}
	return " " + g + "   "
}

// viewSummary は選択中の日の件数と、表示期間の合計・連続日数。表示期間が取得済みでなければ空。
func (m Model) viewSummary() string {
	cal, ok := m.calendars[m.login]
	w := m.window()
	if !ok || !cal.Range.Covers(w) {
		return ""
	}
	return fmt.Sprintf("%s: %s\nTotal (%d weeks): %d  |  Streak: %s",
		m.selected, plural(cal.Count(m.selected), "contribution"),
		graphWeeks, cal.Total(w), plural(cal.Streak(m.selected), "day"))
}

// viewStatus は取得中の表示やエラー。
func (m Model) viewStatus() string {
	if m.err != nil {
		return errorStyle.Render("Error: "+errorText(m.err)) + dimStyle.Render("  (r to retry)")
	}
	if _, busy := m.inflight[m.login]; busy {
		return m.spinner() + " Fetching contributions..."
	}
	return ""
}

func (m Model) viewOverlay() string {
	switch o := m.overlay.(type) {
	case userInput:
		return "User: " + o.text + cursorStyle.Render("▏")
	case orgInput:
		return "Organization: " + o.text + cursorStyle.Render("▏")
	case memberPicker:
		return m.viewMemberPicker(o)
	case helpOverlay:
		return viewHelp()
	}
	return ""
}

func (m Model) viewMemberPicker(p memberPicker) string {
	header := fmt.Sprintf("Members of %s  filter: %s", p.org, p.filter) + cursorStyle.Render("▏")
	switch p.members.State() {
	case NotAsked, Loading:
		return header + "\n\n" + m.spinner() + " Fetching members..."
	case Failure:
		return header + "\n\n" + errorStyle.Render("Error: "+errorText(p.members.Err()))
	case Success:
	}

	candidates := p.candidates()
	if len(candidates) == 0 {
		return header + "\n\n" + dimStyle.Render("(no matching members)")
	}

	rows := m.listRows()
	start := max(0, p.cursor-rows+1)
	end := min(len(candidates), start+rows)
	lines := []string{header, ""}
	for i := start; i < end; i++ {
		if i == p.cursor {
			lines = append(lines, cursorStyle.Render("> "+candidates[i]))
		} else {
			lines = append(lines, "  "+candidates[i])
		}
	}
	lines = append(lines, dimStyle.Render(fmt.Sprintf("%d/%d", p.cursor+1, len(candidates))))
	return strings.Join(lines, "\n")
}

func viewHelp() string {
	lines := make([]string, 0, 2+len(mainKeys))
	lines = append(lines, titleStyle.Render("Keybindings"), "")
	for _, k := range mainKeys {
		lines = append(lines, fmt.Sprintf("  %-10s %s", k.keys, k.desc))
	}
	return strings.Join(lines, "\n")
}

// footer は画面下部のキー操作のヒント。開いているものに応じて出し分ける。
func (m Model) footer() string {
	switch m.overlay.(type) {
	case userInput, orgInput:
		return "enter show  esc cancel"
	case memberPicker:
		return "type to filter  ↑/↓ select  enter show  esc cancel"
	case helpOverlay:
		return "? esc close"
	}
	if m.login == "" {
		return "r retry  q quit"
	}
	return "←→ day  ↑↓ week  t today  u user  o org  m me  r reload  ? help  q quit"
}

// listRows はメンバー一覧に表示する行数。端末の高さが分からなければ既定値を使う。
func (m Model) listRows() int {
	if m.height <= 0 {
		return defaultRows
	}
	return min(max(m.height-reservedLines, 3), 20)
}

func (m Model) spinner() string {
	return spinnerFrames[m.frame%len(spinnerFrames)]
}

// errorText はエラーの種別に応じた利用者向けの文面を返す。
func errorText(err error) string {
	switch {
	case errors.Is(err, ErrNotFound):
		return "user or organization not found. Please check the spelling."
	case errors.Is(err, ErrUnauthorized):
		return "authentication failed. Run `gh auth login` and try again."
	case errors.Is(err, ErrTimeout):
		return "request timed out. Please check your network."
	default:
		return err.Error()
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
