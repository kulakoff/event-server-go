package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kulakoff/event-server-go/internal/config"
	"github.com/kulakoff/event-server-go/internal/handlers"
	"github.com/kulakoff/event-server-go/internal/repository"
	"github.com/kulakoff/event-server-go/internal/storage"
	"github.com/kulakoff/event-server-go/internal/syslog_custom"
	"github.com/kulakoff/event-server-go/internal/utils"
	"log/slog"
	"os"
	"time"
)

func main() {
	startServer()
	//todo2()
}

// main logic
func startServer() {
	// TODO: added log level from ENV
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Info("app started")

	// load main config
	cfg, err := config.New("config.json")
	if err != nil {
		logger.Warn("Error loading config file", "error", err)
	}

	// clickhouse init
	chDsn := cfg.Clickhouse
	ch, err := storage.NewClickhouseHttpClient(logger, &chDsn)
	if err != nil {
		logger.Error("Error init Clickhouse", "error", err)
		os.Exit(1)
	}

	// mongodb init
	mongo, err := storage.NewMongoDb(logger, cfg.MongoDb)
	if err != nil {
		logger.Error("Error init MongoDB", "error", err)
		os.Exit(1)
	}

	// postgres init
	psqlStorage, err := storage.NewPSQLStorage(logger, cfg.Postgres)
	if err != nil {
		logger.Error("Error init PSQLStorage", "error", err)
		os.Exit(1)
	}
	defer psqlStorage.Close()

	// init postgres storage
	repo, err := repository.NewPostgresRepository(psqlStorage.DB, logger)

	//test
	dm_ID, _ := repo.Households.GetDomophoneByIP(context.Background(), "37.235.188.212")
	logger.Debug("test", "demophobeId", dm_ID)

	stream, _ := repo.Cameras.GetStreamByIP(context.Background(), "37.235.188.212")
	logger.Debug("test", "stream", stream)

	entrace, _ := repo.Households.GetEntrace(context.Background(), dm_ID)
	logger.Debug("test", "entrace", entrace)

	// load spam filter
	spamFilers, err := config.LoadSpamFilters("spamwords.json")
	if err != nil {
		logger.Warn("Error loading spam filters", "error", err)
	}

	// ----- Beward syslog_custom server
	bewardHandler := handlers.NewBewardHandler(logger, spamFilers.Beward, ch, mongo, repo)
	bewardServer := syslog_custom.New(cfg.Hw.Beward.Port, "Beward", logger, bewardHandler)
	go bewardServer.Start()

	// ----- Qtech syslog_custom server
	qtechHandler := handlers.NewQtechHandler(logger, spamFilers.Qtech, ch, mongo, repo)
	qtechServer := syslog_custom.New(cfg.Hw.Qtech.Port, "Qtech", logger, qtechHandler)
	go qtechServer.Start()

	// Block main thread
	select {}
}

// test
func todo() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Info("start app, todo")

	// load main config
	cfg, err := config.New("config.json")
	if err != nil {
		logger.Warn("Error loading config file", "error", err)
	}

	// clickhouse init
	chDsn := cfg.Clickhouse
	ch, err := storage.NewClickhouse(logger, &chDsn)
	if err != nil {
		logger.Error("Error init Clickhouse", "error", err)
		os.Exit(1)
	}

	// fake data
	flatId := 20
	now := time.Now()
	imageId := "66fd704d7d5e19d2a3901aab"
	imageIdGUIDv4 := utils.ToGUIDv4(imageId)
	eventGUIDv4 := uuid.New().String()

	testMessage := map[string]interface{}{
		"date":       int32(now.Unix()),
		"event_uuid": eventGUIDv4,
		"hidden":     0,
		"image_id":   imageIdGUIDv4,
		"flat_id":    flatId,
		"domophone": map[string]interface{}{
			"camera_id":             8,
			"domophone_description": "✅ Подъезд Beward",
			"domophone_id":          6,
			"domophone_output":      0,
			"entrance_id":           23,
			"house_id":              11,
		},
		"event":  5,
		"opened": 1,
		"face": map[string]interface{}{
			"faceId": "17",
			"height": 204,
			"left":   529,
			"top":    225,
			"width":  160,
		},
		"rfid":    "",
		"code":    "",
		"phones":  map[string]interface{}{"user_phone": ""},
		"preview": 2,
	}

	testMessageStr, err := json.Marshal(testMessage)
	if err != nil {
		logger.Error("failed parse json", err)
	}

	fmt.Println(string(testMessageStr))

	ch.InsertPlog(string(testMessageStr))
}

func todo2() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Info("start app, todo")

	// load main config
	cfg, err := config.New("config.json")
	if err != nil {
		logger.Warn("Error loading config file", "error", err)
	}

	// clickhouse init
	chClient, err := storage.NewClickhouseHttpClient(logger, &cfg.Clickhouse)
	if err != nil {
		logger.Error("Error init Clickhouse", "error", err)
		os.Exit(1)
	}

	syslogData := map[string]interface{}{
		"timestamp": int32(time.Now().Unix()),
		"ip":        "192.168.13.33",
		"sub_id":    "",
		"unit":      "beward",
		"msg":       "blabla",
	}
	syslogMsg, _ := json.Marshal(syslogData)

	chClient.Insert("syslog", string(syslogMsg))

	// fake data
	flatId := 20
	now := time.Now()
	imageId := "66fd704d7d5e19d2a3901aab"
	imageIdGUIDv4 := utils.ToGUIDv4(imageId)
	eventGUIDv4 := uuid.New().String()

	testMessage := map[string]interface{}{
		"date":       int32(now.Unix()),
		"event_uuid": eventGUIDv4,
		"hidden":     0,
		"image_id":   imageIdGUIDv4,
		"flat_id":    flatId,
		"domophone": map[string]interface{}{
			"camera_id":             8,
			"domophone_description": "✅ Подъезд Beward",
			"domophone_id":          6,
			"domophone_output":      0,
			"entrance_id":           23,
			"house_id":              11,
		},
		"event":  5,
		"opened": 1,
		"face": map[string]interface{}{
			"faceId": "17",
			"height": 204,
			"left":   529,
			"top":    225,
			"width":  160,
		},
		"rfid":    "",
		"code":    "",
		"phones":  map[string]interface{}{"user_phone": ""},
		"preview": 2,
	}
	testMessageJson, _ := json.Marshal(testMessage)

	err = chClient.Insert("plog", string(testMessageJson))
	if err != nil {
		fmt.Print(err)
		return
	}

}
