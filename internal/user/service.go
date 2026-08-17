package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/example/blog-api/pkg/middleware"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken   = errors.New("用户名已存在")
	ErrInvalidPassword = errors.New("用户名或密码错误")
	ErrNotFound        = errors.New("用户不存在")
)

type Service interface {
	Register(ctx context.Context, input RegisterInput) (*User, error)
	Login(ctx context.Context, input LoginInput) (string, *User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
}

type service struct {
	repo      Repository
	jwtSecret string
	tokenTTL  time.Duration
}

func NewService(repo Repository, jwtSecret string, tokenTTL time.Duration) Service {
	return &service{repo: repo, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

func (s *service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	username := strings.TrimSpace(input.Username)
	if len(username) < 3 || len(username) > 64 {
		return nil, errors.New("用户名长度应为 3 到 64 个字符")
	}
	if len(input.Password) < 6 {
		return nil, errors.New("密码长度至少为 6 位")
	}

	existing, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "user",
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, u.ID)
}

func (s *service) Login(ctx context.Context, input LoginInput) (string, *User, error) {
	u, err := s.repo.FindByUsername(ctx, strings.TrimSpace(input.Username))
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, ErrInvalidPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		return "", nil, ErrInvalidPassword
	}

	now := time.Now()
	claims := &middleware.Claims{
		UserID:   u.ID,
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}
	return signed, u, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}
