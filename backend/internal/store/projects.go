package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusWaiting    = "waiting"
	StatusComplete   = "complete"
)

func validStatus(s string) bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusWaiting, StatusComplete:
		return true
	}
	return false
}

type TaskRef struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Priority *int   `json:"priority"`
}

type Project struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	CategoryID *int64    `json:"category_id"`
	Status     string    `json:"status"`
	DueDate    *string   `json:"due_date"`
	Notes      string    `json:"notes"`
	Pinned     bool      `json:"pinned"`
	OpenCount  int       `json:"open_count"`
	DoneCount  int       `json:"done_count"`
	NextTask   *TaskRef  `json:"next_task"` // highest-priority open task
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProjectDetail struct {
	Project
	OpenTasks      []Task `json:"open_tasks"`
	CompletedTasks []Task `json:"completed_tasks"`
}

type ProjectInput struct {
	Name       Opt[string] `json:"name"`
	CategoryID Opt[int64]  `json:"category_id"`
	Status     Opt[string] `json:"status"`
	DueDate    Opt[string] `json:"due_date"`
	Notes      Opt[string] `json:"notes"`
	Pinned     Opt[bool]   `json:"pinned"`
}

const projectSelect = `
SELECT p.id, p.name, p.category_id, p.status, to_char(p.due_date, 'YYYY-MM-DD'), p.notes, p.pinned,
       COALESCE(c.open, 0), COALESCE(c.done, 0), n.id, n.name, n.priority, p.created_at, p.updated_at
FROM projects p
LEFT JOIN LATERAL (
    SELECT count(*) FILTER (WHERE completed_on IS NULL) AS open,
           count(*) FILTER (WHERE completed_on IS NOT NULL) AS done
    FROM tasks WHERE project_id = p.id
) c ON true
LEFT JOIN LATERAL (
    SELECT id, name, priority FROM tasks
    WHERE project_id = p.id AND completed_on IS NULL
    ORDER BY priority NULLS LAST, due_date NULLS LAST, id LIMIT 1
) n ON true
`

func scanProject(row pgx.Row) (Project, error) {
	var p Project
	var nextID *int64
	var nextName *string
	var nextPri *int
	err := row.Scan(&p.ID, &p.Name, &p.CategoryID, &p.Status, &p.DueDate, &p.Notes, &p.Pinned,
		&p.OpenCount, &p.DoneCount, &nextID, &nextName, &nextPri, &p.CreatedAt, &p.UpdatedAt)
	if nextID != nil {
		p.NextTask = &TaskRef{ID: *nextID, Name: *nextName, Priority: nextPri}
	}
	return p, err
}

// ListProjects returns unfinished projects, or all of them when includeComplete is set.
func (s *Store) ListProjects(ctx context.Context, includeComplete bool) ([]Project, error) {
	rows, err := s.pool.Query(ctx, projectSelect+`
		WHERE $1 OR p.status <> 'complete'
		ORDER BY p.due_date NULLS LAST, p.name`, includeComplete)
	if err != nil {
		return nil, err
	}
	projects, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (Project, error) { return scanProject(r) })
	if err != nil {
		return nil, err
	}
	if projects == nil {
		projects = []Project{}
	}
	return projects, nil
}

func (s *Store) GetProject(ctx context.Context, id int64) (*ProjectDetail, error) {
	p, err := scanProject(s.pool.QueryRow(ctx, projectSelect+` WHERE p.id = $1`, id))
	if err != nil {
		return nil, notFound(err)
	}
	d := &ProjectDetail{Project: p}
	if d.OpenTasks, err = s.queryTasks(ctx, ` WHERE t.project_id = $1 AND t.completed_on IS NULL`+openOrder, id); err != nil {
		return nil, err
	}
	if d.CompletedTasks, err = s.queryTasks(ctx, ` WHERE t.project_id = $1 AND t.completed_on IS NOT NULL
		ORDER BY t.completed_on DESC, t.completed_at DESC`, id); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Store) CreateProject(ctx context.Context, in ProjectInput) (*ProjectDetail, error) {
	name, err := cleanName("name", in.Name.V)
	if err != nil {
		return nil, err
	}
	status := StatusTodo
	if in.Status.Valid {
		if !validStatus(in.Status.V) {
			return nil, invalid("unknown status %q", in.Status.V)
		}
		status = in.Status.V
	}
	due, err := cleanDate("due date", in.DueDate)
	if err != nil {
		return nil, err
	}
	var id int64
	err = s.pool.QueryRow(ctx, `
		INSERT INTO projects (name, category_id, status, due_date, notes)
		VALUES ($1, $2, $3, $4::date, $5) RETURNING id`,
		name, in.CategoryID.Ptr(), status, due, in.Notes.V).Scan(&id)
	if err != nil {
		return nil, dbError(err)
	}
	if in.Pinned.Valid && in.Pinned.V {
		return s.UpdateProject(ctx, id, ProjectInput{Pinned: in.Pinned})
	}
	return s.GetProject(ctx, id)
}

// UpdateProject applies a partial update, enforcing the pinning rules:
// one pinned project per category (pinning replaces the previous one), pinning needs
// a category, and changing a pinned project's category or completing it unpins it.
func (s *Store) UpdateProject(ctx context.Context, id int64, in ProjectInput) (*ProjectDetail, error) {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var (
			name, status, notes string
			categoryID          *int64
			due                 *string
			pinned              bool
		)
		err := tx.QueryRow(ctx, `
			SELECT name, category_id, status, to_char(due_date, 'YYYY-MM-DD'), notes, pinned
			FROM projects WHERE id = $1 FOR UPDATE`, id).
			Scan(&name, &categoryID, &status, &due, &notes, &pinned)
		if err != nil {
			return notFound(err)
		}

		if in.Name.Set {
			if name, err = cleanName("name", in.Name.V); err != nil {
				return err
			}
		}
		if in.Notes.Set {
			notes = in.Notes.V
		}
		if in.DueDate.Set {
			if due, err = cleanDate("due date", in.DueDate); err != nil {
				return err
			}
		}
		if in.Status.Set {
			if !in.Status.Valid || !validStatus(in.Status.V) {
				return invalid("unknown status %q", in.Status.V)
			}
			status = in.Status.V
			if status == StatusComplete {
				pinned = false
			}
		}
		if in.CategoryID.Set && !samePtr(categoryID, in.CategoryID.Ptr()) {
			categoryID = in.CategoryID.Ptr()
			pinned = false
		}
		if in.Pinned.Set {
			pinned = in.Pinned.Valid && in.Pinned.V
			if pinned && categoryID == nil {
				return invalid("give the project a category before pinning it")
			}
			if pinned && status == StatusComplete {
				return invalid("a completed project can't be pinned")
			}
		}
		if pinned {
			if _, err := tx.Exec(ctx, `UPDATE projects SET pinned = false, updated_at = now()
				WHERE category_id = $1 AND pinned AND id <> $2`, *categoryID, id); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `
			UPDATE projects SET name = $2, category_id = $3, status = $4, due_date = $5::date, notes = $6,
			       pinned = $7, updated_at = now()
			WHERE id = $1`, id, name, categoryID, status, due, notes, pinned)
		return dbError(err)
	})
	if err != nil {
		return nil, err
	}
	return s.GetProject(ctx, id)
}

// DeleteProject removes a project and keeps its tasks as standalone tasks,
// each taking the category it had through the project.
func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			UPDATE tasks t SET category_id = p.category_id, project_id = NULL, updated_at = now()
			FROM projects p WHERE p.id = $1 AND t.project_id = p.id`, id); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return nil
	})
}
