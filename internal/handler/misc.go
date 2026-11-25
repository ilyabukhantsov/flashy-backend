package handler

import (
	"net/http"

	"github.com/flashy-team/flashy-backend/internal/domain"
)

// StatsHandler — статистика користувача.
type StatsHandler struct{ stats StatsUseCase }

func NewStatsHandler(s StatsUseCase) *StatsHandler { return &StatsHandler{stats: s} }

func (h *StatsHandler) Me(w http.ResponseWriter, r *http.Request) {
	st, err := h.stats.Get(UserFromContext(r.Context()).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// NotificationHandler — реєстрація пристроїв для push.
type NotificationHandler struct{ n NotificationUseCase }

func NewNotificationHandler(n NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{n: n}
}

func (h *NotificationHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FCMToken string `json:"fcmToken"`
		Platform string `json:"platform"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	d, err := h.n.RegisterDevice(UserFromContext(r.Context()).ID, req.FCMToken, domain.Platform(req.Platform))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, d)
}
