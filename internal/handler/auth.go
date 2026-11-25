package handler

import (
	"net/http"

	"github.com/flashy-team/flashy-backend/internal/domain"
)

// AuthHandler обробляє реєстрацію, вхід та профіль. Лише валідує запит і делегує сервісу.
type AuthHandler struct{ auth AuthUseCase }

func NewAuthHandler(a AuthUseCase) *AuthHandler { return &AuthHandler{auth: a} }

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct{ Name, Surname, Email, Password string }
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	u, err := h.auth.Register(req.Name, req.Surname, req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct{ Email, Password string }
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	token, u, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: SessionCookie, Value: token, Path: "/",
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.Logout(extractToken(r)); err != nil {
		writeError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: SessionCookie, Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	u := UserFromContext(r.Context())
	if u == nil {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, u)
}
