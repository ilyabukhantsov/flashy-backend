// Package memory — in-memory реалізація репозиторіїв для розробки та тестів.
// У production-версії замінюється адаптерами PostgreSQL та Redis.
package memory

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flashy-team/flashy-backend/internal/domain"
	"github.com/flashy-team/flashy-backend/internal/pkg/idgen"
)

// ---------- Users ----------

type UserRepo struct {
	mu   sync.RWMutex
	byID map[string]domain.User
}

func NewUserRepo() *UserRepo { return &UserRepo{byID: map[string]domain.User{}} }

func (r *UserRepo) Create(u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range r.byID {
		if strings.EqualFold(e.Email, u.Email) {
			return domain.ErrEmailTaken
		}
	}
	r.byID[u.ID] = *u
	return nil
}

func (r *UserRepo) GetByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.byID {
		if strings.EqualFold(u.Email, email) {
			c := u
			return &c, nil
		}
	}
	return nil, domain.ErrNotFound
}

// ---------- Sessions ----------

type SessionRepo struct {
	mu     sync.RWMutex
	byHash map[string]domain.Session
}

func NewSessionRepo() *SessionRepo { return &SessionRepo{byHash: map[string]domain.Session{}} }

func (r *SessionRepo) Create(s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byHash[s.TokenHash] = *s
	return nil
}

func (r *SessionRepo) GetByTokenHash(hash string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byHash[hash]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &s, nil
}

func (r *SessionRepo) DeleteByTokenHash(hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byHash, hash)
	return nil
}

// ---------- Decks & cards ----------

type DeckRepo struct {
	mu    sync.RWMutex
	decks map[string]domain.Deck
	cards map[string][]domain.Card
}

func NewDeckRepo() *DeckRepo {
	return &DeckRepo{decks: map[string]domain.Deck{}, cards: map[string][]domain.Card{}}
}

func (r *DeckRepo) CreateDeck(d *domain.Deck) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.decks[d.ID] = *d
	return nil
}

func (r *DeckRepo) GetDeck(id string) (*domain.Deck, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.decks[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &d, nil
}

func (r *DeckRepo) filter(pred func(domain.Deck) bool) []domain.Deck {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Deck, 0)
	for _, d := range r.decks {
		if pred(d) {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (r *DeckRepo) ListDecksByOwner(ownerID string) ([]domain.Deck, error) {
	return r.filter(func(d domain.Deck) bool { return d.OwnerID == ownerID }), nil
}

func (r *DeckRepo) ListPublicDecks() ([]domain.Deck, error) {
	return r.filter(func(d domain.Deck) bool { return d.IsPublic && !d.IsHidden }), nil
}

func (r *DeckRepo) UpdateDeck(d *domain.Deck) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.decks[d.ID]; !ok {
		return domain.ErrNotFound
	}
	r.decks[d.ID] = *d
	return nil
}

func (r *DeckRepo) AddCard(c *domain.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.decks[c.DeckID]; !ok {
		return domain.ErrNotFound
	}
	r.cards[c.DeckID] = append(r.cards[c.DeckID], *c)
	return nil
}

func (r *DeckRepo) ListCards(deckID string) ([]domain.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Card, len(r.cards[deckID]))
	copy(out, r.cards[deckID])
	return out, nil
}

// ---------- Card progress ----------

type ProgressRepo struct {
	mu sync.RWMutex
	m  map[string]domain.CardProgress
}

func NewProgressRepo() *ProgressRepo { return &ProgressRepo{m: map[string]domain.CardProgress{}} }

func progressKey(userID, cardID string) string { return userID + "|" + cardID }

func (r *ProgressRepo) Get(userID, cardID string) (*domain.CardProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.m[progressKey(userID, cardID)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &p, nil
}

func (r *ProgressRepo) Save(p *domain.CardProgress) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[progressKey(p.UserID, p.CardID)] = *p
	return nil
}

func (r *ProgressRepo) CountDue(userID string, now time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, p := range r.m {
		if p.UserID == userID && !p.DueAt.After(now) {
			n++
		}
	}
	return n, nil
}

// ---------- Study sessions ----------

type StudySessionRepo struct {
	mu sync.RWMutex
	m  map[string]domain.StudySession
}

func NewStudySessionRepo() *StudySessionRepo {
	return &StudySessionRepo{m: map[string]domain.StudySession{}}
}

func (r *StudySessionRepo) Create(s *domain.StudySession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[s.ID] = *s
	return nil
}

func (r *StudySessionRepo) Get(id string) (*domain.StudySession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.m[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &s, nil
}

func (r *StudySessionRepo) Update(s *domain.StudySession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.m[s.ID] = *s
	return nil
}

func (r *StudySessionRepo) ListByUser(userID string, status domain.SessionStatus) ([]domain.StudySession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.StudySession, 0)
	for _, s := range r.m {
		if s.UserID == userID && (status == "" || s.Status == status) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.Before(out[j].StartedAt) })
	return out, nil
}

// ---------- Review log ----------

type ReviewLogRepo struct {
	mu   sync.RWMutex
	logs []domain.ReviewLog
}

func NewReviewLogRepo() *ReviewLogRepo { return &ReviewLogRepo{} }

func (r *ReviewLogRepo) Save(l *domain.ReviewLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, *l)
	return nil
}

func (r *ReviewLogRepo) FindByClientID(sessionID, clientReviewID string) (*domain.ReviewLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, l := range r.logs {
		if l.SessionID == sessionID && l.ClientReviewID == clientReviewID {
			c := l
			return &c, nil
		}
	}
	return nil, domain.ErrNotFound
}

// ---------- Stats ----------

type StatsRepo struct {
	mu sync.RWMutex
	m  map[string]domain.UserStats
}

func NewStatsRepo() *StatsRepo { return &StatsRepo{m: map[string]domain.UserStats{}} }

func (r *StatsRepo) Get(userID string) (*domain.UserStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.m[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &s, nil
}

func (r *StatsRepo) Save(s *domain.UserStats) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[s.UserID] = *s
	return nil
}

// ---------- Devices ----------

type DeviceRepo struct {
	mu sync.RWMutex
	m  map[string]domain.Device
}

func NewDeviceRepo() *DeviceRepo { return &DeviceRepo{m: map[string]domain.Device{}} }

func (r *DeviceRepo) Upsert(d *domain.Device) (*domain.Device, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, e := range r.m {
		if e.FCMToken == d.FCMToken {
			e.UserID, e.Platform = d.UserID, d.Platform
			r.m[id] = e
			return &e, nil
		}
	}
	if d.ID == "" {
		d.ID = idgen.New()
	}
	r.m[d.ID] = *d
	c := *d
	return &c, nil
}

func (r *DeviceRepo) ListByUser(userID string) ([]domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Device, 0)
	for _, d := range r.m {
		if d.UserID == userID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *DeviceRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.m, id)
	return nil
}

func (r *DeviceRepo) ListUserIDs() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	set := map[string]struct{}{}
	for _, d := range r.m {
		set[d.UserID] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}
