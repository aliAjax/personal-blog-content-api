package comment

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/example/blog-api/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type commentRequest struct {
	ParentID    *int64 `json:"parent_id"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	Content     string `json:"content"`
}

type statusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) PublicList(w http.ResponseWriter, r *http.Request) {
	articleID, ok := parseArticleID(w, r)
	if !ok {
		return
	}
	comments, err := h.service.PublicList(r.Context(), articleID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, comments)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	articleID, ok := parseArticleID(w, r)
	if !ok {
		return
	}
	var req commentRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	c, err := h.service.Create(r.Context(), CreateInput{
		ArticleID:   articleID,
		ParentID:    cloneParentID(req.ParentID),
		AuthorName:  req.AuthorName,
		AuthorEmail: req.AuthorEmail,
		Content:     req.Content,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrArticleClosed) {
			status = http.StatusNotFound
		}
		if errors.Is(err, ErrInvalidParent) {
			status = http.StatusUnprocessableEntity
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, c)
}

func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	comments, err := h.service.AdminList(r.Context(), r.URL.Query().Get("status"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, comments)
}

func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req statusRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}
	c, err := h.service.SetStatus(r.Context(), id, req.Status)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, c)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
		}
		response.Error(w, status, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseArticleID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "文章 ID 格式错误")
		return 0, false
	}
	return id, true
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "评论 ID 格式错误")
		return 0, false
	}
	return id, true
}
