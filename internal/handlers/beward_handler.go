package handlers

import (
	"github.com/kulakoff/event-server-go/internal/storage"
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

// BewardHandler handles messages specific to Beward panels
type BewardHandler struct {
	logger    *slog.Logger
	spamWords []string
	storage   *storage.ClikhouseHandler
}

// NewBewardHandler creates a new BewardHandler
func NewBewardHandler(logger *slog.Logger, filters []string, storage *storage.ClikhouseHandler) *BewardHandler {
	return &BewardHandler{
		logger:    logger,
		spamWords: filters,
		storage:   storage,
	}
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

// node examle
//// Opening a door by RFID key
//if (msg.includes("Opening door by RFID") || msg.includes("Opening door by external RFID")) {
//const rfid = msg.match(/\b([0-9A-Fa-f]{14})\b/g)?.[0] || null;
//const isExternalReader = msg.includes('external') || rfid && rfid[6] === '0' && rfid[7] === '0';
//const door = isExternalReader ? 1 : 0;
//await API.openDoor({date: now, ip: host, door, detail: rfid, by: "rfid"});
//}

func (h *BewardHandler) ExtractRFIDKey(message string) string {
	rfidRegex := regexp.MustCompile(`\b([0-9A-Fa-f]{14})\b`)
	match := rfidRegex.FindStringSubmatch(message)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func (h *BewardHandler) APICall() error {
	// Implement API call to RBT
	return nil
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

		// external reader
		var isExternalReader bool
		if strings.Contains(message.Message, "external") {
			isExternalReader = true
		} else {
			isExternalReader = false
		}

		// rfid
		rfidKey := h.ExtractRFIDKey(message.Message)
		if rfidKey != "" {
			h.logger.Debug("RFID key found", "srcIP", srcIP, "host", message.HostName, "rfid", rfidKey)
		} else {
			h.logger.Warn("RFID key not found", "srcIP", srcIP, "host", message.HostName)
		}

		// door
		var door int
		if isExternalReader {
			door = 1
		} else {
			door = 0
		}

		h.logger.Info("Open by RFID", "door", door, "rfid", rfidKey)
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
