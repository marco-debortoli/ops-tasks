package store

import "testing"

func TestMonthDays(t *testing.T) {
	cases := []struct {
		first     string
		wantLen   int
		wantFirst string
		wantLast  string
	}{
		{"2026-09-01", 30, "2026-09-01", "2026-09-30"},
		{"2026-12-01", 31, "2026-12-01", "2026-12-31"}, // across a year boundary
		{"2026-02-01", 28, "2026-02-01", "2026-02-28"},
		{"2028-02-01", 29, "2028-02-01", "2028-02-29"}, // leap year
	}
	for _, c := range cases {
		days := MonthDays(date(c.first), nil)
		if len(days) != c.wantLen {
			t.Errorf("%s: %d days, want %d", c.first, len(days), c.wantLen)
			continue
		}
		if days[0].Date != c.wantFirst || days[len(days)-1].Date != c.wantLast {
			t.Errorf("%s: runs %s..%s, want %s..%s", c.first, days[0].Date, days[len(days)-1].Date, c.wantFirst, c.wantLast)
		}
	}

	days := MonthDays(date("2026-09-01"), map[string]int{"2026-09-24": 5, "2026-10-01": 9})
	if days[23].Count != 5 {
		t.Errorf("2026-09-24 count = %d, want 5", days[23].Count)
	}
	total := 0
	for _, d := range days {
		total += d.Count
	}
	if total != 5 {
		t.Errorf("month total = %d, want 5 (counts outside the month are ignored)", total)
	}
}
