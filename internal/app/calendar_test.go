package app

import (
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	t.Parallel()

	d := NewDate(2026, time.March, 1)
	if got := d.AddDays(-1); got != NewDate(2026, time.February, 28) {
		t.Errorf("AddDays(-1) = %s, want 2026-02-28", got)
	}
	if got := NewDate(2026, time.September, 24).StartOfWeek(); got != NewDate(2026, time.September, 20) {
		t.Errorf("StartOfWeek = %s, want 2026-09-20 (Sun)", got)
	}
	if got, err := ParseDate("2026-09-24"); err != nil || got != NewDate(2026, time.September, 24) {
		t.Errorf("ParseDate = %s, %v", got, err)
	}
	if _, err := ParseDate("2026/09/24"); err == nil {
		t.Error("ParseDate should fail for invalid format")
	}
	// タイムゾーンをまたいでも、そのタイムゾーンでの暦日になる。
	jst := time.FixedZone("JST", 9*60*60)
	if got := DateOf(time.Date(2026, time.February, 16, 1, 56, 0, 0, jst)); got != NewDate(2026, time.February, 16) {
		t.Errorf("DateOf(JST) = %s, want 2026-02-16", got)
	}
}

func TestIntensity(t *testing.T) {
	t.Parallel()

	for count, want := range map[int]int{-1: 0, 0: 0, 1: 1, 2: 1, 3: 2, 5: 2, 6: 3, 9: 3, 10: 4, 100: 4} {
		if got := Intensity(count); got != want {
			t.Errorf("Intensity(%d) = %d, want %d", count, got, want)
		}
	}
}

func TestCalendarTotalAndStreak(t *testing.T) {
	t.Parallel()

	d := func(day int) Date { return NewDate(2026, time.September, day) }
	cal := NewCalendar(DateRange{From: d(1), To: d(10)}, map[Date]int{
		d(1): 5, d(3): 1, d(4): 2, d(5): 3,
	})

	if got := cal.Total(DateRange{From: d(1), To: d(10)}); got != 11 {
		t.Errorf("Total = %d, want 11", got)
	}
	tests := []struct {
		end  Date
		want int
	}{
		{d(5), 3}, // 3〜5 日が連続
		{d(6), 0}, // 当日が 0 件なら 0
		{d(1), 1}, // 取得範囲より前には遡らない
	}
	for _, tt := range tests {
		if got := cal.Streak(tt.end); got != tt.want {
			t.Errorf("Streak(%s) = %d, want %d", tt.end, got, tt.want)
		}
	}
}

func TestCalendarMerge(t *testing.T) {
	t.Parallel()

	d := func(day int) Date { return NewDate(2026, time.September, day) }
	newer := NewCalendar(DateRange{From: d(11), To: d(20)}, map[Date]int{d(15): 1})

	t.Run("隣接していれば期間と件数を合わせる", func(t *testing.T) {
		t.Parallel()
		older := NewCalendar(DateRange{From: d(1), To: d(10)}, map[Date]int{d(5): 2})
		got := newer.Merge(older)
		if got.Range != (DateRange{From: d(1), To: d(20)}) {
			t.Errorf("Range = %v", got.Range)
		}
		if got.Count(d(5)) != 2 || got.Count(d(15)) != 1 {
			t.Errorf("counts = %d, %d", got.Count(d(5)), got.Count(d(15)))
		}
		if newer.Count(d(5)) != 0 {
			t.Error("Merge must not mutate the receiver")
		}
	})

	t.Run("隙間があれば置き換える", func(t *testing.T) {
		t.Parallel()
		far := NewCalendar(DateRange{From: d(1), To: d(5)}, map[Date]int{d(2): 3})
		if got := newer.Merge(far); got.Range != far.Range || got.Count(d(15)) != 0 {
			t.Errorf("Merge with gap = %v", got.Range)
		}
	})
}
