package services

import (
	"log/slog"
)

type QtechService struct{}

func (b *QtechService) ProcessSyslogMessage(message string) error {
	slog.Info("Process QTECH messages: ", message)
	return nil
}
