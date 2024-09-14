package services

import "log"

func GetIntercomService(panelType string) IntercomService {
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
