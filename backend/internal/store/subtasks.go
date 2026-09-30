package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Subtask struct {
	ID       int64  `json:"id"`
	TaskID   int64  `json:"task_id"`
	Name     string `json:"name"`
	Done     bool   `json:"done"`
	Position int    `json:"position"`
}

type SubtaskInput struct {
	Name Opt[string] `json:"name"`
	Done Opt[bool]   `json:"done"`
}

const subtaskSelect = `SELECT id, task_id, name, done, position FROM subtasks`

func (s *Store) listSubtasks(ctx context.Context, taskID int64) ([]Subtask, error) {
	rows, err := s.pool.Query(ctx, subtaskSelect+` WHERE task_id = $1 ORDER BY position, id`, taskID)
	if err != nil {
		return nil, err
	}
	subs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Subtask])
	if err != nil {
		return nil, err
	}
	if subs == nil {
		subs = []Subtask{}
	}
	return subs, nil
}

func (s *Store) touchTask(ctx context.Context, taskID int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE tasks SET updated_at = now() WHERE id = $1`, taskID)
	return err
}

func (s *Store) CreateSubtask(ctx context.Context, taskID int64, in SubtaskInput) (*Subtask, error) {
	name, err := cleanName("name", in.Name.V)
	if err != nil {
		return nil, err
	}
	var sub Subtask
	err = s.pool.QueryRow(ctx, `
		INSERT INTO subtasks (task_id, name, position)
		VALUES ($1, $2, (SELECT COALESCE(max(position), -1) + 1 FROM subtasks WHERE task_id = $1))
		RETURNING id, task_id, name, done, position`, taskID, name).
		Scan(&sub.ID, &sub.TaskID, &sub.Name, &sub.Done, &sub.Position)
	if err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrNotFound // the only foreign key is the task
		}
		return nil, err
	}
	return &sub, s.touchTask(ctx, taskID)
}

func (s *Store) UpdateSubtask(ctx context.Context, id int64, in SubtaskInput) (*Subtask, error) {
	var sub Subtask
	if err := s.pool.QueryRow(ctx, subtaskSelect+` WHERE id = $1`, id).
		Scan(&sub.ID, &sub.TaskID, &sub.Name, &sub.Done, &sub.Position); err != nil {
		return nil, notFound(err)
	}
	if in.Name.Set {
		name, err := cleanName("name", in.Name.V)
		if err != nil {
			return nil, err
		}
		sub.Name = name
	}
	if in.Done.Set {
		sub.Done = in.Done.V
	}
	if _, err := s.pool.Exec(ctx, `UPDATE subtasks SET name = $2, done = $3 WHERE id = $1`, id, sub.Name, sub.Done); err != nil {
		return nil, err
	}
	return &sub, s.touchTask(ctx, sub.TaskID)
}

func (s *Store) DeleteSubtask(ctx context.Context, id int64) error {
	var taskID int64
	if err := s.pool.QueryRow(ctx, `DELETE FROM subtasks WHERE id = $1 RETURNING task_id`, id).Scan(&taskID); err != nil {
		return notFound(err)
	}
	return s.touchTask(ctx, taskID)
}

// ReorderSubtasks sets the order of a task's subtasks; ids must list each of them exactly once.
func (s *Store) ReorderSubtasks(ctx context.Context, taskID int64, ids []int64) ([]Subtask, error) {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT id FROM subtasks WHERE task_id = $1 FOR UPDATE`, taskID)
		if err != nil {
			return err
		}
		existing, err := pgx.CollectRows(rows, pgx.RowTo[int64])
		if err != nil {
			return err
		}
		want := make(map[int64]bool, len(existing))
		for _, id := range existing {
			want[id] = true
		}
		if len(ids) != len(existing) {
			return invalid("the new order must list every subtask of the task once")
		}
		for i, id := range ids {
			if !want[id] {
				return invalid("the new order must list every subtask of the task once")
			}
			delete(want, id)
			if _, err := tx.Exec(ctx, `UPDATE subtasks SET position = $2 WHERE id = $1`, id, i); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `UPDATE tasks SET updated_at = now() WHERE id = $1`, taskID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.listSubtasks(ctx, taskID)
}
