package app

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

// today は 2026-09-24（木）。テストはすべてこの日を「今日」として動かす。
var today = NewDate(2026, time.September, 24)

// sampleCounts は r の期間に決まった規則で件数を割り当てたもの。
func sampleCounts(r DateRange) map[Date]int {
	counts := map[Date]int{}
	for d := r.From; !d.After(r.To); d = d.AddDays(1) {
		counts[d] = (d.day * 7) % 12
	}
	return counts
}

func gotCalendar(login string, r DateRange) GotCalendar {
	return GotCalendar{Login: login, Range: r, Calendar: NewCalendar(r, sampleCounts(r))}
}

// press は keys を順に押した結果と、最後のキーで返った副作用を返す。
func press(t *testing.T, m Model, keys ...string) (Model, []Effect) {
	t.Helper()
	var effs []Effect
	for _, k := range keys {
		m, effs = Update(m, KeyPressed{Key: k})
	}
	return m, effs
}

func assertEffects(t *testing.T, got []Effect, want ...Effect) {
	t.Helper()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("effects = %#v, want %#v", got, want)
	}
}

// started は起動して自分（octocat）の 1 年分を取得し終えた状態を返す。
func started(t *testing.T) Model {
	t.Helper()
	m, effs := Init(Flags{Today: today})
	assertEffects(t, effs, FetchViewer{})

	m, effs = Update(m, GotViewer{Login: "octocat"})
	year := fetchRangeEnding(today)
	assertEffects(t, effs, FetchCalendar{Login: "octocat", Range: year})

	m, effs = Update(m, gotCalendar("octocat", year))
	assertEffects(t, effs)
	return m
}

func TestStartup(t *testing.T) {
	t.Parallel()

	m, _ := Init(Flags{Today: today})
	if len(Subscriptions(m)) == 0 {
		t.Error("should subscribe to ticks while loading")
	}

	m = started(t)
	if m.login != "octocat" || m.selected != today {
		t.Errorf("login = %q, selected = %s", m.login, m.selected)
	}
	if fetchRangeEnding(today).From != NewDate(2025, time.September, 25) {
		t.Errorf("one fetch should span 365 days, got from %s", fetchRangeEnding(today).From)
	}
	if subs := Subscriptions(m); len(subs) != 0 {
		t.Errorf("should not subscribe after loading, got %v", subs)
	}
}

func TestDateNavigation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		keys []string
		want Date
	}{
		{"← で前日", []string{"left"}, NewDate(2026, time.September, 23)},
		{"h で前日", []string{"h"}, NewDate(2026, time.September, 23)},
		{"↑ で前の週", []string{"up"}, NewDate(2026, time.September, 17)},
		{"k で前の週", []string{"k"}, NewDate(2026, time.September, 17)},
		{"↓ で次の週", []string{"up", "up", "down"}, NewDate(2026, time.September, 17)},
		{"→ は今日より先に進めない", []string{"right", "l"}, today},
		{"↓ は今日で止まる", []string{"left", "j"}, today},
		{"t で今日に戻る", []string{"up", "up", "left", "t"}, today},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, effs := press(t, started(t), tt.keys...)
			if m.selected != tt.want {
				t.Errorf("selected = %s, want %s", m.selected, tt.want)
			}
			// 1 年分を取得済みなので、この範囲の移動では通信しない。
			assertEffects(t, effs)
		})
	}
}

func TestFetchOlderYear(t *testing.T) {
	t.Parallel()

	m := started(t)
	m.selected = NewDate(2025, time.October, 10)

	// 表示期間（2025-09-07〜）が取得済みの範囲（2025-09-25〜）からはみ出すので、その前の 1 年分を取りに行く。
	m, effs := press(t, m, "up")
	older := fetchRangeEnding(NewDate(2025, time.September, 24))
	assertEffects(t, effs, FetchCalendar{Login: "octocat", Range: older})

	// 取得中は重ねて依頼しない。
	m, effs = press(t, m, "up")
	assertEffects(t, effs)

	m, effs = Update(m, gotCalendar("octocat", older))
	assertEffects(t, effs)
	if got := m.calendars["octocat"].Range; got != (DateRange{From: older.From, To: today}) {
		t.Errorf("merged range = %v", got)
	}
}

func TestShowUser(t *testing.T) {
	t.Parallel()

	m, effs := press(t, started(t), "u", "b", "o", "x", "backspace", "b")
	assertEffects(t, effs)
	if o, ok := m.overlay.(userInput); !ok || o.text != "bob" {
		t.Fatalf("overlay = %#v", m.overlay)
	}

	m, effs = press(t, m, "enter")
	bobYear := fetchRangeEnding(today)
	assertEffects(t, effs, FetchCalendar{Login: "bob", Range: bobYear})
	if m.login != "bob" || m.overlay != nil {
		t.Errorf("login = %q, overlay = %#v", m.login, m.overlay)
	}

	// m で自分に戻る。自分のデータは取得済みなので通信しない。
	m, effs = press(t, m, "m")
	assertEffects(t, effs)
	if m.login != "octocat" {
		t.Errorf("login = %q, want octocat", m.login)
	}

	// 切り替えた後に届いた bob の結果もキャッシュされ、次に bob を開いたときは通信しない。
	m, _ = Update(m, gotCalendar("bob", bobYear))
	_, effs = press(t, m, "u", "b", "o", "b", "enter")
	assertEffects(t, effs)
}

func TestUserInputKeys(t *testing.T) {
	t.Parallel()

	t.Run("入力中の q や h は文字として扱う", func(t *testing.T) {
		t.Parallel()
		m, effs := press(t, started(t), "u", "q", "h")
		assertEffects(t, effs)
		if o, ok := m.overlay.(userInput); !ok || o.text != "qh" {
			t.Errorf("overlay = %#v", m.overlay)
		}
	})
	t.Run("空のまま enter しても何もしない", func(t *testing.T) {
		t.Parallel()
		m, effs := press(t, started(t), "u", "enter")
		assertEffects(t, effs)
		if _, ok := m.overlay.(userInput); !ok {
			t.Errorf("overlay = %#v", m.overlay)
		}
	})
	t.Run("esc で閉じる", func(t *testing.T) {
		t.Parallel()
		m, _ := press(t, started(t), "u", "a", "esc")
		if m.overlay != nil || m.login != "octocat" {
			t.Errorf("overlay = %#v, login = %q", m.overlay, m.login)
		}
	})
	t.Run("ctrl+c は入力中でも終了する", func(t *testing.T) {
		t.Parallel()
		_, effs := press(t, started(t), "u", "ctrl+c")
		assertEffects(t, effs, Quit{})
	})
}

func TestMemberPicker(t *testing.T) {
	t.Parallel()

	m, effs := press(t, started(t), "o", "a", "c", "m", "e", "enter")
	assertEffects(t, effs, FetchOrgMembers{Org: "acme"})
	if len(Subscriptions(m)) == 0 {
		t.Error("should subscribe to ticks while fetching members")
	}

	// 別の Organization の結果は捨てる。
	m, _ = Update(m, GotOrgMembers{Org: "other", Members: []string{"mallory"}})
	if p := m.overlay.(memberPicker); p.members.State() != Loading {
		t.Fatalf("members state = %v", p.members.State())
	}

	m, _ = Update(m, GotOrgMembers{Org: "acme", Members: []string{"alice", "Bob", "carol"}})
	m, _ = press(t, m, "O") // 大文字小文字を区別せず "o" を含む: Bob, carol
	if got := m.overlay.(memberPicker).candidates(); !reflect.DeepEqual(got, []string{"Bob", "carol"}) {
		t.Fatalf("candidates = %v", got)
	}

	m, _ = press(t, m, "down", "down") // 末尾で止まる
	m, effs = press(t, m, "enter")
	assertEffects(t, effs, FetchCalendar{Login: "carol", Range: fetchRangeEnding(today)})
	if m.login != "carol" || m.overlay != nil {
		t.Errorf("login = %q, overlay = %#v", m.login, m.overlay)
	}
}

func TestCalendarError(t *testing.T) {
	t.Parallel()

	m, _ := press(t, started(t), "u", "g", "h", "o", "s", "t", "enter")
	r := fetchRangeEnding(today)
	m, _ = Update(m, GotCalendar{Login: "ghost", Range: r, Err: ErrNotFound})
	if !errors.Is(m.err, ErrNotFound) {
		t.Fatalf("err = %v", m.err)
	}

	// エラーの間は、日付を動かしても取り直さない。
	m, effs := press(t, m, "left")
	assertEffects(t, effs)

	// r でやり直す。
	_, effs = press(t, m, "r")
	assertEffects(t, effs, FetchCalendar{Login: "ghost", Range: fetchRangeEnding(today)})

	// 表示していないユーザーのエラーは出さない。
	m, _ = press(t, started(t), "u", "b", "o", "b", "enter", "m")
	m, _ = Update(m, GotCalendar{Login: "bob", Range: r, Err: ErrTimeout})
	if m.err != nil {
		t.Errorf("err = %v, want nil", m.err)
	}
}

func TestViewerError(t *testing.T) {
	t.Parallel()

	m, _ := Init(Flags{Today: today})
	m, _ = Update(m, GotViewer{Err: ErrUnauthorized})
	if m.viewer.State() != Failure {
		t.Fatalf("viewer state = %v", m.viewer.State())
	}
	_, effs := press(t, m, "r")
	assertEffects(t, effs, FetchViewer{})
}

func TestQuitAndHelp(t *testing.T) {
	t.Parallel()

	_, effs := press(t, started(t), "q")
	assertEffects(t, effs, Quit{})

	m, _ := press(t, started(t), "?")
	if _, ok := m.overlay.(helpOverlay); !ok {
		t.Fatalf("overlay = %#v", m.overlay)
	}
	// ヘルプを開いている間の q はヘルプを閉じる。
	m, effs = press(t, m, "q")
	assertEffects(t, effs)
	if m.overlay != nil {
		t.Errorf("overlay = %#v", m.overlay)
	}
}

func TestUpdateDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	before := started(t)
	_, _ = Update(before, gotCalendar("alice", fetchRangeEnding(today)))
	_, _ = press(t, before, "r")

	if _, ok := before.calendars["alice"]; ok {
		t.Error("Update mutated calendars of the input model")
	}
	if _, ok := before.calendars["octocat"]; !ok {
		t.Error("reload mutated calendars of the input model")
	}
	if len(before.inflight) != 0 {
		t.Error("Update mutated inflight of the input model")
	}
}
