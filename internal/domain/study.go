package domain

import "time"

// ---- Перелічення (Рис. 3.1.3) ----

type Rating string

const (
	RatingAgain Rating = "AGAIN"
	RatingHard  Rating = "HARD"
	RatingGood  Rating = "GOOD"
	RatingEasy  Rating = "EASY"
)

func (r Rating) IsValid() bool {
	switch r {
	case RatingAgain, RatingHard, RatingGood, RatingEasy:
		return true
	}
	return false
}

type CardState string

const (
	CardStateNew      CardState = "NEW"
	CardStateLearning CardState = "LEARNING"
	CardStateReview   CardState = "REVIEW"
	CardStateLearned  CardState = "LEARNED"
)

type StudyMode string

const (
	ModeFlashcard StudyMode = "FLASHCARD"
	ModeTest      StudyMode = "TEST"
	ModeWriting   StudyMode = "WRITING"
)

func (m StudyMode) IsValid() bool {
	return m == ModeFlashcard || m == ModeTest || m == ModeWriting
}

type SessionStatus string

const (
	SessionInProgress SessionStatus = "IN_PROGRESS"
	SessionFinished   SessionStatus = "FINISHED"
	SessionAbandoned  SessionStatus = "ABANDONED"
)

// IntervalCalculator — абстракція розрахунку інтервалу повторення (принцип DIP).
// Реалізація (SRSScheduler) живе в шарі service.
type IntervalCalculator interface {
	NextInterval(p *CardProgress, r Rating) time.Duration
}

// ---- CardProgress ----

type CardProgress struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	CardID         string    `json:"cardId"`
	State          CardState `json:"state"`
	IntervalSec    int       `json:"intervalSec"`
	Repetitions    int       `json:"repetitions"`
	Lapses         int       `json:"lapses"`
	DueAt          time.Time `json:"dueAt"`
	LastReviewedAt time.Time `json:"lastReviewedAt"`
}

func NewCardProgress(id, userID, cardID string) *CardProgress {
	return &CardProgress{ID: id, UserID: userID, CardID: cardID, State: CardStateNew}
}

// IsDue — нова картка або така, строк повторення якої настав.
func (p *CardProgress) IsDue(now time.Time) bool {
	return p.State == CardStateNew || !p.DueAt.After(now)
}

func (p *CardProgress) ApplyRating(r Rating, calc IntervalCalculator, now time.Time) {
	interval := calc.NextInterval(p, r)
	p.IntervalSec = int(interval.Seconds())
	p.DueAt = now.Add(interval)
	p.LastReviewedAt = now
	p.Repetitions++
	switch r {
	case RatingAgain:
		p.Lapses++
		p.State = CardStateLearning
	case RatingHard:
		p.State = CardStateLearning
	case RatingGood:
		p.State = CardStateReview
	case RatingEasy:
		p.State = CardStateLearned
	}
}

func (p *CardProgress) Reset() {
	p.State, p.IntervalSec, p.Repetitions, p.Lapses = CardStateNew, 0, 0, 0
	p.DueAt, p.LastReviewedAt = time.Time{}, time.Time{}
}

// ---- StudySession ----

type StudySession struct {
	ID            string        `json:"id"`
	UserID        string        `json:"userId"`
	DeckID        string        `json:"deckId"`
	Mode          StudyMode     `json:"mode"`
	Status        SessionStatus `json:"status"`
	ReviewedCount int           `json:"reviewedCount"`
	CorrectCount  int           `json:"correctCount"`
	StartedAt     time.Time     `json:"startedAt"`
	FinishedAt    *time.Time    `json:"finishedAt,omitempty"`
}

func NewStudySession(id, userID, deckID string, mode StudyMode, now time.Time) (*StudySession, error) {
	if !mode.IsValid() {
		return nil, Invalid("unknown study mode %q", mode)
	}
	return &StudySession{ID: id, UserID: userID, DeckID: deckID, Mode: mode,
		Status: SessionInProgress, StartedAt: now}, nil
}

func (s *StudySession) IsActive() bool { return s.Status == SessionInProgress }

func (s *StudySession) RecordReview(r Rating) {
	s.ReviewedCount++
	if r != RatingAgain {
		s.CorrectCount++
	}
}

func (s *StudySession) Finish(now time.Time) error {
	if !s.IsActive() {
		return ErrConflict
	}
	s.Status = SessionFinished
	t := now
	s.FinishedAt = &t
	return nil
}

// ---- ReviewLog / ReviewResult ----

// ReviewLog — історія одного повторення картки.
type ReviewLog struct {
	ID              string
	SessionID       string
	CardID          string
	ClientReviewID  string
	Rating          Rating
	PrevIntervalSec int
	NewIntervalSec  int
	NewState        CardState
	NextDueAt       time.Time
	ReviewedAt      time.Time
}

// ReviewResult — відповідь клієнту на оцінку картки.
type ReviewResult struct {
	NewState    CardState `json:"newState"`
	IntervalSec int       `json:"intervalSec"`
	NextDueAt   time.Time `json:"nextDueAt"`
}

// ---- UserStats ----

type UserStats struct {
	UserID        string    `json:"userId"`
	CurrentStreak int       `json:"currentStreak"`
	LongestStreak int       `json:"longestStreak"`
	LastStudyDate time.Time `json:"lastStudyDate"`
}

func dayOf(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// UpdateStreak оновлює серію занять: той самий день — без змін,
// наступний день — +1, пропуск — серія починається заново.
func (s *UserStats) UpdateStreak(now time.Time) {
	today := dayOf(now)
	if s.LastStudyDate.IsZero() {
		s.CurrentStreak = 1
	} else {
		diff := int(today.Sub(dayOf(s.LastStudyDate)).Hours() / 24)
		switch {
		case diff <= 0:
			return
		case diff == 1:
			s.CurrentStreak++
		default:
			s.CurrentStreak = 1
		}
	}
	s.LastStudyDate = today
	if s.CurrentStreak > s.LongestStreak {
		s.LongestStreak = s.CurrentStreak
	}
}
