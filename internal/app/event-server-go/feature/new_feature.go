package feature

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/config"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/repository"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/storage"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/utils"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PREVIEW_NONE  = 0
	PREVIEW_IPCAM = 1
	PREVIEW_FRS   = 2

	TTL_CAMSHOT_HOURS = time.Hour * 24 * 30 * 6

	EVENT_UNANSWERED_CALL      = 1
	EVENT_ANSWERED_CALL        = 2
	EVENT_OPENED_BY_KEY        = 3
	EVENT_OPENED_BY_APP        = 4
	EVENT_OPENED_BY_FACE       = 5
	EVENT_OPENED_BY_CODE       = 6
	EVENT_OPENED_GATES_BY_CALL = 7
	EVENT_OPENED_BY_VEHICLE    = 9

	MONGO_SCREENSHOT_NAME = "camshot"
)

// FIXME:
const API_URL_RBT = "https://rbt-demo.lanta.me:55544/internal"
const IMAGE_UUID_STUB = "00000000-0000-0000-0000-000000000000"
const (
	DOOR_MAIN      = 0
	DOOR_SECONDARY = 1
)

// DoorOpenEvent - parse structure from php backend
type DoorOpenEvent struct {
	Date        int64  `json:"date"`
	DomophoneId int    `json:"domophone_id"`
	IP          string `json:"ip"`
	SubID       *int64 `json:"sub_id,omitempty"`
	EventType   int    `json:"event_type"`
	Door        int    `json:"door"`
	Detail      string `json:"detail"`
}

type StreamProcessorConfig struct {
	StreamName     string
	GroupName      string
	WorkersCount   int
	BatchSize      int
	BlockTime      time.Duration
	PendingMinIdle time.Duration
}

type StreamProcessor struct {
	logger  *slog.Logger
	redis   *storage.RedisStorage
	fsFiles *storage.MongoHandler
	storage *storage.ClickhouseHttpClient
	config  StreamProcessorConfig
	wg      sync.WaitGroup
	repo    *repository.PostgresRepository
	frsApi  *config.FrsApi
}

func NewStreamProcessor(
	logger *slog.Logger,
	redisStorage *storage.RedisStorage,
	fsFiles *storage.MongoHandler,
	storage *storage.ClickhouseHttpClient,
	config StreamProcessorConfig,
	repo *repository.PostgresRepository,
	frsApi *config.FrsApi,
) *StreamProcessor {
	return &StreamProcessor{
		logger:  logger,
		redis:   redisStorage,
		fsFiles: fsFiles,
		storage: storage,
		config:  config,
		repo:    repo,
		frsApi:  frsApi,
	}
}

// Start - process stream messages
func (s *StreamProcessor) Start(ctx context.Context) error {
	// Init consumer group
	if err := s.initConsumerGroup(ctx); err != nil {
		return fmt.Errorf("failed to init consumer group: %w", err)
	}

	// Start workers
	for i := 1; i <= s.config.WorkersCount; i++ {
		s.wg.Add(1)
		go s.worker(ctx, fmt.Sprintf("worker_%d", i))
	}

	s.logger.Info("Stream processor started",
		"workers", s.config.WorkersCount,
		"stream", s.config.StreamName,
		"group", s.config.GroupName)

	return nil
}

// initConsumerGroup make consumer group
func (s *StreamProcessor) initConsumerGroup(ctx context.Context) error {
	err := s.redis.Client.XGroupCreateMkStream(
		ctx,
		s.config.StreamName,
		s.config.GroupName,
		"0",
	).Err()

	if err != nil {
		// Consumer group already exist, normal!
		s.logger.Debug("Consumer group already exists", "group", s.config.GroupName)
	} else {
		s.logger.Info("Consumer group created", "group", s.config.GroupName)
	}

	return nil
}

// worker - main worker for message process
func (s *StreamProcessor) worker(ctx context.Context, workerName string) {
	defer s.wg.Done()

	s.logger.Info("Worker started", "worker", workerName)
	processedCount := 0

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Worker stopping",
				"worker", workerName,
				"processed", processedCount)
			return
		default:
			// Read messages from stream
			result, err := s.redis.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    s.config.GroupName,
				Consumer: workerName,
				Streams:  []string{s.config.StreamName, ">"},
				Count:    int64(s.config.BatchSize),
				Block:    s.config.BlockTime,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// timeout - new messages not found. TODO: refactor "error.is"
					continue
				}
				s.logger.Error("Error reading from stream",
					"worker", workerName,
					"error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Process messages fom stream
			for _, stream := range result {
				for _, message := range stream.Messages {
					processedCount++

					// process event task
					if s.processEvent(ctx, message, workerName) {
						// confirm processing
						err := s.redis.Client.XAck(
							ctx,
							stream.Stream,
							s.config.GroupName,
							message.ID,
						).Err()

						if err != nil {
							s.logger.Error("Failed to ack message",
								"worker", workerName,
								"message_id", message.ID,
								"error", err)
						} else {
							s.logger.Debug("Message acknowledged",
								"worker", workerName,
								"message_id", message.ID)
						}
					} else {
						s.logger.Warn("Processing failed, will retry",
							"worker", workerName,
							"message_id", message.ID)
					}

					// logging every 10 messages
					if processedCount%10 == 0 {
						s.logger.Debug("Worker progress",
							"worker", workerName,
							"processed", processedCount)
					}
				}
			}
		}
	}
}

// processEvent - process single event
func (s *StreamProcessor) processEvent(ctx context.Context, message redis.XMessage, workerName string) bool {
	// get payload from message
	payload, ok := message.Values["payload"].(string)
	if !ok {
		s.logger.Error("Invalid payload format",
			"worker", workerName,
			"message_id", message.ID)
		return false
	}

	var event DoorOpenEvent
	err := json.Unmarshal([]byte(payload), &event)
	if err != nil {
		s.logger.Error("Failed to unmarshal event",
			"worker", workerName,
			"message_id", message.ID,
			"error", err)
		return false
	}

	// >> processing event
	s.logger.Debug("Processing door event",
		"worker", workerName,
		"ip", event.IP,
		"event_type", event.EventType,
		"door", event.Door,
		"detail", event.Detail)

	return s.storeEvent(ctx, event)
}

// storeEvent - storage data
func (s *StreamProcessor) storeEvent(ctx context.Context, event DoorOpenEvent) bool {
	type EventType int
	switch EventType(event.EventType) {
	case EVENT_OPENED_BY_APP: // Event open by APP (API)
		return s.processOpenByAPP(ctx, event)
	case EVENT_OPENED_BY_FACE: // Event open by FRS service
		return s.processOpenByFRS(ctx, event)
	default:
		s.logger.Warn("unsupported Event type")
		return false
	}
}

// processOpenByAPP - process events open mobile app
func (s *StreamProcessor) processOpenByAPP(ctx context.Context, event DoorOpenEvent) bool {
	// TODO: implement process alt door open by app

	s.logger.Debug("processOpenByAPP")

	var faceData map[string]interface{} // face data stub
	imageGUIDv4 := ""
	preview := PREVIEW_IPCAM

	// get entrance
	entrance, err := s.repo.Households.GetEntrance(ctx, event.DomophoneId, event.Door)
	if err != nil {
		s.logger.Error("Failed to get entrance", "event", event)
		return false
	}

	// Entrance not usage camera
	if entrance.CameraID == nil {
		s.logger.Debug("Entrance not usage camera, set PREVIEW mode 0")
		preview = PREVIEW_NONE
		imageGUIDv4 = IMAGE_UUID_STUB
	}

	// camera found
	if entrance.CameraID != nil {
		camera, err := s.repo.Cameras.GetCamera(context.Background(), *entrance.CameraID)
		if err != nil {
			s.logger.Error("Failed to get camera")
			return false
		}

		// 01 - get screenshot from domophone camera
		imgURL := API_URL_RBT + "/frs/camshot/" + strconv.Itoa(camera.CameraID)
		s.logger.Debug("Open by RFID", "_imageUrl >>", imgURL)

		var camScreenShot []byte
		camScreenShot, err = utils.DownloadFile(imgURL)
		if err != nil {
			preview = PREVIEW_NONE
			s.logger.Warn("Failed to get image from camera API, set preview mode 0", "err", err)
		}

		// check best quality image from FRS service
		bqResponse, _ := utils.GetBestQuality(s.frsApi, *entrance.CameraID, time.Now())
		if bqResponse != nil {
			camScreenShot, err = utils.DownloadFile(bqResponse.Data.Screenshot)
			if err != nil {
				s.logger.Error("Failed to download screenshot")
				return false
			}

			preview = PREVIEW_FRS

			faceData = map[string]interface{}{
				"left":   bqResponse.Data.Left,
				"top":    bqResponse.Data.Top,
				"width":  bqResponse.Data.Width,
				"height": bqResponse.Data.Height,
			}
		} else {
			s.logger.Debug("NO image from FRS service")
		}

		metadata := map[string]interface{}{
			"contentType": "image/jpeg",
			"expire":      int32(time.Unix(event.Date, 0).Add(TTL_CAMSHOT_HOURS).Unix()),
		}

		// save data to MongoDb
		fileId, err := s.fsFiles.SaveFile(MONGO_SCREENSHOT_NAME, metadata, camScreenShot)
		if err != nil {
			s.logger.Debug("MongoDB SaveFile", "err", err)
		}

		// generate image_uuid
		imageGUIDv4 = utils.ToGUIDv4(fileId)
	}

	flatList, err := s.repo.Households.FlatIDsByDomophoneIDAndPhone(ctx, event.DomophoneId, event.Detail)
	if err != nil {
		s.logger.Debug("Failed to get flatIDs", "err", err)
		return false
	}

	for _, flatID := range flatList {
		eventGUIDv4 := uuid.New().String()
		plogData := map[string]interface{}{
			"date":       event.Date,
			"event_uuid": eventGUIDv4,
			"hidden":     0,
			"image_uuid": imageGUIDv4,
			"flat_id":    flatID,
			"domophone": map[string]interface{}{
				"camera_id":             *entrance.CameraID,
				"domophone_description": entrance.Entrance,
				"domophone_id":          event.DomophoneId,
				"domophone_output":      entrance.DomophoneOutput,
				"entrance_id":           entrance.HouseEntranceID,
				"house_id":              entrance.AddressHouseID,
			},
			"event":  EVENT_OPENED_BY_APP,
			"opened": 1, // bool
			"face":   faceData,
			"rfid":   "",
			"code":   "",
			"phones": map[string]interface{}{
				"user_phone": event.Detail,
			},
			"preview": preview, // 0 no image, 1 - image from DVR, 2 - image from FRS
		}

		plogDataString, err := json.Marshal(plogData)
		if err != nil {
			s.logger.Debug("Failed marshal JSON")
		}

		err = s.storage.Insert("plog", string(plogDataString))
		if err != nil {
			s.logger.Error("Error insert to plog", "err", err)
		}
	}

	return true
}

// processOpenByFRS - process events open by FRS service
func (s *StreamProcessor) processOpenByFRS(ctx context.Context, event DoorOpenEvent) bool {
	var faceId, frsEventId string
	var faceData map[string]interface{}
	door := DOOR_MAIN

	if event.EventType == EVENT_OPENED_BY_FACE {
		eventDetail := strings.Split(event.Detail, "|")
		if len(eventDetail) == 2 {
			faceId = eventDetail[0]
			frsEventId = eventDetail[1]
		}
	} else {
		s.logger.Warn("event type err")
		return false
	}

	eventGUIDv4 := uuid.New().String()

	// get entrance
	entrance, err := s.repo.Households.GetEntrance(ctx, event.DomophoneId, door)
	if err != nil {
		s.logger.Error("Failed to get entrance")
		return false
	}

	// ----- get home data
	house, err := s.repo.Households.GetHouseByEntranceID(ctx, entrance.HouseEntranceID)
	if err != nil {
		s.logger.Warn("Failed to get house", "error", err)
	}

	camera, err := s.repo.Cameras.GetCamera(ctx, *entrance.CameraID)
	if err != nil {
		s.logger.Error("Failed to get camera")
		return false
	}

	// get screenShot from FRS service
	var camScreenShot []byte
	bqResponse, _ := utils.GetBestQualityByEvent(s.frsApi, *entrance.CameraID, frsEventId)
	if bqResponse != nil {
		camScreenShot, err = utils.DownloadFile(bqResponse.Data.Screenshot)
		faceData = map[string]interface{}{
			"left":   bqResponse.Data.Left,
			"top":    bqResponse.Data.Top,
			"width":  bqResponse.Data.Width,
			"height": bqResponse.Data.Height,
		}
	}

	// push crutch
	// hash for push event
	hash := fmt.Sprintf("%x", md5.Sum([]byte(uuid.New().String())))
	shotKey := "shot_" + hash
	if err := s.redis.Client.SetEx(ctx, shotKey, camScreenShot, 15*60*time.Second).Err(); err != nil {
		s.logger.Debug("failed to save screenshot to Redis", "err", err)
	}

	metadata := map[string]interface{}{
		"contentType": "image/jpeg",
		"expire":      int32(time.Unix(event.Date, 0).Add(TTL_CAMSHOT_HOURS).Unix()),
	}

	// save data to MongoDb
	fileId, err := s.fsFiles.SaveFile(MONGO_SCREENSHOT_NAME, metadata, camScreenShot)
	if err != nil {
		s.logger.Debug("MongoDB SaveFile", "err", err)
	}

	// generate image_uuid
	imageGUIDv4 := utils.ToGUIDv4(fileId)

	flatList, _ := s.repo.Households.GetFlatsByFaceIdFrs(ctx, faceId, strconv.Itoa(event.DomophoneId))

	// push crutch
	flatDetail, err := s.repo.Households.GetFlatByID(ctx, flatList[0])
	if err != nil {
		s.logger.Debug("Failed to get flat", "err", err)
		return false
	}

	plogData := map[string]interface{}{
		"date":       event.Date,
		"event_uuid": eventGUIDv4,
		"hidden":     0,
		"image_uuid": imageGUIDv4,
		"flat_id":    flatList[0],
		"domophone": map[string]interface{}{
			"camera_id":             camera.CameraID,
			"domophone_description": entrance.Entrance,
			"domophone_id":          event.DomophoneId,
			"domophone_output":      entrance.DomophoneOutput,
			"entrance_id":           entrance.HouseEntranceID,
			"house_id":              entrance.AddressHouseID,
		},
		"event":   EVENT_OPENED_BY_FACE,
		"opened":  1, // bool
		"face":    faceData,
		"rfid":    "",
		"code":    "",
		"phones":  map[string]interface{}{},
		"preview": PREVIEW_FRS, // 0 no image, 1 - image from DVR, 2 - image from FRS
	}
	plogDataString, err := json.Marshal(plogData)
	if err != nil {
		s.logger.Debug("Failed marshal JSON")
	}

	err = s.storage.Insert("plog", string(plogDataString))
	if err != nil {
		s.logger.Error("Error insert to plog", "err", err)
	}

	// FIXME: push crutch
	// get watchers
	watchers, err := s.repo.Households.GetWatchersByFlatID(ctx, flatDetail.HouseFlatID)
	if err != nil {
		s.logger.Warn("Failed to get watchers", "error", err)
	}

	if watchers != nil && len(watchers) > 0 {
		for _, watcher := range watchers {
			// watch FACE EVENT
			if watcher.EventType == strconv.Itoa(EVENT_OPENED_BY_FACE) {
				msgTitle := "Открытие двери"
				addressStr := house.HouseFull + "кв." + flatDetail.Flat
				msgBody := fmt.Sprintf("Адрес: %s\nПерсона: %s\n", addressStr, "имя жильца")
				device, _ := s.repo.Households.GetMobileDeviceByID(ctx, watcher.DeviceID)

				go utils.SendPush(hash, msgTitle, msgBody, device.PushToken, device.PushTokenType, device.Platform)
			}
		}
	} else {
		s.logger.Debug("Watchers not found")
	}

	return true
}

// pendingWorker - process failed tasks
func (s *StreamProcessor) pendingWorker(ctx context.Context) {
	defer s.wg.Done()

	workerName := "pending_recovery"
	s.logger.Info("Pending worker started", "worker", workerName)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Pending worker stopping", "worker", workerName)
			return
		default:
			// process pending tasks
			messages, _, err := s.redis.Client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
				Stream:   s.config.StreamName,
				Group:    s.config.GroupName,
				Consumer: workerName,
				MinIdle:  s.config.PendingMinIdle,
				Count:    10,
				Start:    "0-0",
			}).Result()

			if err != nil {
				s.logger.Error("Pending worker error",
					"worker", workerName,
					"error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if len(messages) > 0 {
				s.logger.Info("Pending worker processing messages",
					"worker", workerName,
					"count", len(messages))

				for _, msg := range messages {
					if s.processEvent(ctx, msg, workerName) {
						s.redis.Client.XAck(ctx, s.config.StreamName, s.config.GroupName, msg.ID)
					}
				}
			}

			time.Sleep(10 * time.Second)
		}
	}
}
