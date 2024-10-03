package draft

import (
	"github.com/kulakoff/event-server-go/internal/utils"
	"log"
)

func GetIntercomService(panelType string, clickHouseClient *utils.ClickhouseClient) IntercomService {
	switch panelType {
	case "Beward":
		return &BewardService{}
	case "Qtech":
		return &QtechService{}
	default:
		log.Fatalf("Unknown panel type: %s", panelType)
		return nil
	}
}
