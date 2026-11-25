// Package repository містить абстракції доступу до даних (шар доступу до даних).
// Сервіси залежать лише від цих інтерфейсів (принцип DIP), тому in-memory
// реалізацію легко замінити на PostgreSQL / Redis без змін бізнес-логіки.
package repository

import (
	"time"

	"github.com/flashy-team/flashy-backend/internal/domain"
)

// Усі методи Get* повертають domain.ErrNotFound, якщо запису немає.

type UserRepository interface {
	Create(u *domain.User) error // domain.ErrEmailTaken, якщо email зайнято
	GetByID(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
}

type SessionRepository interface {
	Create(s *domain.Session) error
	GetByTokenHash(hash string) (*domain.Session, error)
	DeleteByTokenHash(hash string) error
}

type DeckRepository interface {
	CreateDeck(d *domain.Deck) error
	GetDeck(id string) (*domain.Deck, error)
	ListDecksByOwner(ownerID string) ([]domain.Deck, error)
	ListPublicDecks() ([]domain.Deck, error)
	UpdateDeck(d *domain.Deck) error
	AddCard(c *domain.Card) error
	ListCards(deckID string) ([]domain.Card, error)
}

type ProgressRepository interface {
	Get(userID, cardID string) (*domain.CardProgress, error)
	Save(p *domain.CardProgress) error
	CountDue(userID string, now time.Time) (int, error)
}

type StudySessionRepository interface {
	Create(s *domain.StudySession) error
	Get(id string) (*domain.StudySession, error)
	Update(s *domain.StudySession) error
	// ListByUser повертає сесії користувача; порожній status — усі.
	ListByUser(userID string, status domain.SessionStatus) ([]domain.StudySession, error)
}

type ReviewLogRepository interface {
	Save(l *domain.ReviewLog) error
	FindByClientID(sessionID, clientReviewID string) (*domain.ReviewLog, error)
}

type StatsRepository interface {
	Get(userID string) (*domain.UserStats, error)
	Save(s *domain.UserStats) error
}

type DeviceRepository interface {
	Upsert(d *domain.Device) (*domain.Device, error) // за FCM-токеном
	ListByUser(userID string) ([]domain.Device, error)
	Delete(id string) error
	ListUserIDs() ([]string, error)
}
