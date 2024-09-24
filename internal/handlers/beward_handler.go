package handlers

import (
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
	"strings"
	"time"
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
	/**
	TODO:
		- add Prometheus metrics per request
		- count motion detect start or stop
		- count open by code
		- count open by button
		- count open by frid key
	*/

	// filter
	if h.FilterMessage(message.Message) {
		// FIXME: remove DEBUG
		h.logger.Debug("Skipping message", "srcIP", srcIP, "host", message.HostName, "message", message.Message)
		return
	}

	h.logger.Info("Processing Beward message", "srcIP", srcIP, "host", message.HostName, "message", message.Message)

	now := time.Now()
	// Implement Beward-specific message processing here

	// Track motion detection
	if strings.Contains(message.Message, "SS_MAINAPI_ReportAlarmHappen") {
		h.logger.Debug("Motion detect start", "srcIP", srcIP, "host", message.HostName, "message", message.Message)
		/**
		TODO:
			- process motion detect start logic
			- add Prometheus metrics "motion detect start" per host
		*/
	}
	if strings.Contains(message.Message, "SS_MAINAPI_ReportAlarmFinish") {
		h.logger.Debug("Motion detect stop", "srcIP", srcIP, "host", message.HostName, "message", message.Message)
		/**
		TODO:
			- process motion detect stop logic
			- add Prometheus metrics "motion detect start" per host
		*/
	}

	// Tracks open door
	if strings.Contains(message.Message, "Opening door by code") {
		h.logger.Debug("Open door by code", "srcIP", srcIP, "host", message.HostName, "message", message.Message)
	}
	if strings.Contains(message.Message, "Opening door by RFID") ||
		strings.Contains(message.Message, "Opening door by external RFID") {
		h.logger.Debug("Open door by RFID", "srcIP", srcIP, "host", message.HostName, "message", message.Message)

	}
	if strings.Contains(message.Message, "door button pressed") {
		h.logger.Debug("Open door by button", "srcIP", srcIP, "host", message.HostName, "message", message.Message)
		var door int
		var detail string

		door = 0
		detail = "main"

		if strings.Contains(message.Message, "Additional") {
			door = 1
			detail = "second"
		}

		h.logger.Debug("Open door by button", "date", now, "ip", message.HostName, "detail", detail, "door", door)
	}

	// Tracks calls
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
