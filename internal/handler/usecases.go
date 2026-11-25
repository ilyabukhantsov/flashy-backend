package handler

import "github.com/flashy-team/flashy-backend/internal/domain"

// Вузькі інтерфейси, від яких залежать хендлери (ISP + DIP).
// Їх реалізують сервіси з пакета service.

type AuthUseCase interface {
	Authenticator
	Register(name, surname, email, password string) (*domain.User, error)
	Login(email, password string) (string, *domain.User, error)
	Logout(token string) error
}

type DeckUseCase interface {
	Create(userID, title, description string) (*domain.Deck, error)
	ListMine(userID string) ([]domain.Deck, error)
	Get(userID, deckID string) (*domain.Deck, []domain.Card, error)
	AddCard(userID, deckID, term, translation, example string) (*domain.Card, error)
	Publish(userID, deckID string) (*domain.Deck, error)
	ListCatalog() ([]domain.Deck, error)
}

type StudyUseCase interface {
	BuildQueue(userID, deckID string, limit int) ([]domain.Card, error)
	StartSession(userID, deckID string, mode domain.StudyMode) (*domain.StudySession, error)
	ListSessions(userID string, status domain.SessionStatus) ([]domain.StudySession, error)
	SubmitReview(userID, sessionID, cardID, clientReviewID string, rating domain.Rating) (*domain.ReviewResult, error)
	Finish(userID, sessionID string) (*domain.StudySession, error)
}

type StatsUseCase interface {
	Get(userID string) (*domain.UserStats, error)
}

type NotificationUseCase interface {
	RegisterDevice(userID, fcmToken string, platform domain.Platform) (*domain.Device, error)
}
