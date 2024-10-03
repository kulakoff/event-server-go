package draft

type IntercomService interface {
	ProcessSyslogMessage(message string) error
}
