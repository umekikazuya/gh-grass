package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF7DB")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
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
	list  list.Model
	input textinput.Model

	// State Data
	targetUser  string
	targetDate  time.Time
	resultCount int
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

	return MainModel{
		uc:    uc,
		state: stateModeSelect,
		list:  l,
		input: ti,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
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
		m.state = stateSelectMember
		return m, nil

	case contributionMsg:
		m.resultCount = msg.count
		m.targetUser = msg.user
		m.state = stateResult
		return m, nil
	}

	// --- State Machine ---
	switch m.state {

	case stateModeSelect:
		var lstCmd tea.Cmd
		m.list, lstCmd = m.list.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				switch i.title {
				case "Self":
					m.targetUser = "" // Will be resolved to current user later
					return m.switchToDateSelect()
				case "Specific User":
					m.state = stateInputUser
					m.input.Placeholder = "Username"
					m.input.SetValue("")
					return m, nil
				case "Organization":
					m.state = stateInputOrg
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
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			val := m.input.Value()
			if val == "" {
				return m, nil // Ignore empty input
			}

			if m.state == stateInputUser {
				m.targetUser = val
				return m.switchToDateSelect()
			} else if m.state == stateInputOrg {
				m.state = stateLoading
				return m, fetchOrgMembersCmd(m.uc, val)
			} else if m.state == stateInputDate {
				t, err := time.Parse("2006-01-02", val)
				if err != nil {
					m.err = fmt.Errorf("invalid date format (use YYYY-MM-DD): %v", err)
					m.state = stateError
					return m, nil
				}
				m.targetDate = t
				m.state = stateLoading
				return m, checkContributionCmd(m.uc, m.targetUser, m.targetDate)
			}
		}
		return m, tiCmd

	case stateSelectMember:
		var lstCmd tea.Cmd
		m.list, lstCmd = m.list.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.targetUser = i.title
				return m.switchToDateSelect()
			}
		}
		return m, lstCmd

	case stateDateSelect:
		var lstCmd tea.Cmd
		m.list, lstCmd = m.list.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				today := time.Now()
				switch i.title {
				case "Today":
					m.targetDate = today
					m.state = stateLoading
					return m, checkContributionCmd(m.uc, m.targetUser, m.targetDate)
				case "Yesterday":
					m.targetDate = today.AddDate(0, 0, -1)
					m.state = stateLoading
					return m, checkContributionCmd(m.uc, m.targetUser, m.targetDate)
				case "Other (Date)":
					m.state = stateInputDate
					m.input.Placeholder = "YYYY-MM-DD"
					m.input.SetValue("")
					return m, nil
				}
			}
		}
		return m, lstCmd

	case stateResult, stateError:
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "q" {
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
	m.state = stateDateSelect
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
		return docStyle.Render("Loading... please wait.")

	case stateResult:
		dateStr := m.targetDate.Format("2006-01-02")
		return docStyle.Render(fmt.Sprintf(
			"%s's contributions on %s:\n\n  %d  \n\n(press q to quit)",
			m.targetUser,
			dateStr,
			m.resultCount,
		))

	case stateError:
		return docStyle.Render(fmt.Sprintf("Error occurred:\n\n%v\n\n(press q to quit)", m.err))
	}
	return ""
}

// --- Messages & Commands ---

type errMsg error

type orgMembersMsg []domain.User

type contributionMsg struct {
	count int
	user  string
}

// fetchOrgMembersCmd は指定した組織名のメンバー一覧を非同期に取得する tea.Cmd を返す。
// 成功時は orgMembersMsg を、失敗時は errMsg を返す。
func fetchOrgMembersCmd(uc *usecase.GrassUsecase, orgName string) tea.Cmd {
	return func() tea.Msg {
		members, err := uc.ListOrganizationMembers(context.Background(), orgName)
		if err != nil {
			return errMsg(err)
		}
		return orgMembersMsg(members)
	}
}

// checkContributionCmd は、指定されたユーザーと日付の貢献数を取得するコマンドを返す。
// user が空文字の場合は実行時に現在のユーザーを解決して使用する。コマンド実行時は貢献数取得に成功すると contributionMsg を、エラー発生時は errMsg を返す。
func checkContributionCmd(uc *usecase.GrassUsecase, user string, date time.Time) tea.Cmd {
	return func() tea.Msg {
		targetUser := user
		// If "Self" was selected (empty user), resolve the actual username
		if targetUser == "" {
			self, err := uc.GetSelf(context.Background())
			if err != nil {
				return errMsg(err)
			}
			targetUser = self.Login
		}

		count, err := uc.GetContributionCount(context.Background(), targetUser, date)
		if err != nil {
			return errMsg(err)
		}
		return contributionMsg{count: count, user: targetUser}
	}
}
