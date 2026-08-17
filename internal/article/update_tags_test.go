package article

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpdateKeepsTagsWhenTagIDsAreOmitted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	a := &Article{ID: 42, UserID: 7, Title: "new title", Slug: "new-title", Content: "body", Status: "draft"}
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE articles
		SET category_id = ?, title = ?, slug = ?, excerpt = ?, content = ?, status = ?, published_at = ?
		WHERE id = ?`)).
		WithArgs(a.CategoryID, a.Title, a.Slug, a.Excerpt, a.Content, a.Status, a.PublishedAt, a.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := NewRepository(db).Update(context.Background(), a, nil); err != nil {
		t.Fatalf("update without tag_ids: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("omitting tag_ids must not rewrite article_tags: %v", err)
	}
}
