package syslog_custom

import (
	"fmt"
	"github.com/leodido/go-syslog/rfc5424"
	"log"
	"log/slog"
	"net"
)

type SyslogServer struct {
	port    int
	unit    string // panel type: beward, qtech, ...
	logger  *slog.Logger
	handler MessageHandler
}

type MessageHandler interface {
	HandleMessage(srcIP, message string)
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
	parser := rfc5424.NewParser()
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		message := string(buffer[:n])
		log.Printf("HW: %s | Raw Message: %s", s.unit, message)

		// Парсинг syslog сообщения
		syslogMessage, err := parser.Parse(buffer[:n], nil)
		if err != nil {
			log.Printf("Error parsing syslog message: %v", err)
			continue
		}

		log.Printf("Parsed Syslog Message: %+v", syslogMessage.Message())

		//s.handler.HandleMessage(srcIP, message)
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
