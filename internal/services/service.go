package services

type IntercomService interface {
	ProcessSyslogMessage(message string) error
}
