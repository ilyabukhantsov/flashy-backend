package service

import (
	"github.com/flashy-team/flashy-backend/internal/domain"
	"github.com/flashy-team/flashy-backend/internal/pkg/idgen"
	"github.com/flashy-team/flashy-backend/internal/repository"
)

// DeckService — колоди, картки та каталог публічних колод.
type DeckService struct {
	decks repository.DeckRepository
	clock Clock
}

func NewDeckService(d repository.DeckRepository, c Clock) *DeckService {
	return &DeckService{decks: d, clock: c}
}

func (s *DeckService) Create(userID, title, description string) (*domain.Deck, error) {
	d, err := domain.NewDeck(idgen.New(), userID, title, description, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.decks.CreateDeck(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DeckService) ListMine(userID string) ([]domain.Deck, error) {
	return s.decks.ListDecksByOwner(userID)
}

// Get повертає колоду з картками. Чужу непублічну колоду приховуємо як 404.
func (s *DeckService) Get(userID, deckID string) (*domain.Deck, []domain.Card, error) {
	d, err := s.decks.GetDeck(deckID)
	if err != nil || !d.CanBeViewedBy(userID) {
		return nil, nil, domain.ErrNotFound
	}
	cards, err := s.decks.ListCards(deckID)
	if err != nil {
		return nil, nil, err
	}
	return d, cards, nil
}

func (s *DeckService) ownedDeck(userID, deckID string) (*domain.Deck, error) {
	d, err := s.decks.GetDeck(deckID)
	if err != nil {
		return nil, err
	}
	if d.OwnerID != userID {
		return nil, domain.ErrForbidden
	}
	return d, nil
}

func (s *DeckService) AddCard(userID, deckID, term, translation, example string) (*domain.Card, error) {
	if _, err := s.ownedDeck(userID, deckID); err != nil {
		return nil, err
	}
	c, err := domain.NewCard(idgen.New(), deckID, term, translation, example, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.decks.AddCard(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Publish публікує колоду в каталозі (потрібна хоча б одна картка).
func (s *DeckService) Publish(userID, deckID string) (*domain.Deck, error) {
	d, err := s.ownedDeck(userID, deckID)
	if err != nil {
		return nil, err
	}
	cards, err := s.decks.ListCards(deckID)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, domain.Invalid("deck must contain at least one card")
	}
	if err := d.Publish(); err != nil {
		return nil, err
	}
	if err := s.decks.UpdateDeck(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DeckService) ListCatalog() ([]domain.Deck, error) {
	return s.decks.ListPublicDecks()
}
