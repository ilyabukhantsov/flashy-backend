package domain

import (
	"testing"
	"time"
)

type fixedCalc struct{}

func (fixedCalc) NextInterval(_ *CardProgress, _ Rating) time.Duration { return time.Hour }

func TestApplyRating(t *testing.T) {
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	p := NewCardProgress("p1", "u1", "c1")
	if !p.IsDue(now) {
		t.Fatal("new card must be due")
	}
	p.ApplyRating(RatingAgain, fixedCalc{}, now)
	if p.State != CardStateLearning || p.Lapses != 1 || p.Repetitions != 1 {
		t.Fatalf("unexpected progress: %+v", p)
	}
	if p.IsDue(now) || !p.IsDue(now.Add(time.Hour)) {
		t.Fatal("due date must be now+1h")
	}
	p.ApplyRating(RatingEasy, fixedCalc{}, now)
	if p.State != CardStateLearned {
		t.Fatalf("want LEARNED, got %s", p.State)
	}
	p.Reset()
	if p.State != CardStateNew || p.Repetitions != 0 {
		t.Fatal("reset failed")
	}
}

func TestUpdateStreak(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 1, d, 12, 0, 0, 0, time.UTC) }
	s := &UserStats{UserID: "u1"}
	s.UpdateStreak(day(1))
	s.UpdateStreak(day(1)) // той самий день
	if s.CurrentStreak != 1 {
		t.Fatalf("want 1, got %d", s.CurrentStreak)
	}
	s.UpdateStreak(day(2))
	s.UpdateStreak(day(3))
	if s.CurrentStreak != 3 || s.LongestStreak != 3 {
		t.Fatalf("want streak 3, got %+v", s)
	}
	s.UpdateStreak(day(6)) // пропуск
	if s.CurrentStreak != 1 || s.LongestStreak != 3 {
		t.Fatalf("streak must reset, got %+v", s)
	}
}

func TestSessionFinishTwice(t *testing.T) {
	s, err := NewStudySession("s", "u", "d", ModeFlashcard, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Finish(time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := s.Finish(time.Now()); err == nil {
		t.Fatal("second Finish must fail")
	}
	if _, err := NewStudySession("s", "u", "d", "BAD", time.Now()); err == nil {
		t.Fatal("invalid mode must fail")
	}
}

func TestDeckRules(t *testing.T) {
	if _, err := NewDeck("1", "u", "  ", "", time.Now()); err == nil {
		t.Fatal("empty title must fail")
	}
	d, _ := NewDeck("1", "owner", "Words", "", time.Now())
	if d.CanBeViewedBy("other") {
		t.Fatal("private deck must be hidden from others")
	}
	_ = d.Publish()
	if !d.CanBeViewedBy("other") {
		t.Fatal("public deck must be visible")
	}
	d.IsHidden = true
	if d.CanBeViewedBy("other") || d.Publish() == nil {
		t.Fatal("hidden deck rules violated")
	}
}
