package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kulakoff/event-server-go/internal/storage"
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"github.com/kulakoff/event-server-go/internal/utils"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"
)

// BewardHandler handles messages specific to Beward panels
type BewardHandler struct {
	logger    *slog.Logger
	spamWords []string
	storage   *storage.ClickhouseHttpClient
	fsFiles   *storage.MongoHandler
}

type OpenDoorMsg struct {
	Date   string `json:"date"`
	IP     string `json:"IP"`
	SubId  string `json:"subId"`
	Event  int    `json:"event"`
	Detail string `json:"detail"`
}

// NewBewardHandler creates a new BewardHandler
func NewBewardHandler(logger *slog.Logger, filters []string, storage *storage.ClickhouseHttpClient, mongo *storage.MongoHandler) *BewardHandler {
	return &BewardHandler{
		logger:    logger,
		spamWords: filters,
		storage:   storage,
		fsFiles:   mongo,
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
	storageMessageJson, err := json.Marshal(storageMessage)
	if err != nil {
		h.logger.Warn("Failed to marshal storage message", "error", err)
	}

	// ----- send log to remote storage
	h.storage.Insert("syslog", string(storageMessageJson))

	// --------------------
	// Implement Beward-specific message processing here

	// Track motion detection
	if strings.Contains(message.Message, "SS_MAINAPI_ReportAlarmHappen") {
		h.HandleMotionDetection(&now, host, true)
		/**
		TODO:
			- process motion detect start logic
			- add Prometheus metrics "motion detect start" per host
		*/
	}
	if strings.Contains(message.Message, "SS_MAINAPI_ReportAlarmFinish") {
		h.HandleMotionDetection(&now, host, false)
		/**
		TODO:
			- process motion detect stop logic
			- add Prometheus metrics "motion detect start" per host
		*/
	}

	// Tracks open door
	if strings.Contains(message.Message, "Opening door by code") {
		h.HandleOpenByCode(&now, host, message.Message)
	}
	if strings.Contains(message.Message, "Opening door by RFID") ||
		strings.Contains(message.Message, "Opening door by external RFID") {
		h.HandleOpenByRFID(&now, host, message.Message)
	}

	if strings.Contains(message.Message, "door button pressed") {
		h.HandleOpenByButton(&now, host, message.Message)
	}

	// Tracks calls
}

// FIXME:
// APICallToRBT Update RFID usage timestamp
func (h *BewardHandler) APICallToRBT(payload *OpenDoorMsg) error {
	//url := "http://172.28.0.2/internal/actions/openDoor"
	url := "https://webhook.site/55437bdc-ee94-48d1-b295-22a9f164b610/openDoor"

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, _, err := utils.SendPostRequest(url, headers, payload)
	if err != nil {
		return err
	}

	h.logger.Debug("Successfully sent OpenDoorMsg")
	return nil
}

func (h *BewardHandler) HandleMotionDetection(timestamp *time.Time, host string, motionActive bool) {
	// implement motion detection logic
	// get streamId by intercom IP and call to API FRS. message motion start or stop
	h.logger.Debug("Motion detect process", "host", host, "motionActive", motionActive)
}

func (h *BewardHandler) HandleOpenByCode(timestamp *time.Time, host, message string) {
	// implement open door by code logic
	h.logger.Debug("Open door by code", "host", host, "message", message)
}

func (h *BewardHandler) HandleOpenByRFID(timestamp *time.Time, host, message string) {
	// implement open door by RFID key logic
	h.logger.Debug("Open door by RFID")
	/**
	TODO:
		1 get external reader
		2 get RFID key
		3 get door (main or addition)
		4 update the RFID key last usage date (API call to RBT)
		5 get streamName, streamId
		6 get best quality image from FRS, if exist
		7 get screenshot from media server if FRS not have image
		8 storage image to MongoDB
		9 create "plog" record to clickhouse
	*/

	// ----- 1
	var isExternalReader bool
	if strings.Contains(message, "external") {
		isExternalReader = true
	} else {
		isExternalReader = false
	}

	// ----- 2
	rfidKey := utils.ExtractRFIDKey(message)
	if rfidKey != "" {
		h.logger.Debug("RFID key found", "host", host, "rfid", rfidKey)
	} else {
		h.logger.Warn("RFID key not found", "host", host)
	}

	// ----- 3
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

	// ----- 4
	rbtMessage := OpenDoorMsg{
		Date:   strconv.FormatInt(timestamp.Unix(), 10),
		IP:     host,
		SubId:  "",
		Event:  Event.OpenByCode,
		Detail: rfidKey,
	}
	err := h.APICallToRBT(&rbtMessage)
	if err != nil {
		h.logger.Error("APICallToRBT", "err", err)
	}

	// ----- 5
	// TODO: implement get "streamName" and "streamID" by ip intercom
	//streamName := 8
	streamId := 8                          // FIXME: change fake data
	fakeTimestamp := "2024-10-02 10:44:15" // FIXME: change fake data
	testTimestamp, _ := time.Parse(time.DateTime, fakeTimestamp)

	// ----- 6
	frsResp, err := utils.GetBestQuality(streamId, testTimestamp)
	if err != nil {
		h.logger.Debug("FRS GetBestQuality", "err", err)
	} else if frsResp != nil {
		h.logger.Debug("FRS GetBestQuality OK", "screenshot", frsResp.Data.Screenshot)
	} else {
		h.logger.Debug("FRS GetBestQuality no frame", "err", err)
	}
	// ----- 7 TODO skip
	// ----- 8

	// test | replace hostname in url
	var imageUrl string
	if frsResp != nil {
		imageUrl = frsResp.Data.Screenshot
	}
	imageUrl = strings.Replace(imageUrl, "localhost", "rbt-demo.lanta.me", -1)

	metadata := map[string]interface{}{
		"contentType": "image/jpg",
		"expire":      int32(testTimestamp.Add(time.Hour * 24 * 30 * 6).Unix()),
	}

	// download file from FRS response
	screenShot, err := utils.DownloadFile(imageUrl)
	if err != nil {
		h.logger.Debug("FRS DownloadFile", "err", err)
	}

	// save data to MongoDb
	fileId, err := h.fsFiles.SaveFile("camshot", metadata, screenShot)
	if err != nil {
		h.logger.Debug("MongoDB SaveFile", "err", err)
	}
	h.logger.Debug("MongoDB SaveFile", "fileId", fileId)

	// ----- 9
	imageGUIDv4 := utils.ToGUIDv4(fileId)
	eventGUIDv4 := uuid.New().String()
	flatId := 20 // FIXME: change fake data
	plogData := map[string]interface{}{
		"date":       1727885547,
		"event_uuid": eventGUIDv4,
		"hidden":     0,
		"image_uuid": imageGUIDv4,
		"flat_id":    flatId,
		"domophone":  `{"camera_id": 8, "domophone_description": "✅ Подъезд Beward", "domophone_id": 6, "domophone_output": 0, "entrance_id": 23, "house_id": 11}`,
		"event":      5,
		"opened":     1,
		"face":       `{"faceId": "17", "height": 192, "left": 575, "top": 306, "width": 155}`,
		"rfid":       "",
		"code":       "",
		"phones":     `{"user_phone": ""}`,
		"preview":    2,
	}
	//plogData := map[string]interface{}{
	//	"date":       int32(now.Unix()),
	//	"event_uuid": eventGUIDv4,
	//	"hidden":     0,
	//	"image_id":   imageGUIDv4,
	//	"flat_id":    flatId,
	//	"domophone": map[string]interface{}{
	//		"camera_id":             8,
	//		"domophone_description": "✅ Подъезд Beward",
	//		"domophone_id":          6,
	//		"domophone_output":      0,
	//		"entrance_id":           23,
	//		"house_id":              11,
	//	},
	//	"event":  5,
	//	"opened": 1,
	//	"face": map[string]interface{}{
	//		"faceId": "17",
	//		"height": 204,
	//		"left":   529,
	//		"top":    225,
	//		"width":  160,
	//	},
	//	"rfid":    "",
	//	"code":    "",
	//	"phones":  map[string]interface{}{"user_phone": ""},
	//	"preview": 2,
	//}
	plogDataString, err := json.Marshal(plogData)
	if err != nil {
		h.logger.Debug("Failed marshal JSON")
	}

	fmt.Println(string(plogDataString))
	err = h.storage.Insert("plog", string(plogDataString))
	if err != nil {
		fmt.Println("INSERT ERR", err)
	}
}

func (h *BewardHandler) HandleOpenByButton(timestamp *time.Time, host, message string) {
	// implement open door by open button
	h.logger.Debug("Open door by button", "host", host, "message", message)
	var door int
	var detail string

	door = 0
	detail = "main"

	if strings.Contains(message, "Additional") {
		door = 1
		detail = "second"
	}

	h.logger.Debug("Open door by button", "host", host, "detail", detail, "door", door)
}

func (h *BewardHandler) HandleCallFlow(timestamp *time.Time, host, message string) {
	// implement call flow logic
}
