package store

import (
	"testing"
	"time"
)

func date(s string) time.Time {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return d
}

func str(s string) *string { return &s }

func names(ts []Task) []string {
	out := []string{}
	for _, t := range ts {
		out = append(out, t.Name)
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSplitOpenTasks(t *testing.T) {
	today := "2026-09-28"
	open := []Task{
		{Name: "late", DueDate: str("2026-09-25")},
		{Name: "late but scheduled ahead", DueDate: str("2026-09-27"), ScheduledDate: str("2026-10-01")},
		{Name: "due today", DueDate: str(today)},
		{Name: "due today, scheduled earlier", DueDate: str(today), ScheduledDate: str("2026-09-20")},
		{Name: "scheduled today", ScheduledDate: str(today)},
		{Name: "rolled forward", ScheduledDate: str("2026-09-21"), DueDate: str("2026-10-03")},
		{Name: "scheduled later", ScheduledDate: str("2026-09-29")},
		{Name: "due later", DueDate: str("2026-10-10")},
		{Name: "nothing set"},
	}
	g, queue := SplitOpenTasks(open, today)

	checks := []struct {
		group string
		got   []Task
		want  []string
	}{
		{"overdue", g.Overdue, []string{"late", "late but scheduled ahead"}},
		{"due today", g.DueToday, []string{"due today", "due today, scheduled earlier"}},
		{"scheduled", g.Scheduled, []string{"scheduled today", "rolled forward"}},
		{"queue", queue, []string{"scheduled later", "due later", "nothing set"}},
	}
	for _, c := range checks {
		if got := names(c.got); !eq(got, c.want) {
			t.Errorf("%s = %v, want %v", c.group, got, c.want)
		}
	}
}

func TestWeekStart(t *testing.T) {
	cases := map[string]string{
		"2026-09-28": "2026-09-28", // Monday
		"2026-09-30": "2026-09-28", // Wednesday
		"2026-10-04": "2026-09-28", // Sunday
		"2026-10-05": "2026-10-05", // next Monday
	}
	for in, want := range cases {
		if got := WeekStart(date(in)).Format(time.DateOnly); got != want {
			t.Errorf("WeekStart(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestHeatmapStart(t *testing.T) {
	got := HeatmapStart(date("2026-09-30"))
	if got.Weekday() != time.Monday {
		t.Fatalf("heatmap should start on a Monday, got %s", got.Weekday())
	}
	// 26 weeks ending with the week of 2026-09-28.
	if want := "2026-04-06"; got.Format(time.DateOnly) != want {
		t.Errorf("HeatmapStart = %s, want %s", got.Format(time.DateOnly), want)
	}
}

func TestStreak(t *testing.T) {
	today := date("2026-09-30")
	cases := []struct {
		name string
		days []string
		want int
	}{
		{"none", nil, 0},
		{"today only", []string{"2026-09-30"}, 1},
		{"run through today", []string{"2026-09-30", "2026-09-29", "2026-09-28", "2026-09-26"}, 3},
		{"nothing yet today keeps yesterday's run", []string{"2026-09-29", "2026-09-28"}, 2},
		{"broken", []string{"2026-09-28", "2026-09-27"}, 0},
		{"across a month boundary", []string{"2026-10-01", "2026-09-30"}, 1},
	}
	for _, c := range cases {
		if got := Streak(c.days, today); got != c.want {
			t.Errorf("%s: Streak = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestPreviousPeriods(t *testing.T) {
	cases := []struct {
		today             string
		week, month, year [2]string
	}{
		{"2026-09-30", // Wednesday
			[2]string{"2026-09-21", "2026-09-23"},
			[2]string{"2026-08-01", "2026-08-30"},
			[2]string{"2025-01-01", "2025-09-30"}},
		{"2026-03-31", // last month is shorter
			[2]string{"2026-03-23", "2026-03-24"},
			[2]string{"2026-02-01", "2026-02-28"},
			[2]string{"2025-01-01", "2025-03-31"}},
		{"2028-02-29", // leap day; last year had none
			[2]string{"2028-02-21", "2028-02-22"},
			[2]string{"2028-01-01", "2028-01-29"},
			[2]string{"2027-01-01", "2027-02-28"}},
		{"2026-01-01", // across the year boundary
			[2]string{"2025-12-22", "2025-12-25"},
			[2]string{"2025-12-01", "2025-12-01"},
			[2]string{"2025-01-01", "2025-01-01"}},
	}
	f := func(r DateRange) [2]string {
		return [2]string{r.From.Format(time.DateOnly), r.To.Format(time.DateOnly)}
	}
	for _, c := range cases {
		w, m, y := PreviousPeriods(date(c.today))
		if f(w) != c.week || f(m) != c.month || f(y) != c.year {
			t.Errorf("PreviousPeriods(%s) = %v %v %v, want %v %v %v", c.today, f(w), f(m), f(y), c.week, c.month, c.year)
		}
	}
}
