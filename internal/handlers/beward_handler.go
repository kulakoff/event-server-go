package handlers

import (
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
)

// BewardHandler handles messages specific to Beward panels
type BewardHandler struct {
	logger *slog.Logger
}

// NewBewardHandler creates a new BewardHandler
func NewBewardHandler(logger *slog.Logger) *BewardHandler {
	return &BewardHandler{logger: logger}
}

// HandleMessage processes Beward-specific messages
func (h *BewardHandler) HandleMessage(srcIP string, message *syslog_custom.SyslogMessage) {
	h.logger.Info("Processing Beward message", "srcIP", srcIP, "message", message)
	// Implement Beward-specific message processing here
}
