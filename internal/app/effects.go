package app

// Effect は Update が外殻に依頼する副作用（Elm の Cmd に相当）。
// 関数ではなくデータで表すことで、Update の戻り値をテストでそのまま比較できる。
// 各 Effect の結果として返る Msg は固定されている。
type Effect interface{ isEffect() }

// FetchViewer は認証中のユーザーを取得する。結果は GotViewer。
type FetchViewer struct{}

// FetchCalendar は Login の Range 期間のコントリビューションを取得する。結果は GotCalendar。
type FetchCalendar struct {
	Login string
	Range DateRange
}

// FetchOrgMembers は Org のメンバー一覧を取得する。結果は GotOrgMembers。
type FetchOrgMembers struct{ Org string }

// Quit はプログラムを終了する。
type Quit struct{}

func (FetchViewer) isEffect()     {}
func (FetchCalendar) isEffect()   {}
func (FetchOrgMembers) isEffect() {}
func (Quit) isEffect()            {}
