package service

import (
	"time"

	"github.com/flashy-team/flashy-backend/internal/domain"
)

// SRSScheduler — розрахунок наступного інтервалу повторення (Рис. 3.1.3).
// Інтервали за специфікацією: Again 5 хв, Hard 15 хв, Good 1 год, Easy 4 дні.
type SRSScheduler struct {
	intervals map[domain.Rating]time.Duration
}

var _ domain.IntervalCalculator = (*SRSScheduler)(nil)

func NewSRSScheduler() *SRSScheduler {
	return &SRSScheduler{intervals: map[domain.Rating]time.Duration{
		domain.RatingAgain: 5 * time.Minute,
		domain.RatingHard:  15 * time.Minute,
		domain.RatingGood:  time.Hour,
		domain.RatingEasy:  4 * 24 * time.Hour,
	}}
}

// NextInterval реалізує domain.IntervalCalculator.
func (s *SRSScheduler) NextInterval(_ *domain.CardProgress, r domain.Rating) time.Duration {
	if d, ok := s.intervals[r]; ok {
		return d
	}
	return s.intervals[domain.RatingAgain]
}

// CalculateNext повертає дату наступного повторення.
func (s *SRSScheduler) CalculateNext(p *domain.CardProgress, r domain.Rating, now time.Time) time.Time {
	return now.Add(s.NextInterval(p, r))
}
