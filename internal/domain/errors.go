package domain

import (
	"errors"
	"fmt"
)

// Доменні помилки. Шар handler зіставляє їх із HTTP-кодами (404, 409, 422, 401, 403).
var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrEmailTaken   = fmt.Errorf("%w: email already registered", ErrConflict)
)

// Invalid створює помилку валідації з описом.
func Invalid(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrValidation, fmt.Sprintf(format, a...))
}
