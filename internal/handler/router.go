// Package handler — шар API Handler: приймає HTTP-запити, валідує їх
// і делегує виконання сервісам бізнес-логіки.
package handler

import "net/http"

// Deps — залежності роутера (інтерфейси, а не конкретні сервіси).
type Deps struct {
	Auth          AuthUseCase
	Decks         DeckUseCase
	Study         StudyUseCase
	Stats         StatsUseCase
	Notifications NotificationUseCase
}

func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()
	protect := AuthMiddleware(d.Auth)
	p := func(f http.HandlerFunc) http.Handler { return protect(f) }

	auth := NewAuthHandler(d.Auth)
	decks := NewDeckHandler(d.Decks)
	study := NewStudyHandler(d.Study)
	stats := NewStatsHandler(d.Stats)
	notif := NewNotificationHandler(d.Notifications)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Публічні
	mux.HandleFunc("POST /auth/register", auth.Register)
	mux.HandleFunc("POST /auth/login", auth.Login)

	// Захищені (cookie flashy_session або Bearer-токен)
	mux.Handle("POST /auth/logout", p(auth.Logout))
	mux.Handle("GET /users/me", p(auth.Me))

	mux.Handle("POST /decks", p(decks.Create))
	mux.Handle("GET /decks", p(decks.List))
	mux.Handle("GET /decks/{id}", p(decks.Get))
	mux.Handle("POST /decks/{id}/cards", p(decks.AddCard))
	mux.Handle("POST /decks/{id}/publish", p(decks.Publish))
	mux.Handle("GET /catalog/decks", p(decks.Catalog))

	mux.Handle("GET /decks/{id}/study/queue", p(study.Queue))
	mux.Handle("POST /study/sessions", p(study.StartSession))
	mux.Handle("GET /study/sessions", p(study.ListSessions))
	mux.Handle("POST /study/sessions/{id}/reviews", p(study.SubmitReview))
	mux.Handle("POST /study/sessions/{id}/finish", p(study.Finish))

	mux.Handle("GET /stats/me", p(stats.Me))
	mux.Handle("POST /devices", p(notif.RegisterDevice))

	return Recover(Logging(mux))
}
