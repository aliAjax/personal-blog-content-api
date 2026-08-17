package tag

import (
	"context"
	"database/sql"
	"errors"
)

type Repository interface {
	List(ctx context.Context) ([]Tag, error)
	FindByID(ctx context.Context, id int64) (*Tag, error)
	FindByName(ctx context.Context, name string) (*Tag, error)
	FindBySlug(ctx context.Context, slug string) (*Tag, error)
	Create(ctx context.Context, t *Tag) error
	Update(ctx context.Context, t *Tag) error
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const columns = "id, name, slug, created_at, updated_at"

func (r *repository) List(ctx context.Context) ([]Tag, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+columns+" FROM tags ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]Tag, 0)
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *repository) FindByID(ctx context.Context, id int64) (*Tag, error) {
	t, err := r.scanOne(r.db.QueryRowContext(ctx, "SELECT "+columns+" FROM tags WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func (r *repository) FindByName(ctx context.Context, name string) (*Tag, error) {
	t, err := r.scanOne(r.db.QueryRowContext(ctx, "SELECT "+columns+" FROM tags WHERE name = ?", name))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*Tag, error) {
	t, err := r.scanOne(r.db.QueryRowContext(ctx, "SELECT "+columns+" FROM tags WHERE slug = ?", slug))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func (r *repository) Create(ctx context.Context, t *Tag) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO tags (name, slug) VALUES (?, ?)",
		t.Name, t.Slug,
	)
	if err != nil {
		return err
	}
	t.ID, err = result.LastInsertId()
	return err
}

func (r *repository) Update(ctx context.Context, t *Tag) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE tags SET name = ?, slug = ? WHERE id = ?",
		t.Name, t.Slug, t.ID,
	)
	return err
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tags WHERE id = ?", id)
	return err
}

func (r *repository) scanOne(row *sql.Row) (*Tag, error) {
	var t Tag
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
