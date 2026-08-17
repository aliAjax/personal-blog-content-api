package tag

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

type tagRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.service.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, tags)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req tagRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	t, err := h.service.Create(r.Context(), CreateInput{Name: req.Name, Slug: req.Slug})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrConflict) {
			status = http.StatusConflict
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, t)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "标签 ID 格式错误")
		return
	}
	var req tagRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	t, err := h.service.Update(r.Context(), id, UpdateInput{Name: req.Name, Slug: req.Slug})
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
	response.JSON(w, http.StatusOK, t)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "标签 ID 格式错误")
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
