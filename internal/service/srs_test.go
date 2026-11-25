package service

import (
	"testing"
	"time"

	"github.com/flashy-team/flashy-backend/internal/domain"
)

func TestSRSIntervals(t *testing.T) {
	s := NewSRSScheduler()
	want := map[domain.Rating]time.Duration{
		domain.RatingAgain: 5 * time.Minute,
		domain.RatingHard:  15 * time.Minute,
		domain.RatingGood:  time.Hour,
		domain.RatingEasy:  96 * time.Hour,
	}
	for r, d := range want {
		if got := s.NextInterval(nil, r); got != d {
			t.Errorf("%s: want %v, got %v", r, d, got)
		}
	}
}
