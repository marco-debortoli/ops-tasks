package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Task struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Priority      *int    `json:"priority"`
	ScheduledDate *string `json:"scheduled_date"`
	DueDate       *string `json:"due_date"`
	ProjectID     *int64  `json:"project_id"`
	ProjectName   *string `json:"project_name"`
	// CategoryID is the task's own category (standalone tasks only);
	// EffectiveCategoryID is what to display: the project's category, or the task's own.
	CategoryID          *int64     `json:"category_id"`
	EffectiveCategoryID *int64     `json:"effective_category_id"`
	CompletedOn         *string    `json:"completed_on"`
	CompletedAt         *time.Time `json:"completed_at"`
	SubtasksDone        int        `json:"subtasks_done"`
	SubtasksTotal       int        `json:"subtasks_total"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type TaskDetail struct {
	Task
	Subtasks []Subtask `json:"subtasks"`
}

type TaskInput struct {
	Name          Opt[string] `json:"name"`
	Description   Opt[string] `json:"description"`
	Priority      Opt[int]    `json:"priority"`
	ScheduledDate Opt[string] `json:"scheduled_date"`
	DueDate       Opt[string] `json:"due_date"`
	ProjectID     Opt[int64]  `json:"project_id"`
	CategoryID    Opt[int64]  `json:"category_id"`
	CompletedOn   Opt[string] `json:"completed_on"`
	// ScheduledToday (create only) schedules the task for the server's today.
	ScheduledToday bool `json:"scheduled_today"`
}

const taskSelect = `
SELECT t.id, t.name, t.description, t.priority,
       to_char(t.scheduled_date, 'YYYY-MM-DD'), to_char(t.due_date, 'YYYY-MM-DD'),
       t.project_id, p.name, t.category_id, COALESCE(p.category_id, t.category_id),
       to_char(t.completed_on, 'YYYY-MM-DD'), t.completed_at,
       COALESCE(st.done, 0), COALESCE(st.total, 0), t.created_at, t.updated_at
FROM tasks t
LEFT JOIN projects p ON p.id = t.project_id
LEFT JOIN LATERAL (
    SELECT count(*) FILTER (WHERE s.done) AS done, count(*) AS total
    FROM subtasks s WHERE s.task_id = t.id
) st ON true
`

// openOrder sorts open tasks: priority, then due date, then oldest first.
const openOrder = ` ORDER BY t.priority NULLS LAST, t.due_date NULLS LAST, t.id`

func scanTask(row pgx.Row) (Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.Name, &t.Description, &t.Priority,
		&t.ScheduledDate, &t.DueDate,
		&t.ProjectID, &t.ProjectName, &t.CategoryID, &t.EffectiveCategoryID,
		&t.CompletedOn, &t.CompletedAt,
		&t.SubtasksDone, &t.SubtasksTotal, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (s *Store) queryTasks(ctx context.Context, where string, args ...any) ([]Task, error) {
	rows, err := s.pool.Query(ctx, taskSelect+where, args...)
	if err != nil {
		return nil, err
	}
	tasks, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (Task, error) { return scanTask(r) })
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, nil
}

func (s *Store) GetTask(ctx context.Context, id int64) (*TaskDetail, error) {
	t, err := scanTask(s.pool.QueryRow(ctx, taskSelect+` WHERE t.id = $1`, id))
	if err != nil {
		return nil, notFound(err)
	}
	subs, err := s.listSubtasks(ctx, id)
	if err != nil {
		return nil, err
	}
	return &TaskDetail{Task: t, Subtasks: subs}, nil
}

// taskRow is the editable part of a task, loaded, changed and written back by UpdateTask.
type taskRow struct {
	name, description string
	priority          *int
	scheduled, due    *string
	projectID         *int64
	categoryID        *int64
	completedOn       *string
}

func validPriority(o Opt[int]) (*int, error) {
	if o.Valid && (o.V < 1 || o.V > 3) {
		return nil, invalid("priority must be 1, 2, 3 or null")
	}
	return o.Ptr(), nil
}

func (s *Store) CreateTask(ctx context.Context, in TaskInput, today string) (*TaskDetail, error) {
	var r taskRow
	var err error
	if r.name, err = cleanName("name", in.Name.V); err != nil {
		return nil, err
	}
	r.description = in.Description.V
	if r.priority, err = validPriority(in.Priority); err != nil {
		return nil, err
	}
	if r.scheduled, err = cleanDate("scheduled date", in.ScheduledDate); err != nil {
		return nil, err
	}
	if in.ScheduledToday {
		r.scheduled = &today
	}
	if r.due, err = cleanDate("due date", in.DueDate); err != nil {
		return nil, err
	}
	r.projectID = in.ProjectID.Ptr()
	if r.projectID == nil {
		// A task in a project always takes the project's category.
		r.categoryID = in.CategoryID.Ptr()
	}
	var id int64
	err = s.pool.QueryRow(ctx, `
		INSERT INTO tasks (name, description, priority, scheduled_date, due_date, project_id, category_id)
		VALUES ($1, $2, $3, $4::date, $5::date, $6, $7) RETURNING id`,
		r.name, r.description, r.priority, r.scheduled, r.due, r.projectID, r.categoryID).Scan(&id)
	if err != nil {
		return nil, dbError(err)
	}
	return s.GetTask(ctx, id)
}

func (s *Store) UpdateTask(ctx context.Context, id int64, in TaskInput) (*TaskDetail, error) {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var r taskRow
		err := tx.QueryRow(ctx, `
			SELECT name, description, priority, to_char(scheduled_date, 'YYYY-MM-DD'), to_char(due_date, 'YYYY-MM-DD'),
			       project_id, category_id, to_char(completed_on, 'YYYY-MM-DD')
			FROM tasks WHERE id = $1 FOR UPDATE`, id).
			Scan(&r.name, &r.description, &r.priority, &r.scheduled, &r.due, &r.projectID, &r.categoryID, &r.completedOn)
		if err != nil {
			return notFound(err)
		}

		if in.Name.Set {
			if r.name, err = cleanName("name", in.Name.V); err != nil {
				return err
			}
		}
		if in.Description.Set {
			r.description = in.Description.V
		}
		if in.Priority.Set {
			if r.priority, err = validPriority(in.Priority); err != nil {
				return err
			}
		}
		if in.ScheduledDate.Set {
			if r.scheduled, err = cleanDate("scheduled date", in.ScheduledDate); err != nil {
				return err
			}
		}
		if in.DueDate.Set {
			if r.due, err = cleanDate("due date", in.DueDate); err != nil {
				return err
			}
		}
		if in.ProjectID.Set && !samePtr(r.projectID, in.ProjectID.Ptr()) {
			if r.projectID != nil && !in.ProjectID.Valid {
				// Moving out of a project keeps the category the task had through it.
				if err := tx.QueryRow(ctx, `SELECT category_id FROM projects WHERE id = $1`, *r.projectID).Scan(&r.categoryID); err != nil {
					return err
				}
			} else {
				r.categoryID = nil
			}
			r.projectID = in.ProjectID.Ptr()
		}
		if in.CategoryID.Set {
			if r.projectID != nil {
				return invalid("a task in a project uses the project's category")
			}
			r.categoryID = in.CategoryID.Ptr()
		}
		if in.CompletedOn.Set {
			if r.completedOn == nil {
				return invalid("only a completed task has a completion date")
			}
			if !in.CompletedOn.Valid {
				return invalid("reopen the task instead of clearing its completion date")
			}
			if r.completedOn, err = cleanDate("completion date", in.CompletedOn); err != nil {
				return err
			}
		}

		_, err = tx.Exec(ctx, `
			UPDATE tasks SET name = $2, description = $3, priority = $4, scheduled_date = $5::date, due_date = $6::date,
			       project_id = $7, category_id = $8, completed_on = $9::date, updated_at = now()
			WHERE id = $1`,
			id, r.name, r.description, r.priority, r.scheduled, r.due, r.projectID, r.categoryID, r.completedOn)
		return dbError(err)
	})
	if err != nil {
		return nil, err
	}
	return s.GetTask(ctx, id)
}

// CompleteTask marks an open task complete on the given date, recording the actual time.
// Subtasks are left as they are.
func (s *Store) CompleteTask(ctx context.Context, id int64, on string) (*TaskDetail, error) {
	_, err := s.pool.Exec(ctx, `
		UPDATE tasks SET completed_on = $2::date, completed_at = now(), updated_at = now()
		WHERE id = $1 AND completed_on IS NULL`, id, on)
	if err != nil {
		return nil, err
	}
	return s.GetTask(ctx, id)
}

func (s *Store) ReopenTask(ctx context.Context, id int64) (*TaskDetail, error) {
	_, err := s.pool.Exec(ctx, `
		UPDATE tasks SET completed_on = NULL, completed_at = NULL, updated_at = now()
		WHERE id = $1 AND completed_on IS NOT NULL`, id)
	if err != nil {
		return nil, err
	}
	return s.GetTask(ctx, id)
}

func (s *Store) DeleteTask(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
