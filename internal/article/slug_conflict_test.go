package article

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
)

type duplicateSlugRepo struct{}

func (duplicateSlugRepo) List(context.Context, ListFilter) ([]Article, error) { return nil, nil }
func (duplicateSlugRepo) FindByID(context.Context, int64, bool) (*Article, error) {
	return nil, nil
}
func (duplicateSlugRepo) Create(context.Context, *Article, []int64) error {
	return &mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'same-slug' for key 'uk_articles_slug'"}
}
func (duplicateSlugRepo) Update(context.Context, *Article, []int64) error { return nil }
func (duplicateSlugRepo) Delete(context.Context, int64) error             { return nil }

func TestCreateDuplicateSlugReturnsConflictWithoutDatabaseDetails(t *testing.T) {
	h := NewHandler(NewService(duplicateSlugRepo{}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/articles", strings.NewReader(`{
		"title":"Second post","slug":"same-slug","content":"body","status":"draft"
	}`))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "Duplicate entry") || strings.Contains(rec.Body.String(), "uk_articles_slug") {
		t.Fatalf("database details leaked: %s", rec.Body.String())
	}
}
