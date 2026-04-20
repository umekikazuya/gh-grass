package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/umekikazuya/gh-grass/internal/domain"
	"github.com/umekikazuya/gh-grass/internal/usecase"
)

// Session State
type sessionState int

const (
	stateModeSelect sessionState = iota
	stateInputUser
	stateInputOrg
	stateSelectMember
	stateDateSelect
	stateInputDate
	stateLoading
	stateResult
	stateError
	stateHelp
)

// item implements list.Item interface
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type MainModel struct {
	uc    *usecase.GrassUsecase
	state sessionState
	err   error

	// UI Components
	list    list.Model
	input   textinput.Model
	spinner spinner.Model

	// State Data
	targetUser       string
	targetDate       time.Time
	resultCalendar   *domain.ContributionCalendar
	resultTargetHits int
	loadingLabel     string
	history          []sessionState
}

// NewInitialModel は指定された GrassUsecase を用いて、モード選択リストと入力フィールドを備えた初期の MainModel を生成します.
// 生成されるモデルはモード選択状態（stateModeSelect）で、リストは "Select Mode" タイトルと三つの選択肢（Self、Specific User、Organization）を持ち、ステータスバーとフィルタリングは無効、テキスト入力はフォーカス済みになります。
func NewInitialModel(uc *usecase.GrassUsecase) MainModel {
	// Initialize Mode Selection List
	items := []list.Item{
		item{title: "Self", desc: "Check your own contributions"},
		item{title: "Specific User", desc: "Check another user's contributions"},
		item{title: "Organization", desc: "Select a member from an organization"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Mode"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return MainModel{
		uc:      uc,
		state:   stateModeSelect,
		list:    l,
		input:   ti,
		spinner: sp,
	}
}

func (m MainModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// pushState は次の状態に遷移する前に現在の状態を履歴に積む。
func (m MainModel) pushState(next sessionState) MainModel {
	m.history = append(m.history, m.state)
	m.state = next
	return m
}

// popState は履歴から直近の状態を取り出して復元する。履歴が空なら何もしない。
// 復元対象のリスト/入力 UI を必要に応じて再初期化する。
func (m MainModel) popState() (MainModel, bool) {
	if len(m.history) == 0 {
		return m, false
	}
	prev := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	m.state = prev
	m.err = nil
	m.loadingLabel = ""
	m = m.restoreState(prev)
	return m, true
}

// restoreState は復元先の状態に合わせて list/input を再設定する。
func (m MainModel) restoreState(s sessionState) MainModel {
	switch s {
	case stateModeSelect:
		m.list.SetItems([]list.Item{
			item{title: "Self", desc: "Check your own contributions"},
			item{title: "Specific User", desc: "Check another user's contributions"},
			item{title: "Organization", desc: "Select a member from an organization"},
		})
		m.list.Title = "Select Mode"
		m.list.ResetSelected()
	case stateInputUser:
		m.input.Placeholder = "Username"
		m.input.SetValue("")
	case stateInputOrg:
		m.input.Placeholder = "Organization Name"
		m.input.SetValue("")
	case stateInputDate:
		m.input.Placeholder = "YYYY-MM-DD"
		m.input.SetValue("")
	case stateDateSelect:
		m.list.SetItems([]list.Item{
			item{title: "Today", desc: time.Now().Format("2006-01-02")},
			item{title: "Yesterday", desc: time.Now().AddDate(0, 0, -1).Format("2006-01-02")},
			item{title: "Other (Date)", desc: "Specify a custom date"},
		})
		m.list.Title = "Select Date"
		m.list.ResetSelected()
	}
	return m
}

// backAvailable は b キーによる戻る操作を現在の状態で許可するか返す。
// テキスト入力中は b を文字として扱うため無効、ロード中も無効。
func (m MainModel) backAvailable() bool {
	switch m.state {
	case stateInputUser, stateInputOrg, stateInputDate, stateLoading:
		return false
	}
	return len(m.history) > 0
}

// helpAvailable は ? キーでヘルプを開くのを現在の状態で許可するか返す。
func (m MainModel) helpAvailable() bool {
	switch m.state {
	case stateInputUser, stateInputOrg, stateInputDate, stateLoading, stateHelp:
		return false
	}
	return true
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case spinner.TickMsg:
		var spCmd tea.Cmd
		m.spinner, spCmd = m.spinner.Update(msg)
		return m, spCmd

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, keys.Help) && m.helpAvailable() {
			m = m.pushState(stateHelp)
			return m, nil
		}
		if key.Matches(msg, keys.Back) && m.backAvailable() {
			popped, ok := m.popState()
			if ok {
				return popped, nil
			}
		}

	// --- Async Result Handling ---
	case errMsg:
		m.err = msg
		m.state = stateError
		return m, nil

	case orgMembersMsg:
		items := make([]list.Item, len(msg))
		for i, u := range msg {
			items[i] = item{title: u.Login, desc: "Organization Member"}
		}
		m.list.SetItems(items)
		m.list.Title = "Select Member"
		m.list.ResetSelected()
		m.loadingLabel = ""
		m.state = stateSelectMember
		return m, nil

	case contributionMsg:
		m.resultCalendar = msg.calendar
		m.resultTargetHits = msg.targetCount
		m.targetUser = msg.user
		m.loadingLabel = ""
		m.state = stateResult
		return m, nil
	}

	// --- State Machine ---
	switch m.state {

	case stateModeSelect:
		var lstCmd tea.Cmd
		m.list, lstCmd = m.list.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && key.Matches(km, keys.Confirm) {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				switch i.title {
				case "Self":
					m.targetUser = "" // Will be resolved to current user later
					return m.switchToDateSelect()
				case "Specific User":
					m = m.pushState(stateInputUser)
					m.input.Placeholder = "Username"
					m.input.SetValue("")
					return m, nil
				case "Organization":
					m = m.pushState(stateInputOrg)
					m.input.Placeholder = "Organization Name"
					m.input.SetValue("")
					return m, nil
				}
			}
		}
		return m, lstCmd

	case stateInputUser, stateInputOrg, stateInputDate:
		var tiCmd tea.Cmd
		m.input, tiCmd = m.input.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && key.Matches(km, keys.Confirm) {
			val := m.input.Value()
			if val == "" {
				return m, nil // Ignore empty input
			}

			if m.state == stateInputUser {
				m.targetUser = val
				return m.switchToDateSelect()
			} else if m.state == stateInputOrg {
				m = m.pushState(stateLoading)
				m.loadingLabel = "Fetching organization members..."
				return m, fetchOrgMembersCmd(m.uc, val)
			} else if m.state == stateInputDate {
				t, err := time.Parse("2006-01-02", val)
				if err != nil {
					m.err = fmt.Errorf("invalid date format (use YYYY-MM-DD): %v", err)
					m.state = stateError
					return m, nil
				}
				m.targetDate = t
				m = m.pushState(stateLoading)
				m.loadingLabel = "Fetching contributions..."
				return m, checkContributionCmd(m.uc, m.targetUser, m.targetDate)
			}
		}
		return m, tiCmd

	case stateDateSelect:
		var lstCmd tea.Cmd
		m.list, lstCmd = m.list.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && key.Matches(km, keys.Confirm) {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				today := time.Now()
				switch i.title {
				case "Today":
					m.targetDate = today
					m = m.pushState(stateLoading)
					m.loadingLabel = "Fetching contributions..."
					return m, checkContributionCmd(m.uc, m.targetUser, m.targetDate)
				case "Yesterday":
					m.targetDate = today.AddDate(0, 0, -1)
					m = m.pushState(stateLoading)
					m.loadingLabel = "Fetching contributions..."
					return m, checkContributionCmd(m.uc, m.targetUser, m.targetDate)
				case "Other (Date)":
					m = m.pushState(stateInputDate)
					m.input.Placeholder = "YYYY-MM-DD"
					m.input.SetValue("")
					return m, nil
				}
			}
		}
		return m, lstCmd

	case stateSelectMember:
		var lstCmd tea.Cmd
		m.list, lstCmd = m.list.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && key.Matches(km, keys.Confirm) {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.targetUser = i.title
				return m.switchToDateSelect()
			}
		}
		return m, lstCmd

	case stateResult, stateError, stateHelp:
		if km, ok := msg.(tea.KeyMsg); ok && key.Matches(km, keys.DoneKey) {
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m MainModel) switchToDateSelect() (MainModel, tea.Cmd) {
	items := []list.Item{
		item{title: "Today", desc: time.Now().Format("2006-01-02")},
		item{title: "Yesterday", desc: time.Now().AddDate(0, 0, -1).Format("2006-01-02")},
		item{title: "Other (Date)", desc: "Specify a custom date"},
	}
	m.list.SetItems(items)
	m.list.Title = "Select Date"
	m.list.ResetSelected()
	m = m.pushState(stateDateSelect)
	return m, nil
}

func (m MainModel) View() string {
	switch m.state {
	case stateModeSelect, stateSelectMember, stateDateSelect:
		return docStyle.Render(m.list.View())

	case stateInputUser, stateInputOrg, stateInputDate:
		return docStyle.Render(fmt.Sprintf(
			"Enter value:\n\n%s\n\n(esc to quit)",
			m.input.View(),
		))

	case stateLoading:
		label := m.loadingLabel
		if label == "" {
			label = "Loading..."
		}
		return docStyle.Render(m.spinner.View() + " " + label)

	case stateResult:
		return docStyle.Render(renderGrassGraph(m.resultCalendar, m.targetUser, m.targetDate, m.resultTargetHits) + "\n" + footerHint(m))

	case stateError:
		return docStyle.Render(fmt.Sprintf("%s\n\n%s", formatError(m.err), footerHint(m)))

	case stateHelp:
		return docStyle.Render(renderHelp() + "\n\n" + footerHint(m))
	}
	return ""
}

// footerHint は画面下部に出すキーバインドヒント。backAvailable/helpAvailable に応じて項目を出し分ける。
func footerHint(m MainModel) string {
	parts := []string{keys.DoneKey.Help().Key + " quit"}
	if m.backAvailable() {
		parts = append(parts, keys.Back.Help().Key+" back")
	}
	if m.helpAvailable() {
		parts = append(parts, keys.Help.Help().Key+" help")
	}
	return "(" + strings.Join(parts, " | ") + ")"
}

// renderHelp はヘルプ画面の本文（キーバインド一覧）を返す。
func renderHelp() string {
	var b strings.Builder
	b.WriteString("Keybindings\n\n")
	for _, binding := range helpEntries() {
		h := binding.Help()
		fmt.Fprintf(&b, "  %-12s %s\n", h.Key, h.Desc)
	}
	return b.String()
}

// --- Messages & Commands ---

type errMsg error

type orgMembersMsg []domain.User

type contributionMsg struct {
	calendar    *domain.ContributionCalendar
	targetCount int
	user        string
}

// fetchOrgMembersCmd は指定した組織名のメンバー一覧を非同期に取得する tea.Cmd を返す。
// 成功時は orgMembersMsg を、失敗時は errMsg を返す。
func fetchOrgMembersCmd(uc *usecase.GrassUsecase, orgName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		members, err := uc.ListOrganizationMembers(ctx, orgName)
		if err != nil {
			return errMsg(err)
		}
		return orgMembersMsg(members)
	}
}

// grassWeeks は stateResult で描画するコントリビューションカレンダーの週数。
const grassWeeks = 4

// checkContributionCmd は、指定されたユーザーと日付を終点とする直近数週間のカレンダーを取得するコマンドを返す。
// user が空文字の場合は実行時に現在のユーザーを解決して使用する。成功時は contributionMsg を、エラー発生時は errMsg を返す。
func checkContributionCmd(uc *usecase.GrassUsecase, user string, date time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		targetUser := user
		// If "Self" was selected (empty user), resolve the actual username
		if targetUser == "" {
			self, err := uc.GetSelf(ctx)
			if err != nil {
				return errMsg(err)
			}
			targetUser = self.Login
		}

		cal, err := uc.GetContributionCalendar(ctx, targetUser, date, grassWeeks)
		if err != nil {
			return errMsg(err)
		}

		targetCount := 0
		targetStr := date.Format("2006-01-02")
		for _, week := range cal.Weeks {
			for _, day := range week {
				if day.Date.Format("2006-01-02") == targetStr {
					targetCount = day.Count
				}
			}
		}

		return contributionMsg{calendar: cal, targetCount: targetCount, user: targetUser}
	}
}

// formatError はエラー内容を種別判定してユーザー向けの文面に整形する。
// 判定できないものは原文を返す。
func formatError(err error) string {
	if err == nil {
		return "Error: unknown error"
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(msg, "Could not resolve to") ||
		strings.Contains(lower, "not found") ||
		strings.Contains(lower, "not resolve"):
		return "Error: user or organization not found. Please check the spelling."
	case strings.Contains(lower, "deadline exceeded") ||
		strings.Contains(lower, "timeout"):
		return "Error: request timed out. Please check your network and try again."
	case strings.Contains(lower, "invalid date format"):
		return "Error: invalid date format. Use YYYY-MM-DD."
	case strings.Contains(lower, "unauthorized") ||
		strings.Contains(lower, "bad credentials"):
		return "Error: authentication failed. Run `gh auth login` and try again."
	default:
		return "Error: " + msg
	}
}

// renderGrassGraph はカレンダーを7列のヒートマップとして整形する。
// target と同じ日付のセルは強調表示され、target より未来の日付は "-" になる。
func renderGrassGraph(cal *domain.ContributionCalendar, user string, target time.Time, targetCount int) string {
	if cal == nil || len(cal.Weeks) == 0 {
		return fmt.Sprintf("%s's contributions on %s:\n\n  (no data)",
			user, target.Format("2006-01-02"))
	}

	weeks := cal.Weeks
	if len(weeks) > grassWeeks {
		weeks = weeks[len(weeks)-grassWeeks:]
	}

	const glyph = "█"

	var b strings.Builder

	b.WriteString("        Sun  Mon  Tue  Wed  Thu  Fri  Sat\n")

	targetStr := target.Format("2006-01-02")
	emptyCell := "     "
	for _, week := range weeks {
		if len(week) == 0 {
			continue
		}
		_, wn := week[0].Date.ISOWeek()
		fmt.Fprintf(&b, "  W%02d  ", wn)

		leadingEmpty := int(week[0].Date.Weekday())
		for i := 0; i < leadingEmpty; i++ {
			b.WriteString(emptyCell)
		}

		for _, day := range week {
			dayStr := day.Date.Format("2006-01-02")
			switch {
			case dayStr > targetStr:
				b.WriteString("  -  ")
			case dayStr == targetStr:
				style := grassIntensity[intensityIndex(day.Count)].Bold(true).Underline(true)
				b.WriteString("[" + style.Render(glyph) + "]  ")
			default:
				b.WriteString(" " + grassIntensity[intensityIndex(day.Count)].Render(glyph) + "   ")
			}
		}

		trailingEmpty := max(0, 7-leadingEmpty-len(week))
		for i := 0; i < trailingEmpty; i++ {
			b.WriteString(emptyCell)
		}
		b.WriteString("\n")
	}

	total, streak := summarizeCalendar(weeks, target)

	fmt.Fprintf(&b, "\n  %s's contributions on %s:  %d\n", user, target.Format("2006-01-02"), targetCount)
	fmt.Fprintf(&b, "  Total (%d weeks):  %d  |  Current streak:  %d day(s)", len(weeks), total, streak)
	return b.String()
}

// summarizeCalendar は表示範囲の合計コントリビューション数と、target を終点とする連続コントリビューション日数を返す。
// target 自体のカウントが 0 の場合、streak は 0。
func summarizeCalendar(weeks [][]domain.ContributionDay, target time.Time) (total, streak int) {
	counts := map[string]int{}
	for _, week := range weeks {
		for _, day := range week {
			total += day.Count
			counts[day.Date.Format("2006-01-02")] = day.Count
		}
	}

	for cursor := target; ; cursor = cursor.AddDate(0, 0, -1) {
		c, ok := counts[cursor.Format("2006-01-02")]
		if !ok || c <= 0 {
			break
		}
		streak++
	}
	return total, streak
}
