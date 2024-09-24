package handlers

import (
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
	"strings"
)

// BewardHandler handles messages specific to Beward panels
type BewardHandler struct {
	logger    *slog.Logger
	spamWords []string
}

// NewBewardHandler creates a new BewardHandler
func NewBewardHandler(logger *slog.Logger, filters []string) *BewardHandler {
	return &BewardHandler{
		logger:    logger,
		spamWords: filters,
	}
}

// HandleMessage processes Beward-specific messages
func (h *BewardHandler) HandleMessage(srcIP string, message *syslog_custom.SyslogMessage) {
	h.logger.Info("Processing Beward message", "srcIP", srcIP, "message", message.Message)
	// Implement Beward-specific message processing here
}

// FilterMessage skip not informational message
func (h *BewardHandler) FilterMessage(message string) bool {
	for _, word := range h.spamWords {
		//if strings.Contains(strings.ToLower(message), word) {}
		if strings.Contains(message, word) {
			return true
		}
	}
	return false
}
