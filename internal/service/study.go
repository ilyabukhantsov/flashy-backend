package service

import (
	"errors"
	"fmt"

	"github.com/flashy-team/flashy-backend/internal/domain"
	"github.com/flashy-team/flashy-backend/internal/pkg/idgen"
	"github.com/flashy-team/flashy-backend/internal/repository"
)

// DefaultQueueLimit — максимальний розмір черги карток за один запит.
const DefaultQueueLimit = 50

// StudyService — центральний сервіс: черга карток і сесія тренування (Рис. 3.1.3).
type StudyService struct {
	decks    repository.DeckRepository
	progress repository.ProgressRepository
	sessions repository.StudySessionRepository
	logs     repository.ReviewLogRepository
	stats    *StatsService
	calc     domain.IntervalCalculator
	clock    Clock
}

func NewStudyService(d repository.DeckRepository, p repository.ProgressRepository,
	s repository.StudySessionRepository, l repository.ReviewLogRepository,
	st *StatsService, calc domain.IntervalCalculator, c Clock) *StudyService {
	return &StudyService{decks: d, progress: p, sessions: s, logs: l, stats: st, calc: calc, clock: c}
}

func (s *StudyService) accessibleDeck(userID, deckID string) (*domain.Deck, error) {
	d, err := s.decks.GetDeck(deckID)
	if err != nil || !d.CanBeViewedBy(userID) {
		return nil, domain.ErrNotFound
	}
	return d, nil
}

// BuildQueue повертає нові картки та картки, яким настав час повторення.
func (s *StudyService) BuildQueue(userID, deckID string, limit int) ([]domain.Card, error) {
	if _, err := s.accessibleDeck(userID, deckID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > DefaultQueueLimit {
		limit = DefaultQueueLimit
	}
	cards, err := s.decks.ListCards(deckID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	queue := make([]domain.Card, 0)
	for _, c := range cards {
		if len(queue) >= limit {
			break
		}
		p, err := s.progress.Get(userID, c.ID)
		if errors.Is(err, domain.ErrNotFound) || (err == nil && p.IsDue(now)) {
			queue = append(queue, c)
		} else if err != nil {
			return nil, err
		}
	}
	return queue, nil
}

func (s *StudyService) StartSession(userID, deckID string, mode domain.StudyMode) (*domain.StudySession, error) {
	if mode == "" {
		mode = domain.ModeFlashcard
	}
	if _, err := s.accessibleDeck(userID, deckID); err != nil {
		return nil, err
	}
	sess, err := domain.NewStudySession(idgen.New(), userID, deckID, mode, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.sessions.Create(sess); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *StudyService) ListSessions(userID string, status domain.SessionStatus) ([]domain.StudySession, error) {
	return s.sessions.ListByUser(userID, status)
}

func (s *StudyService) ownedSession(userID, sessionID string) (*domain.StudySession, error) {
	sess, err := s.sessions.Get(sessionID)
	if err != nil || sess.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return sess, nil
}

// SubmitReview застосовує оцінку картки. Операція ідемпотентна за clientReviewID:
// повторна відправка після обриву зв'язку повертає попередній результат.
func (s *StudyService) SubmitReview(userID, sessionID, cardID, clientReviewID string,
	rating domain.Rating) (*domain.ReviewResult, error) {
	if !rating.IsValid() {
		return nil, domain.Invalid("unknown rating %q", rating)
	}
	sess, err := s.ownedSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	if clientReviewID != "" {
		if l, err := s.logs.FindByClientID(sessionID, clientReviewID); err == nil {
			return &domain.ReviewResult{NewState: l.NewState, IntervalSec: l.NewIntervalSec, NextDueAt: l.NextDueAt}, nil
		}
	}
	if !sess.IsActive() {
		return nil, fmt.Errorf("%w: session already finished", domain.ErrConflict)
	}
	if err := s.ensureCardInDeck(sess.DeckID, cardID); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	p, err := s.progress.Get(userID, cardID)
	if errors.Is(err, domain.ErrNotFound) {
		p = domain.NewCardProgress(idgen.New(), userID, cardID)
	} else if err != nil {
		return nil, err
	}
	prev := p.IntervalSec
	p.ApplyRating(rating, s.calc, now)
	if err := s.progress.Save(p); err != nil {
		return nil, err
	}
	log := &domain.ReviewLog{ID: idgen.New(), SessionID: sessionID, CardID: cardID,
		ClientReviewID: clientReviewID, Rating: rating, PrevIntervalSec: prev,
		NewIntervalSec: p.IntervalSec, NewState: p.State, NextDueAt: p.DueAt, ReviewedAt: now}
	if err := s.logs.Save(log); err != nil {
		return nil, err
	}
	sess.RecordReview(rating)
	if err := s.sessions.Update(sess); err != nil {
		return nil, err
	}
	return &domain.ReviewResult{NewState: p.State, IntervalSec: p.IntervalSec, NextDueAt: p.DueAt}, nil
}

func (s *StudyService) ensureCardInDeck(deckID, cardID string) error {
	cards, err := s.decks.ListCards(deckID)
	if err != nil {
		return err
	}
	for _, c := range cards {
		if c.ID == cardID {
			return nil
		}
	}
	return domain.ErrNotFound
}

// Finish закриває сесію та оновлює серію занять.
func (s *StudyService) Finish(userID, sessionID string) (*domain.StudySession, error) {
	sess, err := s.ownedSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	if err := sess.Finish(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("%w: session already finished", domain.ErrConflict)
	}
	if err := s.sessions.Update(sess); err != nil {
		return nil, err
	}
	if err := s.stats.RecordStudy(userID); err != nil {
		return nil, err
	}
	return sess, nil
}
