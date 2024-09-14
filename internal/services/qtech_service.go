package services

import "log"

type QtechService struct{}

func (b *QtechService) ProcessSyslogMessage(message string) error {
	log.Println("Process QTECH messages: ", message)
	return nil
}
