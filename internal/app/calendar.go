package app

import "maps"

// fetchSpanDays は 1 回の取得で読み込む日数。GitHub API は 1 年を超える期間を受け付けない。
const fetchSpanDays = 365

// Calendar は取得済みの日ごとのコントリビューション数。Range の外は未取得を表す。
type Calendar struct {
	Range  DateRange
	counts map[Date]int
}

// NewCalendar は期間 r と日ごとの件数から Calendar を作る。counts は複製して保持する。
func NewCalendar(r DateRange, counts map[Date]int) Calendar {
	return Calendar{Range: r, counts: maps.Clone(counts)}
}

// Count は d のコントリビューション数を返す。未取得または 0 件なら 0。
func (c Calendar) Count(d Date) int {
	return c.counts[d]
}

// Merge は c と o を合わせた Calendar を返す。
// 期間が隣接も重複もしていない場合は、未取得の隙間を取得済みと誤認しないよう o で置き換える。
func (c Calendar) Merge(o Calendar) Calendar {
	if o.Range.From.After(c.Range.To.AddDays(1)) || o.Range.To.Before(c.Range.From.AddDays(-1)) {
		return o
	}

	counts := maps.Clone(c.counts)
	if counts == nil {
		counts = map[Date]int{}
	}
	maps.Copy(counts, o.counts)

	r := c.Range
	if o.Range.From.Before(r.From) {
		r.From = o.Range.From
	}
	if o.Range.To.After(r.To) {
		r.To = o.Range.To
	}
	return Calendar{Range: r, counts: counts}
}

// Total は期間 r のコントリビューション数の合計を返す。
func (c Calendar) Total(r DateRange) int {
	total := 0
	for d := r.From; !d.After(r.To); d = d.AddDays(1) {
		total += c.counts[d]
	}
	return total
}

// Streak は end を終点として、コントリビューションが 1 件以上の日が何日続いているかを返す。
// end 自体が 0 件なら 0。取得済みの範囲より前には遡らない。
func (c Calendar) Streak(end Date) int {
	streak := 0
	for d := end; c.Range.Contains(d) && c.counts[d] > 0; d = d.AddDays(-1) {
		streak++
	}
	return streak
}

// fetchRangeEnding は to を終点とする 1 回分の取得期間を返す。
func fetchRangeEnding(to Date) DateRange {
	return DateRange{From: to.AddDays(-(fetchSpanDays - 1)), To: to}
}

// Intensity は件数を草の濃さ（0〜4）に変換する。0 は草なし、4 が最も濃い。
func Intensity(count int) int {
	switch {
	case count <= 0:
		return 0
	case count <= 2:
		return 1
	case count <= 5:
		return 2
	case count <= 9:
		return 3
	default:
		return 4
	}
}
