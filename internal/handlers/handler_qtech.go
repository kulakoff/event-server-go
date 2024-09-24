package handlers

import (
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
	"strings"
)

// QtechHandler handles messages specific to Beward panels
type QtechHandler struct {
	logger    *slog.Logger
	spamWords []string
}

// NewQtechHandler creates a new QtechHandler
func NewQtechHandler(logger *slog.Logger, filters []string) *QtechHandler {
	return &QtechHandler{
		logger:    logger,
		spamWords: filters,
	}
}

// HandleMessage processes Beward-specific messages
func (h *QtechHandler) HandleMessage(srcIP string, message *syslog_custom.SyslogMessage) {
	h.logger.Info("Processing Qtech message", "srcIP", srcIP, "message", message.Message)
	// Implement Qtech-specific message processing here
}

// FilterMessage skip not informational message
func (h *QtechHandler) FilterMessage(message string) bool {
	for _, word := range h.spamWords {
		//if strings.Contains(strings.ToLower(message), word) {}
		if strings.Contains(message, word) {
			return true
		}
	}
	return false
}
