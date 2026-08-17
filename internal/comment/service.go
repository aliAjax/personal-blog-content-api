package comment

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrNotFound      = errors.New("评论不存在")
	ErrInvalidStatus = errors.New("评论状态只能是 pending、approved 或 rejected")
	ErrArticleClosed = errors.New("文章不存在或未发布")
)

type Service interface {
	PublicList(ctx context.Context, articleID int64) ([]Comment, error)
	Create(ctx context.Context, input CreateInput) (*Comment, error)
	AdminList(ctx context.Context, status string) ([]Comment, error)
	SetStatus(ctx context.Context, id int64, status string) (*Comment, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) PublicList(ctx context.Context, articleID int64) ([]Comment, error) {
	exists, err := s.repo.ArticleExistsPublished(ctx, articleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrArticleClosed
	}
	return s.repo.List(ctx, ListFilter{ArticleID: articleID, Status: "approved"})
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Comment, error) {
	if strings.TrimSpace(input.AuthorName) == "" {
		return nil, errors.New("评论人昵称不能为空")
	}
	if strings.TrimSpace(input.Content) == "" {
		return nil, errors.New("评论内容不能为空")
	}
	exists, err := s.repo.ArticleExistsPublished(ctx, input.ArticleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrArticleClosed
	}

	c := &Comment{
		ArticleID:   input.ArticleID,
		ParentID:    input.ParentID,
		AuthorName:  strings.TrimSpace(input.AuthorName),
		AuthorEmail: strings.TrimSpace(input.AuthorEmail),
		Content:     strings.TrimSpace(input.Content),
		Status:      "pending",
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, c.ID)
}

func (s *service) AdminList(ctx context.Context, status string) ([]Comment, error) {
	status = strings.TrimSpace(status)
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return nil, ErrInvalidStatus
	}
	return s.repo.List(ctx, ListFilter{Status: status})
}

func (s *service) SetStatus(ctx context.Context, id int64, status string) (*Comment, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "approved" && status != "rejected" && status != "pending" {
		return nil, ErrInvalidStatus
	}
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrNotFound
	}
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}
