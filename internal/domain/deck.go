package domain

import (
	"strings"
	"time"
)

// Deck — колода карток (Рис. 3.1.2).
type Deck struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"ownerId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"isPublic"`
	IsOfficial  bool      `json:"isOfficial"`
	IsHidden    bool      `json:"isHidden"`
	CopiesCount int       `json:"copiesCount"`
	AvgRating   float64   `json:"avgRating"`
	CreatedAt   time.Time `json:"createdAt"`
}

func NewDeck(id, ownerID, title, description string, now time.Time) (*Deck, error) {
	d := &Deck{ID: id, OwnerID: ownerID, Description: description, CreatedAt: now}
	if err := d.Rename(title); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Deck) Rename(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return Invalid("title is required")
	}
	d.Title = title
	return nil
}

// Publish публікує колоду в каталозі. Колоду, приховану модератором, публікувати не можна.
func (d *Deck) Publish() error {
	if d.IsHidden {
		return ErrForbidden
	}
	d.IsPublic = true
	return nil
}

func (d *Deck) Unpublish() { d.IsPublic = false }

// CanBeViewedBy — колоду бачить власник або будь-хто, якщо вона публічна й не прихована.
func (d *Deck) CanBeViewedBy(userID string) bool {
	return d.OwnerID == userID || (d.IsPublic && !d.IsHidden)
}

// Card — одна картка колоди.
type Card struct {
	ID          string    `json:"id"`
	DeckID      string    `json:"deckId"`
	Term        string    `json:"term"`
	Translation string    `json:"translation"`
	Example     string    `json:"example"`
	CreatedAt   time.Time `json:"createdAt"`
}

func NewCard(id, deckID, term, translation, example string, now time.Time) (*Card, error) {
	term, translation = strings.TrimSpace(term), strings.TrimSpace(translation)
	if term == "" || translation == "" {
		return nil, Invalid("term and translation are required")
	}
	return &Card{ID: id, DeckID: deckID, Term: term, Translation: translation, Example: example, CreatedAt: now}, nil
}
