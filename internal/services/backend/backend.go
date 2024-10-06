package backend

import (
	"github.com/kulakoff/event-server-go/internal/utils"
	"log/slog"
)

type OpenDoorMsg struct {
	Date   string `json:"date"`
	IP     string `json:"IP"`
	SubId  string `json:"subId"`
	Event  int    `json:"event"`
	Detail string `json:"detail"`
}

func stub() map[string]interface{} {
	return map[string]interface{}{
		"streamId": 8,
	}
}

func APICallToRBT(payload OpenDoorMsg) error {
	//url := "http://172.28.0.2/internal/actions/openDoor"
	url := "https://webhook.site/55437bdc-ee94-48d1-b295-22a9f164b610/openDoor"

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, _, err := utils.SendPostRequest(url, headers, payload)
	if err != nil {
		return err
	}

	slog.Debug("Successfully sent OpenDoorMsg")
	return nil
}

func draft() {

}
