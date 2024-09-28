package test

import (
	"fmt"
	"github.com/kulakoff/event-server-go/internal/config"
	"github.com/kulakoff/event-server-go/internal/storage"
	"github.com/kulakoff/event-server-go/internal/utils"
	"log/slog"
	"os"
)

func Draft() {
	// mongodb init
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Info("app started")

	// load main config
	cfg, err := config.New("config.json")
	if err != nil {
		logger.Warn("Error loading config file", "error", err)
	}

	mongo, err := storage.NewMongoDb(logger, cfg.MongoDb)
	if err != nil {
		logger.Error("Error init MongoDB", "error", err)
		os.Exit(1)
	}

	//utils.GetBestQuality(8, "2024-09-27 16:26:23")
	response, err := utils.DownloadFile("https://en.opensuse.org/images/a/ab/Opensuse_lightray.png")
	if err != nil {
		fmt.Println("error ")
		return
	}
	metadata := map[string]interface{}{
		"contentType": "image/png",
		"expire":      1742307721,
	}

	fileId, err := mongo.SaveFile("camshot", metadata, response)
	if err != nil {
		logger.Warn("Error saving file", "error", err)
	}

	fmt.Println(fileId)

	//err = utils.SaveFile("screenshot.png", response)
	//if err != nil {
	//	fmt.Println("error save file", err)
	//	return
	//}

}
