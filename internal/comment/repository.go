package comment

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]Comment, error)
	FindByID(ctx context.Context, id int64) (*Comment, error)
	Create(ctx context.Context, c *Comment) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	Delete(ctx context.Context, id int64) error
	ArticleExistsPublished(ctx context.Context, articleID int64) (bool, error)
	ParentBelongsToArticle(ctx context.Context, parentID, articleID int64) (bool, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const columns = "id, article_id, parent_id, author_name, author_email, content, status, created_at, updated_at"

func (r *repository) List(ctx context.Context, filter ListFilter) ([]Comment, error) {
	query := "SELECT " + columns + " FROM comments WHERE 1 = 1"
	args := make([]any, 0)
	if filter.ArticleID > 0 {
		query += " AND article_id = ?"
		args = append(args, filter.ArticleID)
	}
	if strings.TrimSpace(filter.Status) != "" {
		query += " AND status = ?"
		args = append(args, strings.TrimSpace(filter.Status))
	}
	query += " ORDER BY created_at ASC, id ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]Comment, 0)
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, *c)
	}
	return comments, rows.Err()
}

func (r *repository) FindByID(ctx context.Context, id int64) (*Comment, error) {
	c, err := scanComment(r.db.QueryRowContext(ctx, "SELECT "+columns+" FROM comments WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *repository) Create(ctx context.Context, c *Comment) error {
	c.ParentID = cloneParentID(c.ParentID)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO comments (article_id, parent_id, author_name, author_email, content, status)
		VALUES (?, ?, ?, ?, ?, ?)`,
		c.ArticleID, c.ParentID, c.AuthorName, c.AuthorEmail, c.Content, c.Status,
	)
	if err != nil {
		return err
	}
	c.ID, err = result.LastInsertId()
	return err
}

func (r *repository) ParentBelongsToArticle(ctx context.Context, parentID, articleID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM comments WHERE id = ? AND article_id = ?",
		parentID, articleID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE comments SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM comments WHERE id = ?", id)
	return err
}

func (r *repository) ArticleExistsPublished(ctx context.Context, articleID int64) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM articles WHERE id = ? AND status = 'published'",
		articleID,
	).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanComment(row scanner) (*Comment, error) {
	var c Comment
	var parentID sql.NullInt64
	err := row.Scan(
		&c.ID, &c.ArticleID, &parentID, &c.AuthorName, &c.AuthorEmail,
		&c.Content, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if parentID.Valid {
		value := parentID.Int64
		c.ParentID = &value
	}
	return &c, nil
}
