package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/storage"
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strconv"
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

// ExtractRFIDKey parse RFID key from message
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
	now := time.Now()

	// filter
	if h.FilterMessage(message.Message) {
		// FIXME: remove DEBUG
		h.logger.Debug("Skipping message", "srcIP", srcIP, "host", message.HostName, "message", message.Message)
		return
	}

	h.logger.Debug("Processing Beward message", "srcIP", srcIP, "host", message.HostName, "message", message.Message)

	// ----- storage message
	var host string
	// use host ip from syslog message
	if net.ParseIP(message.HostName) != nil && message.HostName != "127.0.0.1" && srcIP != message.HostName {
		host = message.HostName
	} else {
		host = srcIP
	}

	storageMessage := storage.SyslogStorageMessage{
		Date:  strconv.FormatInt(time.Now().Unix(), 10),
		Ip:    host,
		SubId: "",
		Unit:  "beward",
		Msg:   message.Message,
	}

	// ----- send log to remote storage
	h.storage.SendLog(storageMessage)

	// --------------------
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
		h.logger.Debug("Open door by RFID")

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

		h.logger.Debug("Open by RFID", "door", door, "rfid", rfidKey)
		/**
		TODO:
			- 1. API call to update RFID usage timestamp
			- 2. get stream name
			- 3. get best quality image from FRS or DVR by stream name
			- 4. save image to mongoDb
			- 5. save plog to clickhouse
		*/

		// 1
		rbtMessage := OpenDoorMsg{
			Date:   strconv.FormatInt(time.Now().Unix(), 10),
			IP:     host,
			SubId:  "",
			Event:  3,
			Detail: rfidKey,
		}
		err := h.APICallToRBT(&rbtMessage)
		if err != nil {
			h.logger.Error("APICallToRBT", "err", err)
		}

		// 2
		// TODO: implement get stream name by ip intercom
		//streamName := 8

		// 3

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

type OpenDoorMsg struct {
	Date   string `json:"date"`
	IP     string `json:"IP"`
	SubId  string `json:"subId"`
	Event  int    `json:"event"`
	Detail string `json:"detail"`
}

func (h *BewardHandler) APICallToRBT(payload *OpenDoorMsg) error {
	//url := "http://172.28.0.2/internal/actions/openDoor"
	url := "https://webhook.site/55437bdc-ee94-48d1-b295-22a9f164b610/openDoor"
	method := "POST"
	fmt.Println(payload)

	// valid payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed marshal payload %w", err)
	}

	// make http request and client
	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	// call request
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", res.Status)
	}

	h.logger.Debug("RFID event sent success")

	return nil
}
