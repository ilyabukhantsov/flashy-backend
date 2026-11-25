// Package push містить адаптери каналів push-сповіщень.
package push

import "log"

// LogSender — заглушка FCMPushSender: лише пише повідомлення в лог.
// Справжній адаптер Firebase Cloud Messaging реалізує той самий інтерфейс service.PushSender.
type LogSender struct{}

func NewLogSender() *LogSender { return &LogSender{} }

func (LogSender) Send(fcmToken, title, body string) error {
	log.Printf("[push-stub] token=%s title=%q body=%q", fcmToken, title, body)
	return nil
}
