package plog

import (
	"github.com/kulakoff/event-server-go/internal/utils"
	"github.com/kulakoff/event-server-go/internal/utils/screenshot"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type StreamData struct {
	Timestamp  time.Time `json:"timestamp"`
	StreamName string    `json:"stream_name"`
	StreamID   int       `json:"stream_id"`
	Url        string    `json:"url"`
}

func ProcessImage(data *StreamData) ([]byte, error) {
	/**
	TODO:
		- get image from FRS
		- if FRS not return image - get  camera screenshot from DVR
		- return nil if err
	*/
	frsResp, err := utils.GetBestQuality(data.StreamID, data.Timestamp)
	if err != nil {
		slog.Debug("FRS GetBestQuality", "err", err)
		return nil, err
	}

	if frsResp != nil {
		// GET screenshot from FRS
		var imageUrl string
		imageUrl = frsResp.Data.Screenshot
		imageUrl = strings.Replace(imageUrl, "localhost", "rbt-demo.lanta.me", -1) // FIXME

		// download file from FRS response
		screenShot, err := utils.DownloadFile(imageUrl)
		if err != nil {
			slog.Debug("FRS DownloadFile", "err", err)
			return nil, err
		}
		return screenShot, nil
	} else {
		// GET screenshot from DVR
		timestampStr := strconv.FormatInt(data.Timestamp.Unix(), 10)
		path := "./images/"
		videoFileName := path + data.StreamName + "_" + timestampStr + ".mp4"
		imageFileName := path + data.StreamName + "_" + timestampStr + ".jpeg"

		// TODO: fix url
		url := data.Url + timestampStr + ".mp4"

		err := screenshot.DownloadVideoScreenshot(url, videoFileName)
		if err != nil {
			slog.Debug("FRS DownloadVideoScreenshot", "err", err)
			return nil, err
		}

		err = screenshot.ExtractFrame(videoFileName, imageFileName)
		if err != nil {
			slog.Debug("FRS ExtractFrame", "err", err)
			return nil, err
		}

		file, err := os.ReadFile(imageFileName)

		return file, nil
	}
}
