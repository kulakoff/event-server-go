package handlers

import (
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
)

// QtechHandler handles messages specific to Beward panels
type QtechHandler struct {
	logger *slog.Logger
}

// NewQtechHandler creates a new QtechHandler
func NewQtechHandler(logger *slog.Logger) *QtechHandler {
	return &QtechHandler{logger: logger}
}

// HandleMessage processes Beward-specific messages
func (h *QtechHandler) HandleMessage(srcIP string, message *syslog_custom.SyslogMessage) {
	h.logger.Info("Processing Qtech message", "srcIP", srcIP, "message", message.Message)
	// Implement Qtech-specific message processing here
}
