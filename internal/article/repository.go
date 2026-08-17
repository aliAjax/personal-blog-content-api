package article

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]Article, error)
	FindByID(ctx context.Context, id int64, publishedOnly bool) (*Article, error)
	Create(ctx context.Context, a *Article, tagIDs []int64) error
	Update(ctx context.Context, a *Article, tagIDs []int64) error
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const articleColumns = `
	a.id, a.user_id, a.category_id, a.title, a.slug, a.excerpt, a.content,
	a.status, a.published_at, a.created_at, a.updated_at,
	c.name, c.slug`

func (r *repository) List(ctx context.Context, filter ListFilter) ([]Article, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM articles a
		LEFT JOIN categories c ON c.id = a.category_id
		WHERE 1 = 1`, articleColumns)
	args := make([]any, 0)

	if filter.Status != "" {
		query += " AND a.status = ?"
		args = append(args, filter.Status)
	}
	if filter.PublishedOnly {
		query += " AND a.status = 'published'"
	}
	if filter.CategoryID > 0 {
		query += " AND a.category_id = ?"
		args = append(args, filter.CategoryID)
	}
	if filter.TagID > 0 {
		query += " AND a.id IN (SELECT article_id FROM article_tags WHERE tag_id = ?)"
		args = append(args, filter.TagID)
	}
	if strings.TrimSpace(filter.Query) != "" {
		query += " AND (a.title LIKE ? OR a.excerpt LIKE ? OR a.content LIKE ?)"
		keyword := "%" + strings.TrimSpace(filter.Query) + "%"
		args = append(args, keyword, keyword, keyword)
	}
	query += " ORDER BY a.created_at DESC, a.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]Article, 0)
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, *a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range articles {
		tags, err := r.findTags(ctx, articles[i].ID)
		if err != nil {
			return nil, err
		}
		articles[i].Tags = tags
	}
	return articles, nil
}

func (r *repository) FindByID(ctx context.Context, id int64, publishedOnly bool) (*Article, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM articles a
		LEFT JOIN categories c ON c.id = a.category_id
		WHERE a.id = ?`, articleColumns)
	if publishedOnly {
		query += " AND a.status = 'published'"
	}

	a, err := scanArticle(r.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.Tags, err = r.findTags(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *repository) Create(ctx context.Context, a *Article, tagIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO articles
			(user_id, category_id, title, slug, excerpt, content, status, published_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.UserID, a.CategoryID, a.Title, a.Slug, a.Excerpt, a.Content, a.Status, a.PublishedAt,
	)
	if err != nil {
		return err
	}
	a.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	if err := replaceTags(ctx, tx, a.ID, tagIDs); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *repository) Update(ctx context.Context, a *Article, tagIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE articles
		SET category_id = ?, title = ?, slug = ?, excerpt = ?, content = ?, status = ?, published_at = ?
		WHERE id = ?`,
		a.CategoryID, a.Title, a.Slug, a.Excerpt, a.Content, a.Status, a.PublishedAt, a.ID,
	)
	if err != nil {
		return err
	}
	// A nil tagIDs means the client omitted tag_ids from the request: keep the
	// article's existing tags untouched. A non-nil slice (including an empty
	// one) replaces the set. This lets a title-only edit preserve tags while a
	// request that sends "tag_ids": [] still clears them.
	if tagIDs != nil {
		if err := replaceTags(ctx, tx, a.ID, tagIDs); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM articles WHERE id = ?", id)
	return err
}

func (r *repository) findTags(ctx context.Context, articleID int64) ([]Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.slug
		FROM article_tags at
		JOIN tags t ON t.id = at.tag_id
		WHERE at.article_id = ?
		ORDER BY t.id ASC`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]Tag, 0)
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func replaceTags(ctx context.Context, tx *sql.Tx, articleID int64, tagIDs []int64) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM article_tags WHERE article_id = ?", articleID); err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if _, err := tx.ExecContext(ctx,
			"INSERT IGNORE INTO article_tags (article_id, tag_id) VALUES (?, ?)",
			articleID, tagID,
		); err != nil {
			return err
		}
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanArticle(row scanner) (*Article, error) {
	var a Article
	var categoryID sql.NullInt64
	var categoryName sql.NullString
	var categorySlug sql.NullString
	var publishedAt sql.NullTime

	err := row.Scan(
		&a.ID, &a.UserID, &categoryID, &a.Title, &a.Slug, &a.Excerpt, &a.Content,
		&a.Status, &publishedAt, &a.CreatedAt, &a.UpdatedAt,
		&categoryName, &categorySlug,
	)
	if err != nil {
		return nil, err
	}
	if categoryID.Valid {
		id := categoryID.Int64
		a.CategoryID = &id
		a.Category = &Category{ID: id, Name: categoryName.String, Slug: categorySlug.String}
	}
	if publishedAt.Valid {
		value := publishedAt.Time
		a.PublishedAt = &value
	}
	a.Tags = make([]Tag, 0)
	return &a, nil
}
