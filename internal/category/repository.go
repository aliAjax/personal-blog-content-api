package category

import (
	"context"
	"database/sql"
	"errors"
)

type Repository interface {
	List(ctx context.Context) ([]Category, error)
	FindByID(ctx context.Context, id int64) (*Category, error)
	FindByName(ctx context.Context, name string) (*Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const columns = "id, name, slug, description, created_at, updated_at"

func (r *repository) List(ctx context.Context) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+columns+" FROM categories ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *repository) FindByID(ctx context.Context, id int64) (*Category, error) {
	c, err := r.scanOne(r.db.QueryRowContext(ctx, "SELECT "+columns+" FROM categories WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *repository) FindByName(ctx context.Context, name string) (*Category, error) {
	c, err := r.scanOne(r.db.QueryRowContext(ctx, "SELECT "+columns+" FROM categories WHERE name = ?", name))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*Category, error) {
	c, err := r.scanOne(r.db.QueryRowContext(ctx, "SELECT "+columns+" FROM categories WHERE slug = ?", slug))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *repository) Create(ctx context.Context, c *Category) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO categories (name, slug, description) VALUES (?, ?, ?)",
		c.Name, c.Slug, c.Description,
	)
	if err != nil {
		return err
	}
	c.ID, err = result.LastInsertId()
	return err
}

func (r *repository) Update(ctx context.Context, c *Category) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE categories SET name = ?, slug = ?, description = ? WHERE id = ?",
		c.Name, c.Slug, c.Description, c.ID,
	)
	return err
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM categories WHERE id = ?", id)
	return err
}

func (r *repository) scanOne(row *sql.Row) (*Category, error) {
	var c Category
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
