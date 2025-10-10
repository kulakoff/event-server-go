package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kulakoff/event-server-go/internal/app/event-server-go/repository/models"

	"github.com/kulakoff/event-server-go/internal/app/event-server-go/repository"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/services/frs"
	storage2 "github.com/kulakoff/event-server-go/internal/app/event-server-go/storage"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/syslog_custom"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/utils"

	"github.com/google/uuid"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/config"
)

const (
	PREVIEW_IPCAM     = 1
	PREVIEW_FRS       = 2
	TTL_CAMSHOT_HOURS = time.Hour * 24 * 30 * 6

	CALL_TYPE_SIP = "sip"
	CALL_TYPE_CMS = "cms"

	// event types
	EVENT_UNANSWERED_CALL      = 1
	EVENT_ANSWERED_CALL        = 2
	EVENT_OPENED_BY_KEY        = 3
	EVENT_OPENED_BY_APP        = 4
	EVENT_OPENED_BY_FACE       = 5
	EVENT_OPENED_BY_CODE       = 6
	EVENT_OPENED_GATES_BY_CALL = 7
	EVENT_OPENED_BY_VEHICLE    = 9
)

type CallData struct {
	CallID      int
	Apartment   int
	DomophoneIP string
	StartTime   *time.Time
	EndTime     *time.Time
	Answered    bool
	DoorOpened  bool
	CallType    string

	// Data for event
	CameraID  int
	CameraFRS string
	Domophone *models.Domophone
	Entrance  *models.HouseEntrance
	FlatID    int

	// Images data
	ScreenshotData []byte
	FaceData       map[string]interface{}
	PreviewType    int

	// Screenshot storage info
	screenshotFileID string
	screenshotsReady bool
	callMutex        sync.Mutex
}

// BewardHandler handles messages specific to Beward panels
type BewardHandler struct {
	logger      *slog.Logger
	spamWords   []string
	storage     *storage2.ClickhouseHttpClient
	fsFiles     *storage2.MongoHandler
	repo        *repository.PostgresRepository
	rbtApi      *config.RbtApi
	frsApi      *config.FrsApi
	activeCalls map[int]*CallData // key: beward callId
	callMutex   sync.Mutex
}

type OpenDoorMsg struct {
	Date   string `json:"date"`
	IP     string `json:"IP"`
	SubId  string `json:"subId"`
	Event  int    `json:"event"`
	Detail string `json:"detail"`
}

// NewBewardHandler creates a new BewardHandler
func NewBewardHandler(
	logger *slog.Logger,
	filters []string,
	storage *storage2.ClickhouseHttpClient,
	mongo *storage2.MongoHandler,
	repo *repository.PostgresRepository,
	rbtApi *config.RbtApi,
	frsApi *config.FrsApi,
) *BewardHandler {
	return &BewardHandler{
		logger:      logger,
		spamWords:   filters,
		storage:     storage,
		fsFiles:     mongo,
		repo:        repo,
		rbtApi:      rbtApi,
		frsApi:      frsApi,
		activeCalls: make(map[int]*CallData),
	}
}

// FilterMessage skip not informational syslog message
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

	// 1 ----- make event timestamp
	// FIXME: load location from system or config
	location, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(location).Truncate(time.Second)

	// 2 ----- filter message
	if h.FilterMessage(message.Message) {
		// FIXME: remove DEBUG
		//h.logger.Debug("HandleMessage || Skipping message", "srcIP", srcIP, "host", message.HostName, "message", message.Message)
		return
	}

	h.logger.Debug("HandleMessage || Processing Beward message", "ip", srcIP, "host", message.HostName, "message", message.Message)

	// 3 ----- storage message
	var host string
	// use host ip from syslog message
	if net.ParseIP(message.HostName) != nil && message.HostName != "127.0.0.1" && srcIP != message.HostName {
		host = message.HostName
	} else {
		host = srcIP
	}

	storageMessage := storage2.SyslogStorageMessage{
		Date:  strconv.FormatInt(time.Now().Unix(), 10),
		Ip:    host,
		SubId: "",
		Unit:  "beward",
		Msg:   message.Message,
	}

	// convert JSON to string
	storageMessageJson, err := json.Marshal(storageMessage)
	if err != nil {
		h.logger.Warn("Failed to marshal storage message", "error", err)
	}

	// 4 ----- send syslog message to remote storage
	h.storage.Insert("syslog", string(storageMessageJson))

	// --------------------
	// Implement Beward-specific message processing here
	// Track debug msg
	if strings.Contains(message.Message, "cancel button") || strings.Contains(message.Message, "Emulating Cancel button press!") {
		h.HandleDebug(&now, host, message.Message)
	}

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

	// Tracks open door by code
	if strings.Contains(message.Message, "Opening door by code") {
		h.HandleOpenByCode(&now, host, message.Message)
	}

	// Tracks open door by RFID key
	if strings.Contains(message.Message, "Opening door by RFID") ||
		strings.Contains(message.Message, "Opening door by external RFID") {
		h.HandleOpenByRFID(&now, host, message.Message)
	}

	// Tracks open door by button
	if strings.Contains(message.Message, "door button pressed") {
		h.HandleOpenByButton(&now, host, message.Message)
	}

	// TODO: implement me
	// Tracks alarm button
	if strings.Contains(message.Message, "Intercom break in detected") {
		h.logger.Debug("processing not implemented", "msg", message.Message)
	}

	// TODO: implement me
	// 		- Tracks calls
	if strings.Contains(message.Message, "CMS handset call started") ||
		strings.Contains(message.Message, "CMS handset talk started") ||
		strings.Contains(message.Message, "Opening door by CMS handset") ||
		strings.Contains(message.Message, "CMS handset call done") ||
		strings.Contains(message.Message, "Calling sip:") ||
		strings.Contains(message.Message, "SIP call") ||
		strings.Contains(message.Message, "SIP talk started") ||
		strings.Contains(message.Message, "SIP call done") ||
		strings.Contains(message.Message, "All calls are done") ||
		strings.Contains(message.Message, "Unable to call CMS") {
		h.HandleCallFlow(&now, host, message.Message)
	}
}

// FIXME:
// APICallToRBT Update RFID usage timestamp
func (h *BewardHandler) APICallToRBT(payload *OpenDoorMsg) error {
	//url := "http://172.28.0.2/internal/actions/openDoor"
	url := "https://webhook.site/5e9d4c4d-73eb-44c4-be20-e1886cbea2b4/openDoor"

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

// HandleMotionDetection - complete
func (h *BewardHandler) HandleMotionDetection(timestamp *time.Time, host string, motionActive bool) {
	// implement motion detection logic
	// get streamId by intercom IP and call to API FRS. message motion start or stop
	h.logger.Debug("HandleMotionDetection", "host", host, "motionActive", motionActive)
	/**
	TODO:
		1 get stream by ip
		2 check FRS enable, not eq "-"
		3 send req to FRS service
	*/
	camera, _ := h.repo.Cameras.GetCameraByIP(context.Background(), host)
	//h.logger.Debug("Motion detect process", "camera", camera.FRS)
	// check if FRS enable
	if *camera.FRS != "-" {
		//h.logger.Debug("Motion detect process, FRS enabled")
		err := frs.MotionDetection(camera.CameraID, motionActive, *camera.FRS)
		if err != nil {
			h.logger.Warn("Failed to send motion detect to FRS service", "error", err)
			return
		}
	}
}

// HandleOpenByCode - complete
func (h *BewardHandler) HandleOpenByCode(timestamp *time.Time, host, message string) {
	// implement open door by code logic
	h.logger.Debug("Open door by code", "host", host, "message", message)

	frsEnabled := false
	preview := PREVIEW_IPCAM
	rbtAPI := h.rbtApi.Internal
	var faceData map[string]interface{}
	door := 0 // main door usage digit code

	// get code
	// TODO: move get code to utils
	parts := strings.SplitN(message, "code", 2)
	if len(parts) < 2 {
		h.logger.Error("Invalid message format - no content after 'code'", "message", message)
		return
	}

	codePart := strings.SplitN(parts[1], ",", 2)[0]
	codeStr := strings.TrimSpace(codePart)

	code, err := strconv.Atoi(codeStr)
	if err != nil {
		h.logger.Error("Failed to convert code to integer",
			"error", err,
			"code_str", codeStr,
			"message", message)
		return
	}
	h.logger.Debug("Successfully extracted code", "code", code, "host", host)

	// get domophone
	domophone, err := h.repo.Households.GetDomophone(context.Background(), "ip", host)
	if err != nil {
		h.logger.Warn("Failed to get domophone", "error", err)
	}

	// get entrance
	entrance, err := h.repo.Households.GetEntrance(context.Background(), domophone.HouseDomophoneID, door)
	if err != nil {
		h.logger.Warn("Failed to get entrance", "error", err)
	}
	h.logger.Debug("Successfully extracted entrance", "entrance", entrance)

	if entrance.CameraID == nil {
		h.logger.Warn("Failed to get camera id")
		return
	}

	// get entrance camera
	camera, err := h.repo.Cameras.GetCamera(context.Background(), *entrance.CameraID)
	if err != nil {
		h.logger.Warn("Failed to get camera", "error", err)
	}

	// check FRS enabled
	if *camera.FRS != "-" {
		h.logger.Debug("HandleDebug, FRS enabled")
		frsEnabled = true
	}

	// 01 - get screenshot from domophone camera
	imgURL := rbtAPI + "/frs/camshot/" + strconv.Itoa(camera.CameraID)
	camScreenShot, err := utils.DownloadFile(imgURL)
	if err != nil {
		h.logger.Debug("RBT DownloadFile", "err", err)
	}

	// 02 get screenshot from FRS
	if frsEnabled {
		h.logger.Debug("HandleMessage | HandleDebug | FRS enabled")
		bqResponse, _ := utils.GetBestQuality(h.frsApi, camera.CameraID, *timestamp)
		if bqResponse != nil {
			h.logger.Debug("FRS BEST bq response", "response", bqResponse)

			faceData = map[string]interface{}{
				"left":   bqResponse.Data.Left,
				"top":    bqResponse.Data.Top,
				"width":  bqResponse.Data.Width,
				"height": bqResponse.Data.Height,
			}
			preview = PREVIEW_FRS

			h.logger.Debug("HandleMessage | HandleDebug | get img from FRS")
			camScreenShot = nil
			camScreenShot, err = utils.DownloadFile(bqResponse.Data.Screenshot)
		}
	}

	metadata := map[string]interface{}{
		"contentType": "image/jpeg",
		"expire":      int32(timestamp.Add(TTL_CAMSHOT_HOURS).Unix()),
	}

	// save data to MongoDb
	fileId, err := h.fsFiles.SaveFile("camshot", metadata, camScreenShot)
	if err != nil {
		h.logger.Debug("MongoDB SaveFile", "err", err)
	}
	h.logger.Debug("MongoDB SaveFile", "fileId", fileId)
	camScreenShot = nil

	// 9
	eventGUIDv4 := uuid.New().String()
	imageGUIDv4 := utils.ToGUIDv4(fileId)

	flatList, _ := h.repo.Households.GetFlatIDsByCode(context.Background(), strconv.Itoa(code))

	plogData := map[string]interface{}{
		"date":       timestamp.Unix(),
		"event_uuid": eventGUIDv4,
		"hidden":     0,
		"image_uuid": imageGUIDv4,
		"flat_id":    flatList[0],
		"domophone": map[string]interface{}{
			"camera_id":             camera.CameraID,
			"domophone_description": entrance.Entrance,
			"domophone_id":          domophone.HouseDomophoneID,
			"domophone_output":      entrance.DomophoneOutput,
			"entrance_id":           entrance.HouseEntranceID,
			"house_id":              entrance.AddressHouseID,
		},
		"event":   Event.OpenByCode,
		"opened":  1, // bool
		"face":    faceData,
		"rfid":    "",
		"code":    code,
		"phones":  map[string]interface{}{},
		"preview": preview, // 0 no image, 1 - image from DVR, 2 - image from FRS
	}

	plogDataString, err := json.Marshal(plogData)
	if err != nil {
		h.logger.Debug("Failed marshal JSON")
	}

	err = h.storage.Insert("plog", string(plogDataString))
	if err != nil {
		fmt.Println("INSERT ERR", err)
	}

	// get flat by code and domophone ip
	// TODO: update last usage code
}

// HandleOpenByRFID - complete
func (h *BewardHandler) HandleOpenByRFID(timestamp *time.Time, host, message string) {
	// implement open door by RFID key logic
	h.logger.Debug("Open door by RFID")
	frsEnabled := false
	isExternalReader := false
	var faceData map[string]interface{}
	preview := PREVIEW_IPCAM

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
	//isExternalReader := false
	if strings.Contains(message, "external") {
		isExternalReader = true
	}

	// ----- 2
	door := 0
	if isExternalReader {
		door = 1
	}

	// ----- 3
	rfidKey := utils.ExtractRFIDKey(message)
	if rfidKey != "" {
		h.logger.Debug("RFID key found", "host", host, "rfid", rfidKey)
	} else {
		h.logger.Warn("RFID key not found", "host", host)
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

	// TODO: implement me
	// ----- 4
	err := h.repo.Households.UpdateRFIDLastSeen(context.Background(), rfidKey)
	if err != nil {
		h.logger.Warn("Failed to update RFID", "error", err)
		return
	}

	//domophoneId, _ := h.repo.Households.GetDomophoneByIP(context.Background(), host)
	/*
		+ 1 получаем домофон по ip
		2 полчаем вход (основной или дополнительный)  на основании считывателя
		3 получаем камеру входа
	*/

	//domophone, _ := h.repo.Households.GetDomophone(context.Background(), "ip", host)

	// ----- 5
	// TODO: implement get "streamName" and "streamID" by ip intercom

	// get domophone
	domophone, err := h.repo.Households.GetDomophone(context.Background(), "ip", host)
	if err != nil {
		h.logger.Warn("Failed to get domophone", "error", err)
	}

	// get entrance
	entrance, err := h.repo.Households.GetEntrance(context.Background(), domophone.HouseDomophoneID, door)
	if err != nil {
		h.logger.Warn("Failed to get entrance", "error", err)
	}

	if entrance.CameraID == nil {
		h.logger.Warn("Failed to get camera id")
		return
	}

	// get entrance camera
	camera, err := h.repo.Cameras.GetCamera(context.Background(), *entrance.CameraID)
	if err != nil {
		h.logger.Warn("Failed to get camera", "error", err)
	}

	/**
	TODO:
		get screenshot from camera
		frs enabled - get GetBestQuality
			if GetBestQuality result - use FRS screenshot
			if not GetBestQuality result - use first camera screen
	*/
	// check FRS enabled
	if *camera.FRS != "-" {
		h.logger.Debug("HandleDebug, FRS enabled")
		frsEnabled = true
	}
	h.logger.Debug("Open by RFID", "frsEnabled >>", frsEnabled)

	rbtAPI := h.rbtApi.Internal

	// download file from RBT
	//_screenShot, err := utils.DownloadFile(_imageUrl)
	//if err != nil {
	//	h.logger.Debug("FRS DownloadFile", "err", err)
	//}
	//
	//// if frs enabled - call bestQuality
	//if *camera.FRS != "-" {
	//
	//	bestQualityResp, err := utils.GetBestQuality(camera.CameraID, *timestamp)
	//	if err != nil {
	//
	//		return
	//	}
	//
	//	if bestQualityResp == nil {
	//		h.logger.Debug("Open by RFID", "GetBestQuality response", bestQualityResp)
	//
	//		// download file from FRS
	//		//https://rbt-demo.lanta.me:8443/rbt-demo-0016/index.m3u8?token=phei9quohmoochoth5es3eo9Koh5ua9i
	//		//https://rbt-demo.lanta.me:8443/rbt-demo-0016/preview.mp4?token=phei9quohmoochoth5es3eo9Koh5ua9i
	//
	//	}
	//} else {
	//	//get image from DVR
	//	h.logger.Debug("Open by RFID, FRS is DIS")
	//	h.logger.Debug("Open by RFID", "DVR URL", camera.DVRStream)
	//}

	// 01 - get screenshot from domophone camera
	imgURL := rbtAPI + "/frs/camshot/" + strconv.Itoa(camera.CameraID)
	h.logger.Debug("Open by RFID", "_imageUrl >>", imgURL)

	camScreenShot, err := utils.DownloadFile(imgURL)
	if err != nil {
		h.logger.Debug("RBT DownloadFile", "err", err)
	}

	// 02 get screenshot from FRS
	if frsEnabled {
		h.logger.Debug("HandleMessage | HandleDebug | FRS enabled")
		bqResponse, _ := utils.GetBestQuality(h.frsApi, camera.CameraID, *timestamp)
		if bqResponse != nil {
			h.logger.Debug("FRS BEST bq response", "response", bqResponse)

			faceData = map[string]interface{}{
				"left":   bqResponse.Data.Left,
				"top":    bqResponse.Data.Top,
				"width":  bqResponse.Data.Width,
				"height": bqResponse.Data.Height,
			}
			preview = PREVIEW_FRS

			h.logger.Debug("HandleMessage | HandleDebug | get img from FRS")
			camScreenShot = nil
			camScreenShot, err = utils.DownloadFile(bqResponse.Data.Screenshot)
		}
	}

	metadata := map[string]interface{}{
		"contentType": "image/jpeg",
		"expire":      int32(timestamp.Add(time.Hour * 24 * 30 * 6).Unix()),
	}

	// save data to MongoDb
	fileId, err := h.fsFiles.SaveFile("camshot", metadata, camScreenShot)
	if err != nil {
		h.logger.Debug("MongoDB SaveFile", "err", err)
	}
	h.logger.Debug("MongoDB SaveFile", "fileId", fileId)
	camScreenShot = nil

	// 9
	eventGUIDv4 := uuid.New().String()
	imageGUIDv4 := utils.ToGUIDv4(fileId)

	flatList, _ := h.repo.Households.GetFlatIDsByRFID(context.Background(), rfidKey)

	// TODO: We're currently updating only one apartment out of the ones found.
	//		Add processing to all apartments using this RFID key.
	plogData := map[string]interface{}{
		"date":       timestamp.Unix(),
		"event_uuid": eventGUIDv4,
		"hidden":     0,
		"image_uuid": imageGUIDv4,
		"flat_id":    flatList[0],
		"domophone": map[string]interface{}{
			"camera_id":             camera.CameraID,
			"domophone_description": entrance.Entrance,
			"domophone_id":          domophone.HouseDomophoneID,
			"domophone_output":      entrance.DomophoneOutput,
			"entrance_id":           entrance.HouseEntranceID,
			"house_id":              entrance.AddressHouseID,
		},
		"event":   Event.OpenByKey,
		"opened":  1, // bool
		"face":    faceData,
		"rfid":    rfidKey,
		"code":    "",
		"phones":  map[string]interface{}{},
		"preview": preview, // 0 no image, 1 - image from DVR, 2 - image from FRS
	}

	plogDataString, err := json.Marshal(plogData)
	if err != nil {
		h.logger.Debug("Failed marshal JSON")
	}

	err = h.storage.Insert("plog", string(plogDataString))
	if err != nil {
		fmt.Println("INSERT ERR", err)
	}
}

// HandleOpenByButton - not implemented
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

// -------

func (h *BewardHandler) HandleCallFlow(timestamp *time.Time, host, message string) {
	// implement call flow logic
	h.logger.Info("⚠️ HandleCallFlow Start")
	callID, err := h.extractCallID(message)
	if err != nil {
		h.logger.Warn("HandleCallFlow extractCallID", "err", err)
	}

	// 01 - Call start +
	if strings.Contains(message, "CMS handset call started for apartment") ||
		strings.Contains(message, "Unable to call CMS apartment") || // CMS disabled
		strings.Contains(message, "Calling sip:") {
		h.HandleCallStart(timestamp, host, message, callID)
		return
	}

	// 02 - Call answered +
	if strings.Contains(message, "CMS handset talk started for apartment") ||
		strings.Contains(message, "SIP talk started for apartment") {
		h.HandleCallAnswered(timestamp, host, message, callID)
		return
	}

	// Door opened +
	if strings.Contains(message, "Opening door by CMS handset for apartment") {
		h.HandleDoorOpen(timestamp, host, message, callID)
		return
	}

	// Call ended, make event
	if strings.Contains(message, "CMS handset call done for apartment") ||
		strings.Contains(message, "SIP call done for apartment") {
		h.HandleCallEnd(timestamp, host, message, callID)
		return
	}

	// Clear call flow
	if strings.Contains(message, "All calls are done for apartment") {
		h.HandleAllCallsDone(timestamp, host, message, callID)
		return
	}
}

// HandleCallStart -  get base event info
func (h *BewardHandler) HandleCallStart(timestamp *time.Time, host string, message string, callID int) {
	h.logger.Info("🎃 - HandleCallStart start")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// get flat number
	apartment, err := h.extractApartment(message)
	if err != nil {
		h.logger.Warn("Failed to extract apartment from call start", "message", message)
		return
	}

	// call type: SIP or CMS
	callType := CALL_TYPE_SIP
	if strings.Contains(message, "CMS handset call started") {
		callType = CALL_TYPE_CMS
	}

	h.callMutex.Lock()
	defer h.callMutex.Unlock()

	// get domophone data
	domophone, err := h.repo.Households.GetDomophone(ctx, "ip", host)
	if err != nil {
		h.logger.Warn("Failed to get domophone", "callID", callID, "err", err)
		return
	}

	// get entrance
	entrance, err := h.repo.Households.GetEntrance(ctx, domophone.HouseDomophoneID, 0)
	if err != nil {
		h.logger.Warn("Failed to get entrance info", "callID", callID, "error", err)
		return
	}

	// get camera data
	camera, err := h.repo.Cameras.GetCamera(ctx, *entrance.CameraID)
	if err != nil {
		h.logger.Warn("Failed to get camera", "callID", callID, "error", err)
		return
	}

	// get flat
	flatID, err := h.repo.Households.GetFlatIDByApartment(ctx, apartment, domophone.HouseDomophoneID)
	if err != nil {
		h.logger.Warn("Failed to get flatID", "callID", callID, "error", err)
		return
	}

	// make data structure
	callData := &CallData{
		CallID:      callID,
		Apartment:   apartment,
		DomophoneIP: host,
		StartTime:   timestamp,
		CallType:    callType,
		Answered:    false,
		DoorOpened:  false,
		Domophone:   domophone,
		Entrance:    entrance,
		FlatID:      flatID,
	}

	// store cameraId if exist
	if entrance.CameraID != nil {
		callData.CameraID = *entrance.CameraID
	}
	if *camera.FRS != "-" {
		callData.CameraFRS = *camera.FRS
	}

	h.activeCalls[callID] = callData

	h.logger.Info("️️️️️️⚠️️️ Call started - data collected",
		"callID", callID,
		"apartment", apartment,
		"domophone", domophone.HouseDomophoneID,
		"flatID", flatID)

	// get cam screenshot
	go h.getCallScreenshots(callData)
}

// utils,  get callID and apartment
func (h *BewardHandler) extractCallID(message string) (int, error) {
	start := strings.Index(message, "[")
	if start == -1 {
		return 0, fmt.Errorf("opening bracket not found")
	}

	end := strings.Index(message, "]")
	if end == -1 || end <= start {
		return 0, fmt.Errorf("closing bracket not found or invalid position")
	}

	callIDStr := message[start+1 : end]
	callIDStr = strings.TrimSpace(callIDStr)

	callID, err := strconv.Atoi(callIDStr)
	if err != nil {
		return 0, fmt.Errorf("call ID is not a valid number: %s", callIDStr)
	}

	return callID, nil
}

func (h *BewardHandler) extractApartment(message string) (int, error) {
	if idx := strings.Index(message, "apartment "); idx != -1 {
		start := idx + len("apartment ")
		end := strings.IndexAny(message[start:], " ,!].")
		if end == -1 {
			end = len(message) - start
		}

		apartmentStr := strings.TrimSpace(message[start : start+end])
		return strconv.Atoi(apartmentStr)
	}
	return 0, fmt.Errorf("apartment not found in message")
}

func (h *BewardHandler) extractSIPCallID(message string) (int, error) {
	if idx := strings.Index(message, "SIP call "); idx != -1 {
		start := idx + len("SIP call ")
		end := strings.IndexAny(message[start:], " ]")
		if end == -1 {
			end = len(message) - start
		}

		sipCallIDStr := strings.TrimSpace(message[start : start+end])
		sipCallID, err := strconv.Atoi(sipCallIDStr)
		if err != nil {
			return 0, fmt.Errorf("SIP call ID '%s' is not a valid number", sipCallIDStr)
		}

		return sipCallID, nil
	}
	return 0, fmt.Errorf("SIP call pattern not found in message")
}

//func (h *BewardHandler) extractSIPCallID(message string) string {
//	// Ищем паттерн "SIP call X"
//	if idx := strings.Index(message, "SIP call "); idx != -1 {
//		start := idx + len("SIP call ")
//		end := strings.IndexAny(message[start:], " ]")
//		if end == -1 {
//			return message[start:]
//		}
//		return message[start : start+end]
//	}
//	return ""
//}

func (h *BewardHandler) HandleCallAnswered(timestamp *time.Time, host string, message string, callID int) {
	// TODO: implement me!
	h.logger.Info("🎃 -  HandleCallAnswered start")

	h.callMutex.Lock()
	defer h.callMutex.Unlock()

	callData, exists := h.activeCalls[callID]
	if !exists {
		h.logger.Debug("Call answered for unknown call", "callID", callID)
		return
	}

	callData.Answered = true

	h.logger.Info("Call answered",
		"callID", callID,
		"apartment", callData.Apartment)
}

func (h *BewardHandler) HandleDoorOpen(timestamp *time.Time, host string, message string, callID int) {
	// TODO: implement me!
	// set door opened for call state
	h.logger.Info("🎃 -  HandleDoorOpen start")

	h.callMutex.Lock()
	defer h.callMutex.Unlock()

	callData, exists := h.activeCalls[callID]
	if !exists {
		h.logger.Debug("Call door open for unknown call", "callID", callID)
	}

	callData.DoorOpened = true
	h.logger.Info("Door has opened by call",
		"callID", callID,
		"apartment", callData.Apartment)
}

func (h *BewardHandler) HandleCallEnd(timestamp *time.Time, host string, message string, callID int) {
	// TODO: implement me!
	h.logger.Info("🎃 - HandleCallEnd start")

	h.callMutex.Lock()
	callData, exists := h.activeCalls[callID]
	defer h.callMutex.Unlock()

	if !exists {
		h.logger.Debug("Call end for unknown call", "callID", callID)
		return
	}

	callData.EndTime = timestamp

	h.logger.Info("Call ended, prepare event",
		"callID", callID,
		"apartment", callData.Apartment,
		"answered", callData.Answered,
		"doorOpen", callData.DoorOpened,
	)

	// process make event

	go h.prepareFinalCallEvent(callData)
}

func (h *BewardHandler) prepareFinalCallEvent(callData *CallData) {
	h.logger.Info("🎃 - prepareFinalCallEvent start")
	startTime := time.Now()

	// Wait until the screenshots are ready (maximum 15 seconds)
	for i := 0; i < 30; i++ {
		callData.callMutex.Lock()
		ready := callData.screenshotsReady
		callData.callMutex.Unlock()

		if ready {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	callData.callMutex.Lock()
	fileID := callData.screenshotFileID
	callData.callMutex.Unlock()

	if fileID == "" {
		h.logger.Warn("Screenshots not ready for final event", "callID", callData.CallID)
		// Все равно сохраняем событие, но без изображения
		fileID = "no_screenshot"
	}

	// Формируем и сохраняем итоговое событие
	h.saveFinalCallEvent(callData, fileID)

	h.logger.Info("Call processing completed",
		"callID", callData.CallID,
		"duration", time.Since(startTime))
}

func (h *BewardHandler) HandleAllCallsDone(timestamp *time.Time, host string, message string, callID int) {
	h.logger.Info("🎃 - HandleAllCallsDone start")
	apartment, err := h.extractApartment(message)
	if err != nil {
		h.logger.Warn("Failed to extract apartment from all calls done", "callID", callID)
		return
	}

	h.callMutex.Lock()
	defer h.callMutex.Unlock()

	// Удаляем все звонки для этой квартиры
	for id, callData := range h.activeCalls {
		if callData.Apartment == apartment {
			delete(h.activeCalls, id)
			h.logger.Debug("Removed call data", "callID", id, "apartment", apartment)
		}
	}

	h.logger.Info("🎃 All calls completed for apartment", "apartment", apartment)
}

//func (h *BewardHandler) processCallEvent(callData *CallData) {
//	// process event
//	//startTime := time.Now()
//	//h.logger.Info("Starting call event processing",
//	//	"callId", callData.CallID,
//	//	"apartment", callData.Apartment)
//	//
//	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	//defer cancel()
//
//	// get
//}

func (h *BewardHandler) getCallScreenshots(callData *CallData) {
	// TODO : implement get image from cam and best screen from FRS service
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h.logger.Info("Starting call screenshots processing", "callId", callData.CallID)

	// + 1 get screen from camera domophone
	if err := h.getCameraScreenshot(ctx, callData); err != nil {
		h.logger.Warn("Failed to get screenshot from domophone camera", "callId", callData.CallID, "error", err)
	} else {
		h.logger.Debug("Camera screenshot completed", "callID", callData.CallID, "dataSize", len(callData.ScreenshotData))
	}

	// 2 get image from FRS (optional)
	// TODO: refactor CameraFRS value
	if callData.CameraID > 0 && callData.CameraFRS != "" {
		if err := h.getFRSBestQuality(ctx, callData); err != nil {
			h.logger.Debug("FRS best quality not available", "err", err)
		} else {
			h.logger.Debug("FRS processing completed")
		}
	}

	// 3 store image
	fileID, err := h.saveScreenshotToMongo(callData)
	if err != nil {
		h.logger.Warn("Failed to save screenshot to Mongo", "callId", callData.CallID, "error", err)
	}

	// 4 update callData
	callData.callMutex.Lock()
	callData.screenshotFileID = fileID
	callData.screenshotsReady = true
	callData.callMutex.Unlock()

	h.logger.Info("Call screenshot processed", "callId", callData.CallID, "fileID", fileID)
}

func (h *BewardHandler) getCameraScreenshot(ctx context.Context, callData *CallData) error {
	if callData.CameraID == 0 {
		return fmt.Errorf("no camera ID available")
	}

	imgURL := h.rbtApi.Internal + "/frs/camshot/" + strconv.Itoa(callData.CameraID)

	h.logger.Debug("Getting camera screenshot", "callID", callData.CallID, "url", imgURL)

	screenshotData, err := utils.DownloadFile(imgURL)
	if err != nil {
		return fmt.Errorf("failed to download screenshot: %w", err)
	}

	callData.ScreenshotData = screenshotData
	callData.PreviewType = PREVIEW_IPCAM

	h.logger.Debug("Camera screenshot obtained", "callID", callData.CallID, "size", len(screenshotData))
	return nil
}
func (h *BewardHandler) getFRSBestQuality(ctx context.Context, callData *CallData) error {
	h.logger.Debug("Getting FRS best quality", "callID", callData.CallID, "camera", callData.CameraID)

	// Get best frame from FRS (call started screenshot)
	bqResponse, err := utils.GetBestQuality(h.frsApi, callData.CameraID, *callData.StartTime)
	if err != nil {
		return fmt.Errorf("failed to get FRS best quality: %w", err)
	}

	if bqResponse == nil || bqResponse.Data.Screenshot == "" {
		return fmt.Errorf("no FRS best quality data available")
	}

	// get screenshot from FRS service
	frsScreenshot, err := utils.DownloadFile(bqResponse.Data.Screenshot)
	if err != nil {
		return fmt.Errorf("failed to download FRS screenshot: %w", err)
	}

	// use FRS image
	callData.ScreenshotData = frsScreenshot
	callData.PreviewType = PREVIEW_FRS

	// set face data
	callData.FaceData = map[string]interface{}{
		"left":   bqResponse.Data.Left,
		"top":    bqResponse.Data.Top,
		"width":  bqResponse.Data.Width,
		"height": bqResponse.Data.Height,
	}

	h.logger.Debug("FRS best quality obtained", "callID", callData.CallID, "faceData", callData.FaceData)
	return nil
}
func (h *BewardHandler) saveScreenshotToMongo(callData *CallData) (string, error) {
	if callData.ScreenshotData == nil {
		return "", fmt.Errorf("no screenshot data available")
	}

	metadata := map[string]interface{}{
		"contentType": "image/jpeg",
		"expire":      int32(time.Now().Add(TTL_CAMSHOT_HOURS).Unix()),
		"call_id":     callData.CallID,           // optional, test field
		"camera_id":   callData.CameraID,         // optional, test field
		"timestamp":   callData.StartTime.Unix(), // optional, test field
	}

	fileID, err := h.fsFiles.SaveFile("camshot", metadata, callData.ScreenshotData)
	if err != nil {
		return "", fmt.Errorf("failed to save file to MongoDB: %w", err)
	}

	// clear memory
	callData.ScreenshotData = nil

	h.logger.Debug("Screenshot saved to MongoDB", "callID", callData.CallID, "fileID", fileID)
	return fileID, nil
}

func (h *BewardHandler) saveFinalCallEvent(callData *CallData, fileID string) {
	h.logger.Info("🎃 - saveFinalCallEvent start")
	// Определяем тип события на основе того, что произошло во время звонка
	eventType := Event.NotAnswered
	opened := 0

	if callData.Answered && callData.DoorOpened {
		eventType = Event.Answered
		opened = 1
	} else if callData.Answered {
		eventType = Event.Answered
	}

	// make GUID from fileID
	imageGUID := utils.ToGUIDv4(fileID)

	//  make plof data
	plogData := map[string]interface{}{
		"date":       callData.StartTime.Unix(),
		"event_uuid": uuid.New().String(),
		"hidden":     0,
		"image_uuid": imageGUID,
		"flat_id":    callData.FlatID,
		"domophone": map[string]interface{}{
			"camera_id":             callData.CameraID,
			"domophone_description": callData.Entrance.Entrance,
			"domophone_id":          callData.Domophone.HouseDomophoneID,
			"domophone_output":      callData.Entrance.DomophoneOutput,
			"entrance_id":           callData.Entrance.HouseEntranceID,
			"house_id":              callData.Entrance.AddressHouseID,
		},
		"event":   eventType,
		"opened":  opened,
		"face":    callData.FaceData,
		"preview": callData.PreviewType,
		//"call_info": map[string]interface{}{
		//	"call_id":     callData.CallID,
		//	"apartment":   callData.Apartment,
		//	"call_type":   callData.CallType,
		//	"answered":    callData.Answered,
		//	"door_opened": callData.DoorOpened,
		//	"duration":    callData.EndTime.Sub(*callData.StartTime).Seconds(),
		//},
		"phones": map[string]interface{}{},
	}

	// Конвертируем в JSON
	plogDataString, err := json.Marshal(plogData)
	if err != nil {
		h.logger.Warn("Failed to marshal final call event", "callID", callData.CallID, "error", err)
		return
	}

	// Сохраняем в хранилище
	err = h.storage.Insert("plog", string(plogDataString))
	if err != nil {
		h.logger.Warn("Failed to insert final call event to plog", "callID", callData.CallID, "error", err)
	} else {
		h.logger.Info("🎃 - Final call event saved to plog",
			"callID", callData.CallID,
			"apartment", callData.Apartment,
			"eventType", eventType,
			"answered", callData.Answered,
			"doorOpened", callData.DoorOpened)
	}
}

// --- debug
func (h *BewardHandler) HandleDebugRFID(timestamp *time.Time, host, message string) {
	h.logger.Debug("HandleMessage | HandleDebug", "timestamp", timestamp)
	//dummy data
	door := 0
	//rbtAPI := "https://rbt-demo.lanta.me:55544/internal"
	rbtAPI := h.rbtApi.Internal
	frsEnabled := false
	fakeRFID := "00000004030201"
	var faceData map[string]interface{}
	preview := 1

	// get domophone
	domophone, err := h.repo.Households.GetDomophone(context.Background(), "ip", host)
	if err != nil {
		h.logger.Warn("Failed to get domophone", "error", err)
	}

	// get entrance
	entrance, err := h.repo.Households.GetEntrance(context.Background(), domophone.HouseDomophoneID, door)
	if err != nil {
		h.logger.Warn("Failed to get entrance", "error", err)
	}

	// get entrance camera
	camera, err := h.repo.Cameras.GetCamera(context.Background(), *entrance.CameraID)
	if err != nil {
		h.logger.Warn("Failed to get camera", "error", err)
	}

	// check FRS enabled
	if *camera.FRS != "-" {
		h.logger.Debug("HandleDebug, FRS enabled")
		frsEnabled = true
	}

	// 01 - get screenshot from domophone camera
	imgURL := rbtAPI + "/frs/camshot/" + strconv.Itoa(camera.CameraID)
	camScreenShot, err := utils.DownloadFile(imgURL)
	if err != nil {
		h.logger.Debug("RBT DownloadFile", "err", err)
	}

	// 02 get screenshot from FRS
	if frsEnabled {
		h.logger.Debug("HandleMessage | HandleDebug | FRS enabled")
		bqResponse, _ := utils.GetBestQuality(h.frsApi, camera.CameraID, *timestamp)
		if bqResponse != nil {
			h.logger.Debug("FRS BEST bq response", "response", bqResponse)

			faceData = map[string]interface{}{
				"left":   bqResponse.Data.Left,
				"top":    bqResponse.Data.Top,
				"width":  bqResponse.Data.Width,
				"height": bqResponse.Data.Height,
			}
			preview = 2

			h.logger.Debug("HandleMessage | HandleDebug | get img from FRS")
			camScreenShot = nil
			camScreenShot, err = utils.DownloadFile(bqResponse.Data.Screenshot)
		}
	}

	metadata := map[string]interface{}{
		"contentType": "image/jpeg",
		"expire":      int32(timestamp.Add(TTL_CAMSHOT_HOURS).Unix()),
	}

	// save data to MongoDb
	fileId, err := h.fsFiles.SaveFile("camshot", metadata, camScreenShot)
	if err != nil {
		h.logger.Debug("MongoDB SaveFile", "err", err)
	}
	h.logger.Debug("MongoDB SaveFile", "fileId", fileId)
	camScreenShot = nil

	// 9
	eventGUIDv4 := uuid.New().String()
	imageGUIDv4 := utils.ToGUIDv4(fileId)

	flatList, _ := h.repo.Households.GetFlatIDsByRFID(context.Background(), fakeRFID)

	plogData := map[string]interface{}{
		"date":       timestamp.Unix(),
		"event_uuid": eventGUIDv4,
		"hidden":     0,
		"image_uuid": imageGUIDv4,
		"flat_id":    flatList[0],
		"domophone": map[string]interface{}{
			"camera_id":             camera.CameraID,
			"domophone_description": entrance.Entrance,
			"domophone_id":          domophone.HouseDomophoneID,
			"domophone_output":      entrance.DomophoneOutput,
			"entrance_id":           entrance.HouseEntranceID,
			"house_id":              entrance.AddressHouseID,
		},
		"event":   Event.OpenByKey,
		"opened":  1, // bool
		"face":    faceData,
		"rfid":    fakeRFID,
		"code":    "",
		"phones":  map[string]interface{}{},
		"preview": preview, // 0 no image, 1 - image from DVR, 2 - image from FRS
	}

	// TODO: add face data from frs
	// from frs response  example bqResponse:
	// {"code":"200","message":"Request completed successfully","data":{"screenshot":"https://rbt-demo.lanta.me/.well-known/frs/screenshots/group_3/d/b/1/4/db141e98cbe2437f83c41e7b7a71454f.jpg","left":554,"top":259,"with":0,"height":205}}}

	plogDataString, err := json.Marshal(plogData)
	if err != nil {
		h.logger.Debug("Failed marshal JSON")
	}

	err = h.storage.Insert("plog", string(plogDataString))
	if err != nil {
		fmt.Println("INSERT ERR", err)
	}
}
func (h *BewardHandler) HandleDebugCode(timestamp *time.Time, host, message string) {
	h.logger.Debug("HandleMessage | HandleDebugCode", "timestamp", timestamp)

	frsEnabled := false
	preview := 1
	rbtAPI := h.rbtApi.Internal
	fakeMsg := "Opening door by code 55544, apartment 1"
	message = fakeMsg
	var faceData map[string]interface{}
	door := 0 // main door usage digit code

	// get code
	parts := strings.SplitN(message, "code", 2)
	if len(parts) < 2 {
		h.logger.Error("Invalid message format - no content after 'code'", "message", message)
		return
	}

	codePart := strings.SplitN(parts[1], ",", 2)[0]
	codeStr := strings.TrimSpace(codePart)

	code, err := strconv.Atoi(codeStr)
	if err != nil {
		h.logger.Error("Failed to convert code to integer",
			"error", err,
			"code_str", codeStr,
			"message", message)
		return
	}
	h.logger.Debug("Successfully extracted code", "code", code, "host", host)

	// get domophone
	domophone, err := h.repo.Households.GetDomophone(context.Background(), "ip", host)
	if err != nil {
		h.logger.Warn("Failed to get domophone", "error", err)
	}

	// get entrance
	entrance, err := h.repo.Households.GetEntrance(context.Background(), domophone.HouseDomophoneID, door)
	if err != nil {
		h.logger.Warn("Failed to get entrance", "error", err)
	}
	h.logger.Debug("Successfully extracted entrance", "entrance", entrance)

	if entrance.CameraID == nil {
		h.logger.Warn("Failed to get camera id")
		return
	}

	// get entrance camera
	camera, err := h.repo.Cameras.GetCamera(context.Background(), *entrance.CameraID)
	if err != nil {
		h.logger.Warn("Failed to get camera", "error", err)
	}

	// check FRS enabled
	if *camera.FRS != "-" {
		h.logger.Debug("HandleDebug, FRS enabled")
		frsEnabled = true
	}

	// 01 - get screenshot from domophone camera
	imgURL := rbtAPI + "/frs/camshot/" + strconv.Itoa(camera.CameraID)
	camScreenShot, err := utils.DownloadFile(imgURL)
	if err != nil {
		h.logger.Debug("RBT DownloadFile", "err", err)
	}

	// 02 get screenshot from FRS
	if frsEnabled {
		h.logger.Debug("HandleMessage | HandleDebug | FRS enabled")
		bqResponse, _ := utils.GetBestQuality(h.frsApi, camera.CameraID, *timestamp)
		if bqResponse != nil {
			h.logger.Debug("FRS BEST bq response", "response", bqResponse)

			faceData = map[string]interface{}{
				"left":   bqResponse.Data.Left,
				"top":    bqResponse.Data.Top,
				"width":  bqResponse.Data.Width,
				"height": bqResponse.Data.Height,
			}
			preview = 2

			h.logger.Debug("HandleMessage | HandleDebug | get img from FRS")
			camScreenShot = nil
			camScreenShot, err = utils.DownloadFile(bqResponse.Data.Screenshot)
		}
	}

	metadata := map[string]interface{}{
		"contentType": "image/jpeg",
		"expire":      int32(timestamp.Add(time.Hour * 24 * 30 * 6).Unix()),
	}

	// save data to MongoDb
	fileId, err := h.fsFiles.SaveFile("camshot", metadata, camScreenShot)
	if err != nil {
		h.logger.Debug("MongoDB SaveFile", "err", err)
	}
	h.logger.Debug("MongoDB SaveFile", "fileId", fileId)
	camScreenShot = nil

	// 9
	eventGUIDv4 := uuid.New().String()
	imageGUIDv4 := utils.ToGUIDv4(fileId)

	flatList, _ := h.repo.Households.GetFlatIDsByCode(context.Background(), strconv.Itoa(code))

	plogData := map[string]interface{}{
		"date":       timestamp.Unix(),
		"event_uuid": eventGUIDv4,
		"hidden":     0,
		"image_uuid": imageGUIDv4,
		"flat_id":    flatList[0],
		"domophone": map[string]interface{}{
			"camera_id":             camera.CameraID,
			"domophone_description": entrance.Entrance,
			"domophone_id":          domophone.HouseDomophoneID,
			"domophone_output":      entrance.DomophoneOutput,
			"entrance_id":           entrance.HouseEntranceID,
			"house_id":              entrance.AddressHouseID,
		},
		"event":   Event.OpenByCode,
		"opened":  1, // bool
		"face":    faceData,
		"rfid":    "",
		"code":    code,
		"phones":  map[string]interface{}{},
		"preview": preview, // 0 no image, 1 - image from DVR, 2 - image from FRS
	}

	plogDataString, err := json.Marshal(plogData)
	if err != nil {
		h.logger.Debug("Failed marshal JSON")
	}

	err = h.storage.Insert("plog", string(plogDataString))
	if err != nil {
		fmt.Println("INSERT ERR", err)
	}
}
func (h *BewardHandler) HandleDebug(timestamp *time.Time, host, message string) {
	h.logger.Debug("HandleMessage | HandleDebugCode", "timestamp", timestamp)
	fakeMsg := "Opening door by RFID 00000004030201, apartment 0"
	h.HandleOpenByRFID(timestamp, host, fakeMsg)
}
