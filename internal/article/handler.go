package article

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/blog-api/pkg/middleware"
	"github.com/example/blog-api/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type articleRequest struct {
	CategoryID *int64  `json:"category_id"`
	Title      string  `json:"title"`
	Slug       string  `json:"slug"`
	Excerpt    string  `json:"excerpt"`
	Content    string  `json:"content"`
	Status     string  `json:"status"`
	TagIDs     []int64 `json:"tag_ids"`
}

type statusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) PublicList(w http.ResponseWriter, r *http.Request) {
	filter := parseListFilter(r)
	filter.PublishedOnly = true
	h.list(w, r, filter)
}

func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	filter := parseListFilter(r)
	filter.Status = r.URL.Query().Get("status")
	h.list(w, r, filter)
}

func (h *Handler) PublicGet(w http.ResponseWriter, r *http.Request) {
	h.get(w, r, true)
}

func (h *Handler) AdminGet(w http.ResponseWriter, r *http.Request) {
	h.get(w, r, false)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req articleRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)

	a, err := h.service.Create(r.Context(), CreateInput{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Slug:       req.Slug,
		Excerpt:    req.Excerpt,
		Content:    req.Content,
		Status:     req.Status,
		TagIDs:     req.TagIDs,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrConflict) {
			status = http.StatusConflict
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, a)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req articleRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	a, err := h.service.Update(r.Context(), id, UpdateInput{
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Slug:       req.Slug,
		Excerpt:    req.Excerpt,
		Content:    req.Content,
		Status:     req.Status,
		TagIDs:     req.TagIDs,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
		}
		if errors.Is(err, ErrConflict) {
			status = http.StatusConflict
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, a)
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
	a, err := h.service.SetStatus(r.Context(), id, req.Status)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, a)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, filter ListFilter) {
	articles, err := h.service.List(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, articles)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, publishedOnly bool) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	a, err := h.service.Get(r.Context(), id, publishedOnly)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, a)
}

func parseListFilter(r *http.Request) ListFilter {
	categoryID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("category_id")), 10, 64)
	tagID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("tag_id")), 10, 64)
	return ListFilter{
		CategoryID: categoryID,
		TagID:      tagID,
		Query:      strings.TrimSpace(r.URL.Query().Get("q")),
	}
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "文章 ID 格式错误")
		return 0, false
	}
	return id, true
}
