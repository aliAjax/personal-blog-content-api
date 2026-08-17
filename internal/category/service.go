package category

import (
	"context"
	"errors"
	"strings"

	"github.com/example/blog-api/pkg/slug"
)

var (
	ErrNotFound = errors.New("分类不存在")
	ErrConflict = errors.New("分类名称或 slug 已存在")
)

type Service interface {
	List(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, input CreateInput) (*Category, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*Category, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) ([]Category, error) {
	return s.repo.List(ctx)
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Category, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("分类名称不能为空")
	}
	slugValue := strings.TrimSpace(input.Slug)
	if slugValue == "" {
		slugValue = slug.Make(name)
	}
	if err := s.ensureUnique(ctx, 0, name, slugValue); err != nil {
		return nil, err
	}

	c := &Category{Name: name, Slug: slugValue, Description: strings.TrimSpace(input.Description)}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, c.ID)
}

func (s *service) Update(ctx context.Context, id int64, input UpdateInput) (*Category, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrNotFound
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("分类名称不能为空")
	}
	slugValue := strings.TrimSpace(input.Slug)
	if slugValue == "" {
		slugValue = slug.Make(name)
	}
	if err := s.ensureUnique(ctx, id, name, slugValue); err != nil {
		return nil, err
	}

	current.Name = name
	current.Slug = slugValue
	current.Description = strings.TrimSpace(input.Description)
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, current.ID)
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

func (s *service) ensureUnique(ctx context.Context, excludeID int64, name, slugValue string) error {
	byName, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return err
	}
	if byName != nil && byName.ID != excludeID {
		return ErrConflict
	}
	bySlug, err := s.repo.FindBySlug(ctx, slugValue)
	if err != nil {
		return err
	}
	if bySlug != nil && bySlug.ID != excludeID {
		return ErrConflict
	}
	return nil
}
