package handler

import (
	"net/http"
	"strconv"

	"github.com/flashy-team/flashy-backend/internal/domain"
)

type StudyHandler struct{ study StudyUseCase }

func NewStudyHandler(s StudyUseCase) *StudyHandler { return &StudyHandler{study: s} }

func (h *StudyHandler) Queue(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	cards, err := h.study.BuildQueue(UserFromContext(r.Context()).ID, r.PathValue("id"), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cards)
}

func (h *StudyHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeckID string `json:"deckId"`
		Mode   string `json:"mode"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	s, err := h.study.StartSession(UserFromContext(r.Context()).ID, req.DeckID, domain.StudyMode(req.Mode))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *StudyHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	status := domain.SessionStatus("")
	if r.URL.Query().Get("status") == "active" {
		status = domain.SessionInProgress
	}
	ss, err := h.study.ListSessions(UserFromContext(r.Context()).ID, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ss)
}

func (h *StudyHandler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CardID         string `json:"cardId"`
		Rating         string `json:"rating"`
		ClientReviewID string `json:"clientReviewId"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	res, err := h.study.SubmitReview(UserFromContext(r.Context()).ID, r.PathValue("id"),
		req.CardID, req.ClientReviewID, domain.Rating(req.Rating))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *StudyHandler) Finish(w http.ResponseWriter, r *http.Request) {
	s, err := h.study.Finish(UserFromContext(r.Context()).ID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}
