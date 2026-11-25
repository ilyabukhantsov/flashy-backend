package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flashy-team/flashy-backend/internal/config"
)

type client struct {
	t     *testing.T
	base  string
	token string
}

func (c *client) do(method, path string, body any) (int, any) {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, c.base+path, &buf)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func field(v any, key string) any { return v.(map[string]any)[key] }

func expect(t *testing.T, got, want int, what string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: want HTTP %d, got %d", what, want, got)
	}
}

func TestEndToEnd(t *testing.T) {
	a := New(config.Config{SessionTTL: time.Hour, ReminderInterval: time.Hour})
	srv := httptest.NewServer(a.Handler)
	defer srv.Close()
	c := &client{t: t, base: srv.URL}

	st, _ := c.do("GET", "/decks", nil)
	expect(t, st, 401, "no token")

	st, _ = c.do("POST", "/auth/register", map[string]string{
		"name": "Ілля", "surname": "Б", "email": "ilya@example.com", "password": "secret123"})
	expect(t, st, 201, "register")
	st, _ = c.do("POST", "/auth/register", map[string]string{
		"name": "Ілля", "email": "ilya@example.com", "password": "secret123"})
	expect(t, st, 409, "duplicate email")
	st, _ = c.do("POST", "/auth/login", map[string]string{"email": "ilya@example.com", "password": "wrong"})
	expect(t, st, 401, "bad password")

	st, body := c.do("POST", "/auth/login", map[string]string{"email": "ilya@example.com", "password": "secret123"})
	expect(t, st, 200, "login")
	c.token = field(body, "token").(string)

	st, body = c.do("POST", "/decks", map[string]string{"title": "English"})
	expect(t, st, 201, "create deck")
	deckID := field(body, "id").(string)

	st, body = c.do("POST", "/decks/"+deckID+"/cards", map[string]string{"term": "cat", "translation": "кіт"})
	expect(t, st, 201, "add card")
	cardID := field(body, "id").(string)

	st, body = c.do("GET", "/decks/"+deckID+"/study/queue", nil)
	expect(t, st, 200, "queue")
	if len(body.([]any)) != 1 {
		t.Fatalf("queue must contain 1 card, got %v", body)
	}

	st, body = c.do("POST", "/study/sessions", map[string]string{"deckId": deckID, "mode": "FLASHCARD"})
	expect(t, st, 201, "start session")
	sessID := field(body, "id").(string)

	review := map[string]string{"cardId": cardID, "rating": "GOOD", "clientReviewId": "x1"}
	st, body = c.do("POST", "/study/sessions/"+sessID+"/reviews", review)
	expect(t, st, 200, "review")
	if field(body, "intervalSec").(float64) != 3600 {
		t.Fatalf("GOOD must give 1h interval, got %v", body)
	}
	st, _ = c.do("POST", "/study/sessions/"+sessID+"/finish", nil)
	expect(t, st, 200, "finish")
	st, _ = c.do("POST", "/study/sessions/"+sessID+"/reviews",
		map[string]string{"cardId": cardID, "rating": "GOOD", "clientReviewId": "x2"})
	expect(t, st, 409, "review after finish")

	st, body = c.do("GET", "/stats/me", nil)
	expect(t, st, 200, "stats")
	if field(body, "currentStreak").(float64) != 1 {
		t.Fatalf("want streak 1, got %v", body)
	}

	st, _ = c.do("POST", "/decks/"+deckID+"/publish", nil)
	expect(t, st, 200, "publish")
	st, body = c.do("GET", "/catalog/decks", nil)
	expect(t, st, 200, "catalog")
	if len(body.([]any)) != 1 {
		t.Fatalf("catalog must contain 1 deck, got %v", body)
	}

	st, _ = c.do("POST", "/devices", map[string]string{"fcmToken": "tok", "platform": "ANDROID"})
	expect(t, st, 201, "register device")
	st, _ = c.do("POST", "/study/sessions", map[string]string{"deckId": deckID, "mode": "NOPE"})
	expect(t, st, 422, "validation")
}
