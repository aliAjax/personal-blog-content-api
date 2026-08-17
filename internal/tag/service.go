package tag

import (
	"context"
	"errors"
	"strings"

	"github.com/example/blog-api/pkg/slug"
)

var (
	ErrNotFound = errors.New("标签不存在")
	ErrConflict = errors.New("标签名称或 slug 已存在")
)

type Service interface {
	List(ctx context.Context) ([]Tag, error)
	Create(ctx context.Context, input CreateInput) (*Tag, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*Tag, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) ([]Tag, error) {
	return s.repo.List(ctx)
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Tag, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("标签名称不能为空")
	}
	slugValue := strings.TrimSpace(input.Slug)
	if slugValue == "" {
		slugValue = slug.Make(name)
	}
	if err := s.ensureUnique(ctx, 0, name, slugValue); err != nil {
		return nil, err
	}

	t := &Tag{Name: name, Slug: slugValue}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, t.ID)
}

func (s *service) Update(ctx context.Context, id int64, input UpdateInput) (*Tag, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrNotFound
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("标签名称不能为空")
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
