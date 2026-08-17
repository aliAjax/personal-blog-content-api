package article

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound      = errors.New("文章不存在")
	ErrInvalidStatus = errors.New("文章状态只能是 draft 或 published")
	ErrConflict      = errors.New("文章 slug 已存在")
)

type Service interface {
	List(ctx context.Context, filter ListFilter) ([]Article, error)
	Get(ctx context.Context, id int64, publishedOnly bool) (*Article, error)
	Create(ctx context.Context, input CreateInput) (*Article, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*Article, error)
	Delete(ctx context.Context, id int64) error
	SetStatus(ctx context.Context, id int64, status string) (*Article, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, filter ListFilter) ([]Article, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Get(ctx context.Context, id int64, publishedOnly bool) (*Article, error) {
	a, err := s.repo.FindByID(ctx, id, publishedOnly)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Article, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, errors.New("文章标题不能为空")
	}
	if strings.TrimSpace(input.Content) == "" {
		return nil, errors.New("文章内容不能为空")
	}

	status, publishedAt, err := normalizeStatus(input.Status, nil)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(input.Title)
	slugValue := normalizedSlug(input.Slug, title)

	a := &Article{
		UserID:      input.UserID,
		CategoryID:  input.CategoryID,
		Title:       title,
		Slug:        slugValue,
		Excerpt:     strings.TrimSpace(input.Excerpt),
		Content:     input.Content,
		Status:      status,
		PublishedAt: publishedAt,
		Tags:        make([]Tag, 0),
	}
	if err := s.repo.Create(ctx, a, input.TagIDs); err != nil {
		return nil, classifyWriteError(err)
	}
	return s.repo.FindByID(ctx, a.ID, false)
}

func (s *service) Update(ctx context.Context, id int64, input UpdateInput) (*Article, error) {
	current, err := s.repo.FindByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrNotFound
	}
	if strings.TrimSpace(input.Title) == "" {
		return nil, errors.New("文章标题不能为空")
	}
	if strings.TrimSpace(input.Content) == "" {
		return nil, errors.New("文章内容不能为空")
	}

	status := input.Status
	if strings.TrimSpace(status) == "" {
		status = current.Status
	}
	normalized, publishedAt, err := normalizeStatus(status, current.PublishedAt)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(input.Title)
	slugValue := normalizedSlug(input.Slug, title)

	current.CategoryID = input.CategoryID
	current.Title = title
	current.Slug = slugValue
	current.Excerpt = strings.TrimSpace(input.Excerpt)
	current.Content = input.Content
	current.Status = normalized
	current.PublishedAt = publishedAt
	if err := s.repo.Update(ctx, current, input.TagIDs); err != nil {
		return nil, classifyWriteError(err)
	}
	return s.repo.FindByID(ctx, current.ID, false)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	current, err := s.repo.FindByID(ctx, id, false)
	if err != nil {
		return err
	}
	if current == nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *service) SetStatus(ctx context.Context, id int64, status string) (*Article, error) {
	current, err := s.repo.FindByID(ctx, id, false)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrNotFound
	}

	normalized, publishedAt, err := normalizeStatus(status, current.PublishedAt)
	if err != nil {
		return nil, err
	}
	current.Status = normalized
	current.PublishedAt = publishedAt
	if err := s.repo.Update(ctx, current, tagIDsOf(current.Tags)); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, current.ID, false)
}

func normalizeStatus(status string, currentPublishedAt *time.Time) (string, *time.Time, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "draft"
	}
	if status != "draft" && status != "published" {
		return "", nil, ErrInvalidStatus
	}
	if status == "draft" {
		return status, nil, nil
	}
	if currentPublishedAt != nil {
		value := *currentPublishedAt
		return status, &value, nil
	}
	now := time.Now().UTC()
	return status, &now, nil
}

func tagIDsOf(tags []Tag) []int64 {
	ids := make([]int64, 0, len(tags))
	for _, t := range tags {
		ids = append(ids, t.ID)
	}
	return ids
}

func classifyWriteError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrConflict
	}
	return err
}
