package handlers

import (
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/repository"
	storage2 "github.com/kulakoff/event-server-go/internal/app/event-server-go/storage"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/syslog_custom"
	"log/slog"
	"strings"
)

// QtechHandler handles messages specific to Beward panels
type QtechHandler struct {
	logger    *slog.Logger
	spamWords []string
	storage   *storage2.ClickhouseHttpClient
	fsFiles   *storage2.MongoHandler
	repo      *repository.PostgresRepository
}

// NewQtechHandler creates a new QtechHandler
func NewQtechHandler(logger *slog.Logger, filters []string, storage *storage2.ClickhouseHttpClient, mongo *storage2.MongoHandler, repo *repository.PostgresRepository) *QtechHandler {
	return &QtechHandler{
		logger:    logger,
		spamWords: filters,
		storage:   storage,
		fsFiles:   mongo,
		repo:      repo,
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
