package user

import (
	"errors"
	"net/http"

	"github.com/example/blog-api/pkg/middleware"
	"github.com/example/blog-api/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	u, err := h.service.Register(r.Context(), RegisterInput{Username: req.Username, Password: req.Password})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrUsernameTaken) {
			status = http.StatusConflict
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, u)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	token, u, err := h.service.Login(r.Context(), LoginInput{Username: req.Username, Password: req.Password})
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, ErrInvalidPassword) {
			status = http.StatusInternalServerError
		}
		response.Error(w, status, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, loginResponse{Token: token, User: u})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Context().Value(middleware.ContextUserID).(int64)
	u, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, u)
}
