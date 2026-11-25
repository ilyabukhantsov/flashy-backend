package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/flashy-team/flashy-backend/internal/domain"
	"github.com/flashy-team/flashy-backend/internal/pkg/idgen"
	"github.com/flashy-team/flashy-backend/internal/repository"
)

// ErrInvalidToken повертає PushSender, якщо токен пристрою недійсний (UNREGISTERED).
var ErrInvalidToken = errors.New("invalid push token")

// PushSender — абстракція зовнішнього каналу push (Firebase Cloud Messaging).
type PushSender interface {
	Send(fcmToken, title, body string) error
}

// NotificationService — реєстрація пристроїв і розсилка нагадувань (Діаграма 3.2.3).
type NotificationService struct {
	devices  repository.DeviceRepository
	progress repository.ProgressRepository
	push     PushSender
	clock    Clock
}

func NewNotificationService(d repository.DeviceRepository, p repository.ProgressRepository,
	push PushSender, c Clock) *NotificationService {
	return &NotificationService{devices: d, progress: p, push: push, clock: c}
}

func (s *NotificationService) RegisterDevice(userID, fcmToken string, platform domain.Platform) (*domain.Device, error) {
	if fcmToken == "" {
		return nil, domain.Invalid("fcmToken is required")
	}
	if !platform.IsValid() {
		return nil, domain.Invalid("unknown platform %q", platform)
	}
	return s.devices.Upsert(&domain.Device{ID: idgen.New(), UserID: userID, FCMToken: fcmToken,
		Platform: platform, RegisteredAt: s.clock.Now()})
}

// SendDueReminders надсилає нагадування користувачам, у яких є картки до повторення.
// Недійсний токен пристрою видаляється, розсилка для інших триває.
func (s *NotificationService) SendDueReminders() (int, error) {
	ids, err := s.devices.ListUserIDs()
	if err != nil {
		return 0, err
	}
	now, sent := s.clock.Now(), 0
	for _, uid := range ids {
		due, err := s.progress.CountDue(uid, now)
		if err != nil || due == 0 {
			continue
		}
		devs, err := s.devices.ListByUser(uid)
		if err != nil {
			continue
		}
		for _, d := range devs {
			err := s.push.Send(d.FCMToken, "Flashy", pluralCards(due))
			switch {
			case errors.Is(err, ErrInvalidToken):
				_ = s.devices.Delete(d.ID)
			case err != nil:
				log.Printf("push to device %s failed: %v", d.ID, err)
			default:
				sent++
			}
		}
	}
	return sent, nil
}

func pluralCards(n int) string {
	return fmt.Sprintf("У вас %d карток до повторення", n)
}
