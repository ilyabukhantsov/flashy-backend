package handler

import "net/http"

type DeckHandler struct{ decks DeckUseCase }

func NewDeckHandler(d DeckUseCase) *DeckHandler { return &DeckHandler{decks: d} }

func (h *DeckHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct{ Title, Description string }
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	d, err := h.decks.Create(UserFromContext(r.Context()).ID, req.Title, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

func (h *DeckHandler) List(w http.ResponseWriter, r *http.Request) {
	ds, err := h.decks.ListMine(UserFromContext(r.Context()).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *DeckHandler) Get(w http.ResponseWriter, r *http.Request) {
	d, cards, err := h.decks.Get(UserFromContext(r.Context()).ID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deck": d, "cards": cards})
}

func (h *DeckHandler) AddCard(w http.ResponseWriter, r *http.Request) {
	var req struct{ Term, Translation, Example string }
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	c, err := h.decks.AddCard(UserFromContext(r.Context()).ID, r.PathValue("id"),
		req.Term, req.Translation, req.Example)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *DeckHandler) Publish(w http.ResponseWriter, r *http.Request) {
	d, err := h.decks.Publish(UserFromContext(r.Context()).ID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// Catalog — CatalogHandler із діаграми: перелік публічних колод.
func (h *DeckHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	ds, err := h.decks.ListCatalog()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ds)
}
