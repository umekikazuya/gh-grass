package app

// RemoteState は非同期に取得するデータの状態。
type RemoteState int

const (
	NotAsked RemoteState = iota
	Loading
	Failure
	Success
)

// Remote は非同期に取得するデータを表す（Elm の RemoteData パターン）。
// 「読み込み中なのに値がある」「失敗なのに値がある」といった矛盾した状態を作れないよう、
// フィールドは非公開にしてコンストラクタ経由でのみ生成する。
type Remote[T any] struct {
	state RemoteState
	value T
	err   error
}

// Pending は読み込み中の Remote を返す。
func Pending[T any]() Remote[T] {
	return Remote[T]{state: Loading}
}

// Loaded は取得に成功した Remote を返す。
func Loaded[T any](v T) Remote[T] {
	return Remote[T]{state: Success, value: v}
}

// Failed は取得に失敗した Remote を返す。
func Failed[T any](err error) Remote[T] {
	return Remote[T]{state: Failure, err: err}
}

// State は現在の状態を返す。
func (r Remote[T]) State() RemoteState {
	return r.state
}

// Value は取得に成功していれば値と true を返す。
func (r Remote[T]) Value() (T, bool) {
	return r.value, r.state == Success
}

// Err は取得に失敗していればそのエラーを返す。
func (r Remote[T]) Err() error {
	return r.err
}
