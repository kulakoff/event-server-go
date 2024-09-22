package syslog_custom

import (
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strings"
	"time"
)

type SyslogServer struct {
	port    int
	unit    string // panel type: beward, qtech, ...
	logger  *slog.Logger
	handler MessageHandler
}

type SyslogMessage struct {
	Format         string    `json:"format"`
	Priority       int       `json:"priority"`
	Version        int       `json:"version"`
	Timestamp      time.Time `json:"timestamp"`
	HostName       string    `json:"hostname"`
	AppName        string    `json:"appName"`
	ProcID         string    `json:"procId"`
	StructuredData string    `json:"structuredData"`
	Message        string    `json:"message"`
}

type MessageHandler interface {
	HandleMessage(srcIP string, message *SyslogMessage)
}

func (s *SyslogServer) Start() {
	//syslog_custom port
	addr := fmt.Sprintf(":%d", s.port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		s.logger.Warn("Error resolving UDP address", "error", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		s.logger.Warn("Error starting UDP listener", "error", err)
	}
	defer conn.Close()

	s.logger.Info("Syslog server running", "unit", s.unit, "port", s.port)

	buffer := make([]byte, 1024)
	for {
		n, srcAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			s.logger.Warn("Error reading from UDP: %v", err)
			continue
		}

		message := string(buffer[:n])
		s.logger.Info("HW: %s | Raw Message: %s", s.unit, message)

		parsedMessage, err := s.ParseMessage(message)
		if err != nil {
			s.logger.Warn("Error parsing message", "error", err)
			continue
		}

		if parsedMessage != nil {
			srcIP := srcAddr.IP.String()
			s.handler.HandleMessage(srcIP, parsedMessage)
			s.logger.Info("process-log", "src-ip", srcIP, "msg-ip", parsedMessage.HostName, "msg", parsedMessage.Message)
		}

	}
}

func New(port int, unit string, logger *slog.Logger, handler MessageHandler) *SyslogServer {
	return &SyslogServer{
		port:    port,
		unit:    unit,
		logger:  logger,
		handler: handler,
	}
}

func (s *SyslogServer) ParseMessage(rawMessage string) (*SyslogMessage, error) {
	rfc5424Regex := regexp.MustCompile(`^<(?P<priority>\d|\d{2}|1[1-8]\d|19[01])>(?P<version>\d{1,2})\s(?P<timestamp>-|(?P<fullyear>[12]\d{3})-(?P<month>0\d|[1][012])-(?P<mday>[012]\d|3[01])T(?P<hour>[01]\d|2[0-4]):(?P<minute>[0-5]\d):(?P<second>[0-5]\d|60)(?:\.(?P<secfrac>\d{1,6}))?(?P<numoffset>Z|[+-]\d{2}:\d{2}))\s(?P<hostname>[\S]{1,255})\s(?P<appname>[\S]{1,48})\s(?P<procid>[\S]{1,128})\s(?P<msgid>[\S]{1,32})\s(?P<structureddata>-|(?:\[.+?\]))(?:\s(?P<message>.+))?$`)

	rawMessage = strings.TrimSpace(rawMessage)

	if matches := rfc5424Regex.FindStringSubmatch(rawMessage); matches != nil {
		regexpGroupNames := rfc5424Regex.SubexpNames()

		// TODO : refactor map. test memory usage
		matchMap := make(map[string]string)
		for i, match := range matches {
			if regexpGroupNames[i] != "" {
				matchMap[regexpGroupNames[i]] = match
			}
		}

		hostname := matchMap["hostname"]
		message := matchMap["message"]

		return &SyslogMessage{
			HostName: hostname,
			Message:  message,
		}, nil

	}

	return nil, fmt.Errorf("ParseMessage, unsupported message format: %s", rawMessage)
}
