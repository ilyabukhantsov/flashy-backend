package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/flashy-team/flashy-backend/internal/domain"
	"github.com/flashy-team/flashy-backend/internal/pkg/idgen"
	"github.com/flashy-team/flashy-backend/internal/repository"
)

// PasswordHasher — абстракція хешування паролів (DIP).
type PasswordHasher interface {
	Hash(raw string) (string, error)
	Verify(raw, hash string) bool
}

// AuthService — реєстрація, вхід і перевірка сесій.
type AuthService struct {
	users    repository.UserRepository
	sessions repository.SessionRepository
	hasher   PasswordHasher
	clock    Clock
	ttl      time.Duration
}

func NewAuthService(u repository.UserRepository, s repository.SessionRepository,
	h PasswordHasher, c Clock, sessionTTL time.Duration) *AuthService {
	return &AuthService{users: u, sessions: s, hasher: h, clock: c, ttl: sessionTTL}
}

func (s *AuthService) Register(name, surname, email, password string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") {
		return nil, domain.Invalid("invalid email")
	}
	if len(password) < 8 {
		return nil, domain.Invalid("password must be at least 8 characters")
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		ID: idgen.New(), Email: email, PasswordHash: hash,
		Status: domain.AccountActive, Role: domain.RoleUser,
		InterfaceLang: "uk", Theme: domain.ThemeLight, CreatedAt: s.clock.Now(),
	}
	if err := u.UpdateProfile(name, surname); err != nil {
		return nil, err
	}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login повертає відкритий токен сесії (у БД зберігається лише його хеш).
func (s *AuthService) Login(email, password string) (string, *domain.User, error) {
	u, err := s.users.GetByEmail(strings.TrimSpace(email))
	if err != nil || !s.hasher.Verify(password, u.PasswordHash) {
		return "", nil, domain.ErrUnauthorized
	}
	if u.IsBlocked() {
		return "", nil, domain.ErrForbidden
	}
	token := idgen.Token()
	now := s.clock.Now()
	sess := &domain.Session{ID: idgen.New(), UserID: u.ID, TokenHash: hashToken(token),
		CreatedAt: now, ExpiresAt: now.Add(s.ttl)}
	if err := s.sessions.Create(sess); err != nil {
		return "", nil, err
	}
	return token, u, nil
}

// Authenticate перевіряє токен сесії та повертає користувача.
func (s *AuthService) Authenticate(token string) (*domain.User, error) {
	sess, err := s.sessions.GetByTokenHash(hashToken(token))
	if err != nil || sess.IsExpired(s.clock.Now()) {
		return nil, domain.ErrUnauthorized
	}
	u, err := s.users.GetByID(sess.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if u.IsBlocked() {
		return nil, domain.ErrForbidden
	}
	return u, nil
}

func (s *AuthService) Logout(token string) error {
	return s.sessions.DeleteByTokenHash(hashToken(token))
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
