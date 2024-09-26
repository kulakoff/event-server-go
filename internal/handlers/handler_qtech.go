package handlers

import (
	"github.com/kulakoff/event-server-go/internal/storage"
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
	"strings"
)

// QtechHandler handles messages specific to Beward panels
type QtechHandler struct {
	logger    *slog.Logger
	spamWords []string
	storage   *storage.ClikhouseHandler
}

// NewQtechHandler creates a new QtechHandler
func NewQtechHandler(logger *slog.Logger, filters []string, storage *storage.ClikhouseHandler) *QtechHandler {
	return &QtechHandler{
		logger:    logger,
		spamWords: filters,
		storage:   storage,
	}
}

// HandleMessage processes Beward-specific messages
func (h *QtechHandler) HandleMessage(srcIP string, message *syslog_custom.SyslogMessage) {
	// filter
	if h.FilterMessage(message.Message) {
		// FIXME: remove DEBUG
		h.logger.Debug("Skipping message", "srcIP", srcIP, "host", message.HostName, "message", message.Message)
		return
	}

	h.logger.Info("Processing Qtech message", "srcIP", srcIP, "message", message.Message)
	// Implement Qtech-specific message processing here
}

// FilterMessage skip not informational message
func (h *QtechHandler) FilterMessage(message string) bool {
	for _, word := range h.spamWords {
		if strings.Contains(message, word) {
			return true
		}
	}
	return false
}
