package app

// Cmd は Update が外殻に依頼する副作用。
type Cmd interface{ isCmd() }

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

func (FetchViewer) isCmd()     {}
func (FetchCalendar) isCmd()   {}
func (FetchOrgMembers) isCmd() {}
func (Quit) isCmd()            {}
