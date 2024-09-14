package services

import "log"

type BewardService struct{}

func (b *BewardService) ProcessSyslogMessage(message string) error {
	log.Println("Process BEWARD messages: ", message)
	return nil
}
