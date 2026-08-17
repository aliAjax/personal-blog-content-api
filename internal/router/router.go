package router

import (
	"database/sql"
	"net/http"

	"github.com/example/blog-api/internal/article"
	"github.com/example/blog-api/internal/category"
	"github.com/example/blog-api/internal/comment"
	"github.com/example/blog-api/internal/tag"
	"github.com/example/blog-api/internal/user"
	"github.com/example/blog-api/pkg/config"
	"github.com/example/blog-api/pkg/middleware"
	"github.com/example/blog-api/pkg/response"
)

func New(db *sql.DB, cfg config.Config) http.Handler {
	userService := user.NewService(user.NewRepository(db), cfg.JWTSecret, cfg.TokenTTL)
	userHandler := user.NewHandler(userService)

	categoryService := category.NewService(category.NewRepository(db))
	categoryHandler := category.NewHandler(categoryService)

	tagService := tag.NewService(tag.NewRepository(db))
	tagHandler := tag.NewHandler(tagService)

	articleService := article.NewService(article.NewRepository(db))
	articleHandler := article.NewHandler(articleService)

	commentService := comment.NewService(comment.NewRepository(db))
	commentHandler := comment.NewHandler(commentService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/v1/auth/register", userHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", userHandler.Login)

	auth := middleware.Authenticate(cfg.JWTSecret)
	adminOnly := middleware.RequireAdmin
	protected := func(handler http.HandlerFunc) http.Handler {
		return auth(adminOnly(http.HandlerFunc(handler)))
	}

	mux.Handle("GET /api/v1/auth/me", auth(http.HandlerFunc(userHandler.Me)))

	mux.HandleFunc("GET /api/v1/categories", categoryHandler.List)
	mux.HandleFunc("GET /api/v1/tags", tagHandler.List)
	mux.HandleFunc("GET /api/v1/articles", articleHandler.PublicList)
	mux.HandleFunc("GET /api/v1/articles/{id}", articleHandler.PublicGet)
	mux.HandleFunc("GET /api/v1/articles/{id}/comments", commentHandler.PublicList)
	mux.HandleFunc("POST /api/v1/articles/{id}/comments", commentHandler.Create)

	mux.Handle("POST /api/v1/admin/categories", protected(categoryHandler.Create))
	mux.Handle("GET /api/v1/admin/categories", protected(categoryHandler.List))
	mux.Handle("PUT /api/v1/admin/categories/{id}", protected(categoryHandler.Update))
	mux.Handle("DELETE /api/v1/admin/categories/{id}", protected(categoryHandler.Delete))

	mux.Handle("POST /api/v1/admin/tags", protected(tagHandler.Create))
	mux.Handle("GET /api/v1/admin/tags", protected(tagHandler.List))
	mux.Handle("PUT /api/v1/admin/tags/{id}", protected(tagHandler.Update))
	mux.Handle("DELETE /api/v1/admin/tags/{id}", protected(tagHandler.Delete))

	mux.Handle("GET /api/v1/admin/articles", protected(articleHandler.AdminList))
	mux.Handle("GET /api/v1/admin/articles/{id}", protected(articleHandler.AdminGet))
	mux.Handle("POST /api/v1/admin/articles", protected(articleHandler.Create))
	mux.Handle("PUT /api/v1/admin/articles/{id}", protected(articleHandler.Update))
	mux.Handle("PATCH /api/v1/admin/articles/{id}/status", protected(articleHandler.SetStatus))
	mux.Handle("DELETE /api/v1/admin/articles/{id}", protected(articleHandler.Delete))

	mux.Handle("GET /api/v1/admin/comments", protected(commentHandler.AdminList))
	mux.Handle("PATCH /api/v1/admin/comments/{id}", protected(commentHandler.SetStatus))
	mux.Handle("DELETE /api/v1/admin/comments/{id}", protected(commentHandler.Delete))

	return middleware.Recover(middleware.CORS(mux))
}
