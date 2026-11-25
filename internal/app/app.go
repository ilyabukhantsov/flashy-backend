// Package app — composition root: збирає репозиторії, сервіси та хендлери докупи.
package app

import (
	"log"
	"net/http"

	"github.com/flashy-team/flashy-backend/internal/adapter/push"
	"github.com/flashy-team/flashy-backend/internal/adapter/security"
	"github.com/flashy-team/flashy-backend/internal/config"
	"github.com/flashy-team/flashy-backend/internal/handler"
	"github.com/flashy-team/flashy-backend/internal/repository/memory"
	"github.com/flashy-team/flashy-backend/internal/scheduler"
	"github.com/flashy-team/flashy-backend/internal/service"
)

type App struct {
	Handler   http.Handler
	Reminders *scheduler.Scheduler
}

func New(cfg config.Config) *App {
	clock := service.SystemClock{}

	// Шар доступу до даних (in-memory; PostgreSQL/Redis підключаються тут же)
	users, sessions := memory.NewUserRepo(), memory.NewSessionRepo()
	decks, progress := memory.NewDeckRepo(), memory.NewProgressRepo()
	studySessions, logs := memory.NewStudySessionRepo(), memory.NewReviewLogRepo()
	statsRepo, devices := memory.NewStatsRepo(), memory.NewDeviceRepo()

	// Бізнес-логіка
	authSvc := service.NewAuthService(users, sessions, security.NewSHA256Hasher(), clock, cfg.SessionTTL)
	deckSvc := service.NewDeckService(decks, clock)
	statsSvc := service.NewStatsService(statsRepo, clock)
	studySvc := service.NewStudyService(decks, progress, studySessions, logs, statsSvc,
		service.NewSRSScheduler(), clock)
	notifSvc := service.NewNotificationService(devices, progress, push.NewLogSender(), clock)

	router := handler.NewRouter(handler.Deps{
		Auth: authSvc, Decks: deckSvc, Study: studySvc, Stats: statsSvc, Notifications: notifSvc,
	})

	reminders := scheduler.New(cfg.ReminderInterval, func() {
		n, err := notifSvc.SendDueReminders()
		if err != nil {
			log.Printf("reminders: %v", err)
			return
		}
		log.Printf("reminders sent: %d", n)
	})
	return &App{Handler: router, Reminders: reminders}
}
