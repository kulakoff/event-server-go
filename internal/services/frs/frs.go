package frs

import (
	"encoding/json"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/utils"
	"time"
)

// FrsUrl FIXME: refactor url dummy
// const FrsUrl = "http://localhost:9051"
// example token
const (
	BearerTokenExample = "fbc55e76-848f-417e-a0c8-809646a5a4f8"
	SuccessCode        = 204
)

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

func GetBestQuality(streamId int, timestamp time.Time, frsUrl, bearerToken string) (*GetBestQualityResponse, error) {
	url := frsUrl + "/api/bestQuality"

	// make headers
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + bearerToken,
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

// TODO: add param bearer token to api call
// MotionDetection - send motion start or stop message to RRS service
func MotionDetection(streamId int, motionActive bool, frsUrl string) error {
	url := frsUrl + "motionDetection"
	var motion string
	if motionActive {
		motion = "t"
	} else {
		motion = "f"
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + BearerTokenExample,
	}

	payload := map[string]interface{}{
		"streamId": streamId,
		"start":    motion,
	}

	_, statusCode, err := utils.SendPostRequest(url, headers, payload)
	if err != nil {
		return fmt.Errorf("error sending request %w", err)
	}
	if statusCode != SuccessCode {
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}
	return nil
}
