package service

import (
	"errors"
	"testing"
	"time"

	"github.com/flashy-team/flashy-backend/internal/domain"
	"github.com/flashy-team/flashy-backend/internal/repository/memory"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time { return c.t }

type fakePush struct {
	sent    int
	invalid map[string]bool
}

func (f *fakePush) Send(token, _, _ string) error {
	if f.invalid[token] {
		return ErrInvalidToken
	}
	f.sent++
	return nil
}

func setup(t *testing.T) (*StudyService, *DeckService, *StatsService, *fakeClock, string, []domain.Card) {
	t.Helper()
	clock := &fakeClock{t: time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)}
	decks := memory.NewDeckRepo()
	deckSvc := NewDeckService(decks, clock)
	statsSvc := NewStatsService(memory.NewStatsRepo(), clock)
	study := NewStudyService(decks, memory.NewProgressRepo(), memory.NewStudySessionRepo(),
		memory.NewReviewLogRepo(), statsSvc, NewSRSScheduler(), clock)

	d, err := deckSvc.Create("u1", "English", "")
	if err != nil {
		t.Fatal(err)
	}
	var cards []domain.Card
	for _, w := range [][2]string{{"cat", "кіт"}, {"dog", "пес"}} {
		c, err := deckSvc.AddCard("u1", d.ID, w[0], w[1], "")
		if err != nil {
			t.Fatal(err)
		}
		cards = append(cards, *c)
	}
	return study, deckSvc, statsSvc, clock, d.ID, cards
}

func TestStudyFlow(t *testing.T) {
	study, _, stats, clock, deckID, cards := setup(t)

	q, _ := study.BuildQueue("u1", deckID, 0)
	if len(q) != 2 {
		t.Fatalf("want 2 new cards in queue, got %d", len(q))
	}
	sess, err := study.StartSession("u1", deckID, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := study.SubmitReview("u1", sess.ID, cards[0].ID, "r1", domain.RatingGood)
	if err != nil || res.IntervalSec != 3600 || res.NewState != domain.CardStateReview {
		t.Fatalf("unexpected result %+v err=%v", res, err)
	}
	// ідемпотентність: той самий clientReviewId не створює дубліката
	again, err := study.SubmitReview("u1", sess.ID, cards[0].ID, "r1", domain.RatingAgain)
	if err != nil || *again != *res {
		t.Fatalf("review must be idempotent: %+v vs %+v (%v)", again, res, err)
	}
	if _, err := study.SubmitReview("u1", sess.ID, cards[1].ID, "r2", domain.RatingAgain); err != nil {
		t.Fatal(err)
	}

	// Усі картки оцінено — черга порожня; через 6 хв знову з'являється картка з Again
	if q, _ := study.BuildQueue("u1", deckID, 0); len(q) != 0 {
		t.Fatalf("queue must be empty, got %d", len(q))
	}
	clock.t = clock.t.Add(6 * time.Minute)
	if q, _ := study.BuildQueue("u1", deckID, 0); len(q) != 1 || q[0].ID != cards[1].ID {
		t.Fatalf("expected the AGAIN card to be due, got %+v", q)
	}

	fin, err := study.Finish("u1", sess.ID)
	if err != nil || fin.ReviewedCount != 2 || fin.CorrectCount != 1 {
		t.Fatalf("unexpected summary %+v err=%v", fin, err)
	}
	if st, _ := stats.Get("u1"); st.CurrentStreak != 1 {
		t.Fatalf("want streak 1, got %d", st.CurrentStreak)
	}
	if _, err := study.SubmitReview("u1", sess.ID, cards[0].ID, "r3", domain.RatingGood); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("review after finish must be conflict, got %v", err)
	}
}

func TestStudyAccessControl(t *testing.T) {
	study, _, _, _, deckID, cards := setup(t)
	if _, err := study.StartSession("intruder", deckID, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("private deck must be 404 for others, got %v", err)
	}
	sess, _ := study.StartSession("u1", deckID, domain.ModeTest)
	if _, err := study.SubmitReview("intruder", sess.ID, cards[0].ID, "", domain.RatingGood); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("foreign session must be 404, got %v", err)
	}
	if _, err := study.SubmitReview("u1", sess.ID, cards[0].ID, "", "BAD"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("bad rating must be validation error, got %v", err)
	}
}

func TestNotifications(t *testing.T) {
	clock := &fakeClock{t: time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)}
	progress, devices := memory.NewProgressRepo(), memory.NewDeviceRepo()
	push := &fakePush{invalid: map[string]bool{"dead": true}}
	n := NewNotificationService(devices, progress, push, clock)

	if _, err := n.RegisterDevice("u1", "good", domain.PlatformAndroid); err != nil {
		t.Fatal(err)
	}
	if _, err := n.RegisterDevice("u1", "dead", domain.PlatformIOS); err != nil {
		t.Fatal(err)
	}
	if sent, _ := n.SendDueReminders(); sent != 0 {
		t.Fatal("no due cards — nothing must be sent")
	}

	p := domain.NewCardProgress("p", "u1", "c1")
	p.ApplyRating(domain.RatingAgain, NewSRSScheduler(), clock.t)
	_ = progress.Save(p)
	clock.t = clock.t.Add(10 * time.Minute)

	sent, _ := n.SendDueReminders()
	if sent != 1 || push.sent != 1 {
		t.Fatalf("want 1 push, got sent=%d push=%d", sent, push.sent)
	}
	if devs, _ := devices.ListByUser("u1"); len(devs) != 1 || devs[0].FCMToken != "good" {
		t.Fatalf("invalid device must be removed, got %+v", devs)
	}
}
