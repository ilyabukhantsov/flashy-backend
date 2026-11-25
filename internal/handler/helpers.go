package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/flashy-team/flashy-backend/internal/domain"
)

// errorBody — єдиний формат помилок { code, message } (специфікація API).
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError зіставляє доменні помилки зі стандартними HTTP-кодами.
func writeError(w http.ResponseWriter, err error) {
	status, code := http.StatusInternalServerError, "internal_error"
	msg := "internal server error"
	switch {
	case errors.Is(err, domain.ErrValidation):
		status, code, msg = http.StatusUnprocessableEntity, "validation_error", err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		status, code, msg = http.StatusUnauthorized, "unauthorized", "authentication required"
	case errors.Is(err, domain.ErrForbidden):
		status, code, msg = http.StatusForbidden, "forbidden", "insufficient permissions"
	case errors.Is(err, domain.ErrNotFound):
		status, code, msg = http.StatusNotFound, "not_found", "resource not found"
	case errors.Is(err, domain.ErrConflict):
		status, code, msg = http.StatusConflict, "conflict", err.Error()
	default:
		log.Printf("internal error: %v", err)
	}
	writeJSON(w, status, errorBody{Code: code, Message: msg})
}

func decode(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return domain.Invalid("invalid JSON body")
	}
	return nil
}
