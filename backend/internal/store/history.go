package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// History is what the history modal shows: one day's completions and a month of daily counts.
type History struct {
	Day   string    `json:"day"`
	Tasks []Task    `json:"tasks"` // completed on Day, in the order they were done
	Month string    `json:"month"` // YYYY-MM
	Days  []HeatDay `json:"days"`  // every day of Month, zero counts included
}

const monthLayout = "2006-01"

// MonthDays lists every day of the month starting at first, with its count (0 when missing).
func MonthDays(first time.Time, counts map[string]int) []HeatDay {
	next := first.AddDate(0, 1, 0)
	days := make([]HeatDay, 0, 31)
	for day := first; day.Before(next); day = day.AddDate(0, 0, 1) {
		ds := day.Format(time.DateOnly)
		days = append(days, HeatDay{Date: ds, Count: counts[ds]})
	}
	return days
}

// History loads the tasks completed on day and the per-day counts for month.
// An empty month means the month containing day.
func (s *Store) History(ctx context.Context, day, month string) (*History, error) {
	d, err := time.Parse(time.DateOnly, day)
	if err != nil {
		return nil, invalid("day must be a YYYY-MM-DD date")
	}
	first := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
	if month != "" {
		if first, err = time.Parse(monthLayout, month); err != nil {
			return nil, invalid("month must be YYYY-MM")
		}
	}
	h := &History{Day: day, Month: first.Format(monthLayout)}

	if h.Tasks, err = s.queryTasks(ctx, ` WHERE t.completed_on = $1::date ORDER BY t.completed_at, t.id`, day); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT to_char(completed_on, 'YYYY-MM-DD'), count(*) FROM tasks
		WHERE completed_on >= $1::date AND completed_on < $2::date GROUP BY completed_on`,
		first.Format(time.DateOnly), first.AddDate(0, 1, 0).Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	var date string
	var n int
	if _, err := pgx.ForEachRow(rows, []any{&date, &n}, func() error { counts[date] = n; return nil }); err != nil {
		return nil, err
	}
	h.Days = MonthDays(first, counts)
	return h, nil
}
