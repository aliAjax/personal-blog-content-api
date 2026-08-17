package seed

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/example/blog-api/pkg/config"
	"github.com/example/blog-api/pkg/logger"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func Run(ctx context.Context, db *sql.DB, cfg config.Config) error {
	adminID, err := ensureAdmin(ctx, db, cfg.SeedAdminUsername, cfg.SeedAdminPassword)
	if err != nil {
		return err
	}

	golangID, err := ensureCategory(ctx, db, "Go 语言", "golang", "Go 开发、工程实践与语言特性")
	if err != nil {
		return err
	}
	dbID, err := ensureCategory(ctx, db, "数据库", "database", "MySQL 与数据存储相关笔记")
	if err != nil {
		return err
	}
	_, err = ensureCategory(ctx, db, "随笔", "life", "生活与工作随笔")
	if err != nil {
		return err
	}

	golangTag, err := ensureTag(ctx, db, "Golang", "golang")
	if err != nil {
		return err
	}
	mysqlTag, err := ensureTag(ctx, db, "MySQL", "mysql")
	if err != nil {
		return err
	}
	dockerTag, err := ensureTag(ctx, db, "Docker", "docker")
	if err != nil {
		return err
	}

	publishedAt := time.Now().UTC()
	publishedID, err := ensureArticle(ctx, db, adminID, golangID, &publishedAt,
		"欢迎来到我的博客",
		"welcome-to-my-blog",
		"这是一个基于 Go 和 MySQL 的博客后台 API。",
		"# 欢迎\n\n这里会分享 Golang、MySQL、Docker 等后端技术。",
		"published",
	)
	if err != nil {
		return err
	}
	_, err = ensureArticle(ctx, db, adminID, dbID, nil,
		"尚未完成的草稿",
		"draft-post",
		"草稿用于验证后台状态管理。",
		"这是一篇还没有发布的草稿，游客在前台不应看到它。",
		"draft",
	)
	if err != nil {
		return err
	}

	if err := linkTags(ctx, db, publishedID, []int64{golangTag, mysqlTag, dockerTag}); err != nil {
		return err
	}
	logger.Info("demo data seeded", "admin", cfg.SeedAdminUsername)
	return nil
}

func ensureAdmin(ctx context.Context, db *sql.DB, username, password string) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = ?", username).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	result, err := db.ExecContext(ctx,
		"INSERT INTO users (username, password_hash, role) VALUES (?, ?, 'admin')",
		username, string(hash),
	)
	if err != nil {
		if isDuplicateKey(err) {
			return lookupID(ctx, db, "SELECT id FROM users WHERE username = ?", username)
		}
		return 0, err
	}
	return result.LastInsertId()
}

func ensureCategory(ctx context.Context, db *sql.DB, name, slugValue, description string) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, "SELECT id FROM categories WHERE slug = ?", slugValue).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	result, err := db.ExecContext(ctx,
		"INSERT INTO categories (name, slug, description) VALUES (?, ?, ?)",
		name, slugValue, description,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return lookupID(ctx, db, "SELECT id FROM categories WHERE slug = ?", slugValue)
		}
		return 0, err
	}
	return result.LastInsertId()
}

func ensureTag(ctx context.Context, db *sql.DB, name, slugValue string) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, "SELECT id FROM tags WHERE slug = ?", slugValue).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	result, err := db.ExecContext(ctx, "INSERT INTO tags (name, slug) VALUES (?, ?)", name, slugValue)
	if err != nil {
		if isDuplicateKey(err) {
			return lookupID(ctx, db, "SELECT id FROM tags WHERE slug = ?", slugValue)
		}
		return 0, err
	}
	return result.LastInsertId()
}

func ensureArticle(
	ctx context.Context,
	db *sql.DB,
	userID int64,
	categoryID int64,
	publishedAt *time.Time,
	title string,
	slugValue string,
	excerpt string,
	content string,
	status string,
) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, "SELECT id FROM articles WHERE slug = ?", slugValue).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	var category any
	if categoryID > 0 {
		category = categoryID
	}
	result, err := db.ExecContext(ctx, `
		INSERT INTO articles
			(user_id, category_id, title, slug, excerpt, content, status, published_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, category, title, slugValue, excerpt, content, status, publishedAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return lookupID(ctx, db, "SELECT id FROM articles WHERE slug = ?", slugValue)
		}
		return 0, err
	}
	return result.LastInsertId()
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func lookupID(ctx context.Context, db *sql.DB, query string, value any) (int64, error) {
	var id int64
	if err := db.QueryRowContext(ctx, query, value).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func linkTags(ctx context.Context, db *sql.DB, articleID int64, tagIDs []int64) error {
	for _, tagID := range tagIDs {
		if _, err := db.ExecContext(ctx,
			"INSERT IGNORE INTO article_tags (article_id, tag_id) VALUES (?, ?)",
			articleID, tagID,
		); err != nil {
			return err
		}
	}
	return nil
}
