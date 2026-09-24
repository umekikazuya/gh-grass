package app

import "time"

// tickInterval は読み込み中表示のアニメーション間隔。
const tickInterval = 100 * time.Millisecond

// Sub は Model の状態に応じて外殻に購読を依頼するイベント源（Elm の Sub に相当）。
type Sub interface{ isSub() }

// Every は Interval ごとに Ticked を届ける。
type Every struct{ Interval time.Duration }

func (Every) isSub() {}

// Subscriptions は現在の Model が必要とする購読を返す。読み込み中だけアニメーション用の Tick を購読する。
func Subscriptions(m Model) []Sub {
	if m.loading() {
		return []Sub{Every{Interval: tickInterval}}
	}
	return nil
}
