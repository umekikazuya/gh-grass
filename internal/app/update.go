package app

import (
	"maps"
	"strings"
	"unicode/utf8"
)

// 複数の画面で共通して使うキー。
const (
	keyEsc   = "esc"
	keyEnter = "enter"
)

// Update は Msg を受け取り、新しい Model と外殻に依頼する副作用を返す。副作用は実行しない。
func Update(m Model, msg Msg) (Model, []Effect) {
	switch msg := msg.(type) {
	case KeyPressed:
		return m.updateKey(msg.Key)
	case WindowResized:
		m.height = msg.Height
		return m, nil
	case Ticked:
		m.frame++
		return m, nil
	case GotViewer:
		return m.gotViewer(msg)
	case GotCalendar:
		return m.gotCalendar(msg)
	case GotOrgMembers:
		return m.gotOrgMembers(msg), nil
	}
	return m, nil
}

// --- キー操作 ---

func (m Model) updateKey(key string) (Model, []Effect) {
	if key == "ctrl+c" {
		return m, []Effect{Quit{}}
	}
	switch o := m.overlay.(type) {
	case userInput:
		return m.updateUserInput(o, key)
	case orgInput:
		return m.updateOrgInput(o, key)
	case memberPicker:
		return m.updateMemberPicker(o, key)
	case helpOverlay:
		if key == "?" || key == keyEsc || key == "q" {
			m.overlay = nil
		}
		return m, nil
	}
	return m.updateMain(key)
}

// updateMain は何も開いていないときのキー操作。
func (m Model) updateMain(key string) (Model, []Effect) {
	switch key {
	case "q":
		return m, []Effect{Quit{}}
	case "left", "h":
		return m.selectDate(m.selected.AddDays(-1))
	case "right", "l":
		return m.selectDate(m.selected.AddDays(1))
	case "up", "k":
		return m.selectDate(m.selected.AddDays(-7))
	case "down", "j":
		return m.selectDate(m.selected.AddDays(7))
	case "t":
		return m.selectDate(m.today)
	case "u":
		m.overlay = userInput{}
	case "o":
		m.overlay = orgInput{}
	case "?":
		m.overlay = helpOverlay{}
	case "m":
		if me, ok := m.viewer.Value(); ok {
			return m.showUser(me)
		}
	case "r":
		return m.reload()
	}
	return m, nil
}

func (m Model) updateUserInput(o userInput, key string) (Model, []Effect) {
	switch key {
	case keyEsc:
		m.overlay = nil
	case keyEnter:
		if login := strings.TrimSpace(o.text); login != "" {
			return m.showUser(login)
		}
	default:
		o.text = editText(o.text, key)
		m.overlay = o
	}
	return m, nil
}

func (m Model) updateOrgInput(o orgInput, key string) (Model, []Effect) {
	switch key {
	case keyEsc:
		m.overlay = nil
	case keyEnter:
		if org := strings.TrimSpace(o.text); org != "" {
			m.overlay = memberPicker{org: org, members: Pending[[]string]()}
			return m, []Effect{FetchOrgMembers{Org: org}}
		}
	default:
		o.text = editText(o.text, key)
		m.overlay = o
	}
	return m, nil
}

func (m Model) updateMemberPicker(p memberPicker, key string) (Model, []Effect) {
	candidates := p.candidates()
	switch key {
	case keyEsc:
		m.overlay = nil
		return m, nil
	case keyEnter:
		if p.cursor < len(candidates) {
			return m.showUser(candidates[p.cursor])
		}
		return m, nil
	case "up", "ctrl+p":
		p.cursor = max(p.cursor-1, 0)
	case "down", "ctrl+n":
		p.cursor = min(p.cursor+1, max(len(candidates)-1, 0))
	default:
		if filter := editText(p.filter, key); filter != p.filter {
			p.filter = filter
			p.cursor = 0
		}
	}
	m.overlay = p
	return m, nil
}

// editText は入力欄の文字列に key を反映する。1 文字のキーは追加、backspace は末尾の 1 文字を削除し、それ以外は無視する。
func editText(text, key string) string {
	switch {
	case key == "backspace":
		if text == "" {
			return text
		}
		_, size := utf8.DecodeLastRuneInString(text)
		return text[:len(text)-size]
	case utf8.RuneCountInString(key) == 1:
		return text + key
	default:
		return text
	}
}

// candidates は絞り込み条件（大文字小文字を区別しない部分一致）に合うメンバー。
func (p memberPicker) candidates() []string {
	members, _ := p.members.Value()
	filter := strings.ToLower(p.filter)
	var out []string
	for _, login := range members {
		if strings.Contains(strings.ToLower(login), filter) {
			out = append(out, login)
		}
	}
	return out
}

// --- 状態遷移 ---

// selectDate は選択中の日付を d にする。今日より先は選べない。
func (m Model) selectDate(d Date) (Model, []Effect) {
	if d.After(m.today) {
		d = m.today
	}
	m.selected = d
	return m.ensureCalendar()
}

// showUser は表示するユーザーを切り替え、開いている入力欄などを閉じる。
func (m Model) showUser(login string) (Model, []Effect) {
	m.login = login
	m.err = nil
	m.overlay = nil
	return m.ensureCalendar()
}

// reload は表示中のユーザーのデータを取り直す。自分の取得に失敗していればそちらをやり直す。
func (m Model) reload() (Model, []Effect) {
	if m.viewer.State() == Failure {
		m.viewer = Pending[string]()
		return m, []Effect{FetchViewer{}}
	}
	if m.login == "" {
		return m, nil
	}
	if _, busy := m.inflight[m.login]; busy {
		return m, nil
	}
	m.calendars = without(m.calendars, m.login)
	m.err = nil
	return m.ensureCalendar()
}

// ensureCalendar はグラフの表示期間が取得済みでなければ、足りない期間の取得を依頼する。
// 1 回の取得は 1 年分なので、日付を 1 日や 1 週間ずつ動かしている間は通信しない。
func (m Model) ensureCalendar() (Model, []Effect) {
	if m.login == "" || m.err != nil {
		return m, nil
	}
	if _, busy := m.inflight[m.login]; busy {
		return m, nil
	}

	w := m.window()
	cal, ok := m.calendars[m.login]
	var r DateRange
	switch {
	case !ok:
		r = fetchRangeEnding(w.To)
	case w.From.Before(cal.Range.From):
		r = fetchRangeEnding(cal.Range.From.AddDays(-1))
	case w.To.After(cal.Range.To):
		r = fetchRangeEnding(w.To)
	default:
		return m, nil
	}

	m.inflight = with(m.inflight, m.login, r)
	return m, []Effect{FetchCalendar{Login: m.login, Range: r}}
}

// --- 副作用の結果 ---

func (m Model) gotViewer(msg GotViewer) (Model, []Effect) {
	if msg.Err != nil {
		m.viewer = Failed[string](msg.Err)
		return m, nil
	}
	m.viewer = Loaded(msg.Login)
	if m.login != "" {
		return m, nil
	}
	return m.showUser(msg.Login)
}

// gotCalendar は取得結果を反映する。
// ユーザーを切り替えた後に届いた結果も、データとしては正しいのでキャッシュに加える。
// エラーは、表示中のユーザーの最新の依頼に対するものだけを表示する。
func (m Model) gotCalendar(msg GotCalendar) (Model, []Effect) {
	latest := m.inflight[msg.Login] == msg.Range
	if latest {
		m.inflight = without(m.inflight, msg.Login)
	}

	if msg.Err != nil {
		if latest && msg.Login == m.login {
			m.err = msg.Err
		}
		return m, nil
	}

	cal := msg.Calendar
	if prev, ok := m.calendars[msg.Login]; ok {
		cal = prev.Merge(cal)
	}
	m.calendars = with(m.calendars, msg.Login, cal)
	if msg.Login != m.login {
		return m, nil
	}
	// 取得中に日付が動いていれば、さらに必要な期間があるかもしれない。
	return m.ensureCalendar()
}

// gotOrgMembers はメンバー一覧を反映する。すでに閉じた、または別の Organization の結果は捨てる。
func (m Model) gotOrgMembers(msg GotOrgMembers) Model {
	p, ok := m.overlay.(memberPicker)
	if !ok || p.org != msg.Org {
		return m
	}
	if msg.Err != nil {
		p.members = Failed[[]string](msg.Err)
	} else {
		p.members = Loaded(msg.Members)
	}
	m.overlay = p
	return m
}

// --- map を不変に扱うための補助 ---

// with は m を複製し、key に v を設定したものを返す。
func with[V any](m map[string]V, key string, v V) map[string]V {
	out := maps.Clone(m)
	if out == nil {
		out = map[string]V{}
	}
	out[key] = v
	return out
}

// without は m を複製し、key を取り除いたものを返す。
func without[V any](m map[string]V, key string) map[string]V {
	out := maps.Clone(m)
	delete(out, key)
	return out
}
