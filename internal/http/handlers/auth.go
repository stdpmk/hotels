package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/stdpmk/hotels/internal/http/response"
	"github.com/stdpmk/hotels/internal/services"
)

type AuthHandler struct {
	svc *services.UsersService
}

func NewAuthHandler(svc *services.UsersService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body", response.CodeBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "email, password and name are required", response.CodeBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Name)
	if errors.Is(err, services.ErrEmailTaken) {
		response.WriteError(w, http.StatusConflict, "email already taken", response.CodeEmailTaken)
		return
	}
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error", response.CodeInternal)
		return
	}

	response.WriteJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body", response.CodeBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		response.WriteError(w, http.StatusBadRequest, "email and password are required", response.CodeBadRequest)
		return
	}

	token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrWrongPassword) {
		response.WriteError(w, http.StatusUnauthorized, "invalid email or password", response.CodeInvalidCredentials)
		return
	}
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error", response.CodeInternal)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
