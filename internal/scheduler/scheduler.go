// Package scheduler — фоновий планувальник (горутина всередині того ж процесу).
package scheduler

import (
	"context"
	"time"
)

type Scheduler struct {
	interval time.Duration
	job      func()
}

func New(interval time.Duration, job func()) *Scheduler {
	return &Scheduler{interval: interval, job: job}
}

// Run виконує job з заданим інтервалом, доки контекст не буде скасовано.
// TODO: запуск у dailyReminderTime кожного користувача.
func (s *Scheduler) Run(ctx context.Context) {
	t := time.NewTicker(s.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.job()
		}
	}
}
