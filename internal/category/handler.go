package category

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

type categoryRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, categories)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	c, err := h.service.Create(r.Context(), CreateInput{Name: req.Name, Slug: req.Slug, Description: req.Description})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrConflict) {
			status = http.StatusConflict
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, c)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "分类 ID 格式错误")
		return
	}
	var req categoryRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	c, err := h.service.Update(r.Context(), id, UpdateInput{Name: req.Name, Slug: req.Slug, Description: req.Description})
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
	response.JSON(w, http.StatusOK, c)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "分类 ID 格式错误")
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
