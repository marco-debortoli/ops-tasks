package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// HeatmapWeeks is how many weeks the stats heatmap covers.
const HeatmapWeeks = 26

type TodayGroups struct {
	Overdue   []Task `json:"overdue"`
	DueToday  []Task `json:"due_today"`
	Scheduled []Task `json:"scheduled"`
	Completed []Task `json:"completed"`
}

type HeatDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type CategoryCount struct {
	CategoryID *int64 `json:"category_id"`
	Count      int    `json:"count"`
}

type Stats struct {
	Week  int `json:"week"`
	Month int `json:"month"`
	Year  int `json:"year"`
	// The same stretch of the previous week, month and year, for the up/down changes.
	PrevWeek   int             `json:"prev_week"`
	PrevMonth  int             `json:"prev_month"`
	PrevYear   int             `json:"prev_year"`
	Streak     int             `json:"streak"`
	Heatmap    []HeatDay       `json:"heatmap"`
	ByCategory []CategoryCount `json:"by_category"`
}

type Dashboard struct {
	Today      string      `json:"today"`
	Timezone   string      `json:"timezone"` // set by the caller; the zone "today" is calculated in
	Categories []Category  `json:"categories"`
	TodayTasks TodayGroups `json:"today_tasks"`
	Queue      []Task      `json:"queue"`
	Projects   []Project   `json:"projects"` // unfinished; pinned ones have Pinned set
	Stats      Stats       `json:"stats"`
}

// SplitOpenTasks sorts open tasks into Today's groups and the queue. The input order is kept
// within each group. A task lands in the first group that matches: overdue, due today,
// then scheduled today or earlier (so missed scheduled tasks roll forward instead of
// dropping into the queue). Everything else goes to the queue.
func SplitOpenTasks(open []Task, today string) (g TodayGroups, queue []Task) {
	g = TodayGroups{Overdue: []Task{}, DueToday: []Task{}, Scheduled: []Task{}, Completed: []Task{}}
	queue = []Task{}
	for _, t := range open {
		switch {
		case t.DueDate != nil && *t.DueDate < today:
			g.Overdue = append(g.Overdue, t)
		case t.DueDate != nil && *t.DueDate == today:
			g.DueToday = append(g.DueToday, t)
		case t.ScheduledDate != nil && *t.ScheduledDate <= today:
			g.Scheduled = append(g.Scheduled, t)
		default:
			queue = append(queue, t)
		}
	}
	return g, queue
}

// WeekStart returns the Monday of the week containing day.
func WeekStart(day time.Time) time.Time {
	offset := (int(day.Weekday()) + 6) % 7 // Monday = 0
	return day.AddDate(0, 0, -offset)
}

// HeatmapStart returns the first day (a Monday) of the heatmap ending in today's week.
func HeatmapStart(today time.Time) time.Time {
	return WeekStart(today).AddDate(0, 0, -7*(HeatmapWeeks-1))
}

// DateRange is an inclusive span of days.
type DateRange struct{ From, To time.Time }

// PreviousPeriods returns the stretches of last week, last month and last year that
// match how far today is into the current one: Monday to the same weekday last week,
// the 1st to the same day last month (or its last day, if shorter), and Jan 1 to the
// same date last year.
func PreviousPeriods(today time.Time) (week, month, year DateRange) {
	ws := WeekStart(today)
	week = DateRange{ws.AddDate(0, 0, -7), today.AddDate(0, 0, -7)}

	ms := time.Date(today.Year(), today.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	lastDay := ms.AddDate(0, 1, -1).Day()
	month = DateRange{ms, ms.AddDate(0, 0, min(today.Day(), lastDay)-1)}

	ys := time.Date(today.Year()-1, 1, 1, 0, 0, 0, 0, time.UTC)
	yEnd := time.Date(today.Year()-1, today.Month(), 1, 0, 0, 0, 0, time.UTC)
	yLast := yEnd.AddDate(0, 1, -1).Day()
	year = DateRange{ys, yEnd.AddDate(0, 0, min(today.Day(), yLast)-1)}
	return week, month, year
}

// Streak counts consecutive days with at least one completion, ending today, or
// yesterday when nothing has been completed yet today. days must be sorted newest first.
func Streak(days []string, today time.Time) int {
	have := make(map[string]bool, len(days))
	for _, d := range days {
		have[d] = true
	}
	day := today
	if !have[day.Format(time.DateOnly)] {
		day = day.AddDate(0, 0, -1)
	}
	n := 0
	for have[day.Format(time.DateOnly)] {
		n++
		day = day.AddDate(0, 0, -1)
	}
	return n
}

func (s *Store) Dashboard(ctx context.Context, today string) (*Dashboard, error) {
	day, err := time.Parse(time.DateOnly, today)
	if err != nil {
		return nil, err
	}
	d := &Dashboard{Today: today}

	if d.Categories, err = s.ListCategories(ctx); err != nil {
		return nil, err
	}
	if d.Categories == nil {
		d.Categories = []Category{}
	}
	// An archived project's open tasks are shelved with it.
	open, err := s.queryTasks(ctx, ` WHERE t.completed_on IS NULL AND p.archived_at IS NULL`+openOrder)
	if err != nil {
		return nil, err
	}
	d.TodayTasks, d.Queue = SplitOpenTasks(open, today)
	if d.TodayTasks.Completed, err = s.queryTasks(ctx, ` WHERE t.completed_on = $1::date ORDER BY t.completed_at DESC`, today); err != nil {
		return nil, err
	}
	if d.Projects, err = s.ListProjects(ctx, ProjectsActive); err != nil {
		return nil, err
	}
	if d.Stats, err = s.stats(ctx, day); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Store) stats(ctx context.Context, today time.Time) (Stats, error) {
	var st Stats
	todayS := today.Format(time.DateOnly)
	weekStart := WeekStart(today).Format(time.DateOnly)
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
	yearStart := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)

	pw, pm, py := PreviousPeriods(today)
	d := func(t time.Time) string { return t.Format(time.DateOnly) }

	err := s.pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE completed_on >= $1::date),
		       count(*) FILTER (WHERE completed_on >= $2::date),
		       count(*) FILTER (WHERE completed_on >= $3::date),
		       count(*) FILTER (WHERE completed_on BETWEEN $5::date AND $6::date),
		       count(*) FILTER (WHERE completed_on BETWEEN $7::date AND $8::date),
		       count(*) FILTER (WHERE completed_on BETWEEN $9::date AND $10::date)
		FROM tasks WHERE completed_on BETWEEN $9::date AND $4::date`,
		weekStart, monthStart, yearStart, todayS,
		d(pw.From), d(pw.To), d(pm.From), d(pm.To), d(py.From), d(py.To)).
		Scan(&st.Week, &st.Month, &st.Year, &st.PrevWeek, &st.PrevMonth, &st.PrevYear)
	if err != nil {
		return st, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT to_char(completed_on, 'YYYY-MM-DD') AS d FROM tasks
		WHERE completed_on <= $1::date ORDER BY d DESC`, todayS)
	if err != nil {
		return st, err
	}
	days, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return st, err
	}
	st.Streak = Streak(days, today)

	start := HeatmapStart(today)
	end := start.AddDate(0, 0, 7*HeatmapWeeks-1)
	rows, err = s.pool.Query(ctx, `
		SELECT to_char(completed_on, 'YYYY-MM-DD'), count(*) FROM tasks
		WHERE completed_on BETWEEN $1::date AND $2::date GROUP BY completed_on`,
		start.Format(time.DateOnly), todayS)
	if err != nil {
		return st, err
	}
	counts := map[string]int{}
	var date string
	var n int
	if _, err := pgx.ForEachRow(rows, []any{&date, &n}, func() error { counts[date] = n; return nil }); err != nil {
		return st, err
	}
	st.Heatmap = make([]HeatDay, 0, 7*HeatmapWeeks)
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		ds := day.Format(time.DateOnly)
		st.Heatmap = append(st.Heatmap, HeatDay{Date: ds, Count: counts[ds]})
	}

	rows, err = s.pool.Query(ctx, `
		SELECT COALESCE(p.category_id, t.category_id), count(*)
		FROM tasks t LEFT JOIN projects p ON p.id = t.project_id
		WHERE t.completed_on BETWEEN $1::date AND $2::date
		GROUP BY 1 ORDER BY 2 DESC`, yearStart, todayS)
	if err != nil {
		return st, err
	}
	st.ByCategory, err = pgx.CollectRows(rows, pgx.RowToStructByPos[CategoryCount])
	if st.ByCategory == nil {
		st.ByCategory = []CategoryCount{}
	}
	return st, err
}
