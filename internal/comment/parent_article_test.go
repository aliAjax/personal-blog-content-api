package comment

import (
	"context"
	"testing"
)

type parentRepo struct {
	created bool
}

func (r *parentRepo) List(context.Context, ListFilter) ([]Comment, error) { return nil, nil }
func (r *parentRepo) FindByID(context.Context, int64) (*Comment, error) {
	return &Comment{ID: 99}, nil
}
func (r *parentRepo) Create(_ context.Context, c *Comment) error {
	r.created = true
	c.ID = 99
	return nil
}
func (r *parentRepo) UpdateStatus(context.Context, int64, string) error { return nil }
func (r *parentRepo) Delete(context.Context, int64) error               { return nil }
func (r *parentRepo) ArticleExistsPublished(context.Context, int64) (bool, error) {
	return true, nil
}
func (r *parentRepo) ParentBelongsToArticle(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func TestCreateRejectsParentFromAnotherArticle(t *testing.T) {
	repo := &parentRepo{}
	parentID := int64(8)
	_, err := NewService(repo).Create(context.Background(), CreateInput{
		ArticleID:  12,
		ParentID:   &parentID,
		AuthorName: "reader",
		Content:    "reply",
	})
	if err == nil {
		t.Fatal("expected a cross-article parent to be rejected")
	}
	if repo.created {
		t.Fatal("invalid reply was persisted")
	}
}
