package service

import (
	"errors"

	"github.com/flashy-team/flashy-backend/internal/domain"
	"github.com/flashy-team/flashy-backend/internal/repository"
)

// StatsService — прогрес і серія занять (streak).
type StatsService struct {
	repo  repository.StatsRepository
	clock Clock
}

func NewStatsService(r repository.StatsRepository, c Clock) *StatsService {
	return &StatsService{repo: r, clock: c}
}

func (s *StatsService) Get(userID string) (*domain.UserStats, error) {
	st, err := s.repo.Get(userID)
	if errors.Is(err, domain.ErrNotFound) {
		return &domain.UserStats{UserID: userID}, nil
	}
	return st, err
}

// RecordStudy фіксує факт заняття сьогодні та оновлює серію.
func (s *StatsService) RecordStudy(userID string) error {
	st, err := s.Get(userID)
	if err != nil {
		return err
	}
	st.UpdateStreak(s.clock.Now())
	return s.repo.Save(st)
}
