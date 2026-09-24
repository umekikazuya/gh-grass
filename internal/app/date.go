package app

import (
	"fmt"
	"time"
)

// Date はタイムゾーンを持たない暦日。
// GitHub のコントリビューションは日単位で集計されるため、時刻やタイムゾーンを持ち込まない。
type Date struct {
	year  int
	month time.Month
	day   int
}

// NewDate は年月日から Date を作る。範囲外の値は time.Date と同様に正規化される。
func NewDate(year int, month time.Month, day int) Date {
	return DateOf(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

// DateOf は t が属する（t 自身のタイムゾーンでの）暦日を返す。
func DateOf(t time.Time) Date {
	y, m, d := t.Date()
	return Date{year: y, month: m, day: d}
}

// ParseDate は YYYY-MM-DD 形式の文字列を Date に変換する。
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, fmt.Errorf("parse date %q: %w", s, err)
	}
	return DateOf(t), nil
}

// Time は d の UTC 0時を返す。
func (d Date) Time() time.Time {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC)
}

// AddDays は n 日後（負なら前）の日付を返す。
func (d Date) AddDays(n int) Date {
	return DateOf(d.Time().AddDate(0, 0, n))
}

// Before は d が o より前の日付かを返す。
func (d Date) Before(o Date) bool {
	return d.Time().Before(o.Time())
}

// After は d が o より後の日付かを返す。
func (d Date) After(o Date) bool {
	return d.Time().After(o.Time())
}

// Weekday は曜日を返す。
func (d Date) Weekday() time.Weekday {
	return d.Time().Weekday()
}

// ISOWeek は ISO 8601 の週番号を返す。
func (d Date) ISOWeek() int {
	_, w := d.Time().ISOWeek()
	return w
}

// StartOfWeek は d を含む週の日曜日を返す（GitHub のグラフに合わせて週は日曜始まり）。
func (d Date) StartOfWeek() Date {
	return d.AddDays(-int(d.Weekday()))
}

// String は YYYY-MM-DD 形式の文字列を返す。
func (d Date) String() string {
	return d.Time().Format(time.DateOnly)
}

// DateRange は From から To まで（両端を含む）の期間。
type DateRange struct {
	From, To Date
}

// Contains は d が期間内かを返す。
func (r DateRange) Contains(d Date) bool {
	return !d.Before(r.From) && !d.After(r.To)
}

// Covers は o が期間内に収まっているかを返す。
func (r DateRange) Covers(o DateRange) bool {
	return r.Contains(o.From) && r.Contains(o.To)
}
