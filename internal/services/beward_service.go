package services

import (
	"log/slog"
)

type BewardService struct{}

func (b *BewardService) ProcessSyslogMessage(message string) error {
	slog.Info("Process BEWARD messages: ", message)
	return nil
}
