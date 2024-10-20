package frs

import (
	"encoding/json"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/utils"
	"time"
)

// FrsUrl FIXME: refactor url dummy
//const FrsUrl = "http://localhost:9051"

type GetBestQualityData struct {
	Height     int    `json:"height"`
	Top        int    `json:"top"`
	Left       int    `json:"left"`
	With       int    `json:"with"`
	Screenshot string `json:"screenshot"`
}

type GetBestQualityResponse struct {
	Code    int                `json:"code"`
	Name    string             `json:"name"`
	Message string             `json:"message"`
	Data    GetBestQualityData `json:"data"`
}

func GetBestQuality(streamId int, timestamp time.Time) (*GetBestQualityResponse, error) {
	url := FrsUrl + "/api/bestQuality"

	// make headers
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// make payload
	payload := map[string]interface{}{
		"streamId": streamId,
		"date":     timestamp.Format("2006-01-02 15:04:05"),
	}

	// call request
	response, statusCode, err := utils.SendPostRequest(url, headers, payload)
	if err != nil {
		return nil, fmt.Errorf("error sending request %w", err)
	}

	// handle 204 status
	if statusCode == 204 {
		fmt.Println("frame not found for the given timestamp")
		return nil, nil
	}

	// Handle successful response (status 200)
	if statusCode == 200 {
		var bestQualityResp GetBestQualityResponse
		err = json.Unmarshal(response, &bestQualityResp)
		if err != nil {
			fmt.Println("error decoding response", err)
			return nil, err
		}

		return &bestQualityResp, nil
	}

	return nil, fmt.Errorf("unexpected status code: %d", statusCode)
}

// MotionDetection - send motion start or stop message to RRS service
func MotionDetection(streamId int, motionActive bool, frsUrl string) error {
	url := frsUrl + "/api/motionDetection"
	var motion string
	if motionActive {
		motion = "t"
	} else {
		motion = "f"
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	payload := map[string]interface{}{
		"streamId": streamId,
		"motion":   motion,
	}

	_, statusCode, err := utils.SendPostRequest(url, headers, payload)
	if err != nil {
		return fmt.Errorf("error sending request %w", err)
	}
	if statusCode != 204 {
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}
	return nil
}
