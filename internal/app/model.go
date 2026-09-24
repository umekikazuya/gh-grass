package app

// graphWeeks はグラフに表示する週数。
const graphWeeks = 4

// Flags は起動時に外から渡す値（Elm の flags に相当）。
// app の中では現在時刻を取得しないので、今日の日付もここで受け取る。
type Flags struct {
	Today Date
}

// Model はアプリケーションの状態。Update はこの値を書き換えず、新しい Model を返す。
// map を更新するときは必ず複製してから書き換える。
type Model struct {
	today    Date
	viewer   Remote[string] // 認証中のユーザー
	login    string         // 表示中のユーザー。viewer の取得前は空
	selected Date           // 選択中の日付

	calendars map[string]Calendar  // 取得済みのカレンダー（ユーザーごと）
	inflight  map[string]DateRange // 取得中の期間（ユーザーごと）
	err       error                // 表示中のユーザーについて直近で失敗した取得のエラー

	overlay overlay // 画面の上に開いている入力欄など。nil なら何も開いていない
	frame   int     // 読み込み中表示のアニメーションのフレーム
	height  int
}

// overlay は結果画面の上に開く小さな画面。
type overlay interface{ isOverlay() }

// userInput はユーザー名の入力欄。
type userInput struct{ text string }

// orgInput は Organization 名の入力欄。
type orgInput struct{ text string }

// memberPicker は Organization のメンバー選択。filter で絞り込み、cursor は絞り込み後の位置。
type memberPicker struct {
	org     string
	members Remote[[]string]
	filter  string
	cursor  int
}

// helpOverlay はキー操作の一覧。
type helpOverlay struct{}

func (userInput) isOverlay()    {}
func (orgInput) isOverlay()     {}
func (memberPicker) isOverlay() {}
func (helpOverlay) isOverlay()  {}

// Init は初期状態と最初の副作用を返す。起動直後は自分の今日のグラフを表示するため、まず認証中のユーザーを取得する。
func Init(flags Flags) (Model, []Cmd) {
	m := Model{
		today:    flags.Today,
		viewer:   Pending[string](),
		selected: flags.Today,
	}
	return m, []Cmd{FetchViewer{}}
}

// window はグラフに表示する期間。選択中の日を含む週までの graphWeeks 週間で、今日より先は含めない。
func (m Model) window() DateRange {
	start := m.selected.StartOfWeek()
	to := start.AddDays(6)
	if to.After(m.today) {
		to = m.today
	}
	return DateRange{From: start.AddDays(-7 * (graphWeeks - 1)), To: to}
}

// loading は何らかの取得を待っているかを返す。
func (m Model) loading() bool {
	if m.viewer.State() == Loading || len(m.inflight) > 0 {
		return true
	}
	p, ok := m.overlay.(memberPicker)
	return ok && p.members.State() == Loading
}
