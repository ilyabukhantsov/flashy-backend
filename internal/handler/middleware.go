package handler

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flashy-team/flashy-backend/internal/domain"
)

const SessionCookie = "flashy_session"

type ctxKey struct{}

// Authenticator — абстракція перевірки сесії (реалізує AuthService).
type Authenticator interface {
	Authenticate(token string) (*domain.User, error)
}

// UserFromContext повертає автентифікованого користувача із контексту запиту.
func UserFromContext(ctx context.Context) *domain.User {
	u, _ := ctx.Value(ctxKey{}).(*domain.User)
	return u
}

// extractToken: Bearer-токен (мобільний клієнт) або cookie flashy_session (веб).
func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	if c, err := r.Cookie(SessionCookie); err == nil {
		return c.Value
	}
	return ""
}

// AuthMiddleware перевіряє сесію до передачі запиту хендлеру.
func AuthMiddleware(a Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				writeError(w, domain.ErrUnauthorized)
				return
			}
			u, err := a.Authenticate(token)
			if err != nil {
				writeError(w, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				writeJSON(w, http.StatusInternalServerError,
					errorBody{Code: "internal_error", Message: "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
