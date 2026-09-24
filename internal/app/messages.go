package app

import "errors"

// Msg は Update に渡される「起きたこと」。外殻（tui）がキー入力や副作用の結果をこの型に変換して渡す。
type Msg interface{ isMsg() }

// KeyPressed はキーが押されたこと。Key は "left"、"ctrl+c"、"a" のようなキーの文字列表現。
type KeyPressed struct{ Key string }

// WindowResized は端末のサイズが変わったこと。
type WindowResized struct{ Width, Height int }

// Ticked は Every の購読により一定間隔で届く。
type Ticked struct{}

// GotViewer は FetchViewer の結果。
type GotViewer struct {
	Login string
	Err   error
}

// GotCalendar は FetchCalendar の結果。
type GotCalendar struct {
	Login    string
	Range    DateRange
	Calendar Calendar
	Err      error
}

// GotOrgMembers は FetchOrgMembers の結果。
type GotOrgMembers struct {
	Org     string
	Members []string
	Err     error
}

func (KeyPressed) isMsg()    {}
func (WindowResized) isMsg() {}
func (Ticked) isMsg()        {}
func (GotViewer) isMsg()     {}
func (GotCalendar) isMsg()   {}
func (GotOrgMembers) isMsg() {}

// 副作用の結果として返すエラーの種別。外殻はこれらで包んで返し、View は種別に応じた文面を出す。
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrTimeout      = errors.New("timeout")
)
