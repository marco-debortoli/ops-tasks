package store

import (
	"context"
	"regexp"

	"github.com/jackc/pgx/v5"
)

type Category struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CategoryInput struct {
	Name  Opt[string] `json:"name"`
	Color Opt[string] `json:"color"`
}

var colorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (s *Store) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, color FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Category])
}

func (s *Store) CreateCategory(ctx context.Context, in CategoryInput) (*Category, error) {
	name, err := cleanName("name", in.Name.V)
	if err != nil {
		return nil, err
	}
	if !colorRe.MatchString(in.Color.V) {
		return nil, invalid("color must be a #rrggbb hex value")
	}
	var c Category
	err = s.pool.QueryRow(ctx, `INSERT INTO categories (name, color) VALUES ($1, $2) RETURNING id, name, color`,
		name, in.Color.V).Scan(&c.ID, &c.Name, &c.Color)
	if err != nil {
		return nil, dbError(err)
	}
	return &c, nil
}

func (s *Store) UpdateCategory(ctx context.Context, id int64, in CategoryInput) (*Category, error) {
	var c Category
	if err := s.pool.QueryRow(ctx, `SELECT id, name, color FROM categories WHERE id = $1`, id).Scan(&c.ID, &c.Name, &c.Color); err != nil {
		return nil, notFound(err)
	}
	if in.Name.Set {
		name, err := cleanName("name", in.Name.V)
		if err != nil {
			return nil, err
		}
		c.Name = name
	}
	if in.Color.Set {
		if !colorRe.MatchString(in.Color.V) {
			return nil, invalid("color must be a #rrggbb hex value")
		}
		c.Color = in.Color.V
	}
	if _, err := s.pool.Exec(ctx, `UPDATE categories SET name = $2, color = $3 WHERE id = $1`, id, c.Name, c.Color); err != nil {
		return nil, dbError(err)
	}
	return &c, nil
}

// DeleteCategory removes a category. Its projects and tasks become uncategorised,
// and a project pinned in it is unpinned (pinning needs a category).
func (s *Store) DeleteCategory(ctx context.Context, id int64) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE projects SET pinned = false WHERE category_id = $1 AND pinned`, id); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return nil
	})
}
