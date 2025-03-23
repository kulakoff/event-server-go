package backend

//import (
//	"github.com/kulakoff/event-server-go/internal/utils"
//	"log/slog"
//)
//
//type OpenDoorMsg struct {
//	Date   string `json:"date"`
//	IP     string `json:"IP"`
//	SubId  string `json:"subId"`
//	Event  int    `json:"event"`
//	Detail string `json:"detail"`
//}
//type Stream struct {
//	ID     int
//	UrlDVR string
//	UrlFRS string
//}
//
//func stub() map[string]interface{} {
//	return map[string]interface{}{
//		"streamId": 8,
//	}
//}
//
//func APICallToRBT(payload OpenDoorMsg) error {
//	//url := "http://172.28.0.2/internal/actions/openDoor"
//	url := "https://webhook.site/55437bdc-ee94-48d1-b295-22a9f164b610/openDoor"
//
//	headers := map[string]string{
//		"Content-Type": "application/json",
//	}
//
//	_, _, err := utils.SendPostRequest(url, headers, payload)
//	if err != nil {
//		return err
//	}
//
//	slog.Debug("Successfully sent OpenDoorMsg")
//	return nil
//}
//
//func GetStremByIp(ip string) (*Stream, error) {
//	//TODO implement this method
//
//	//FIXME: stub stream
//	if ip == "192.168.13.152" {
//		return &Stream{
//			ID:     8,
//			UrlDVR: "https://dvr-example.com/stream-name/index.m3u8",
//			UrlFRS: "http://localhost:9051",
//		}, nil
//	}
//	if ip == "192.168.88.25" {
//		return &Stream{
//			ID:     8,
//			UrlDVR: "https://dvr-example.com/stream-name/index.m3u8",
//			UrlFRS: "https://webhook.site/5c40e512-64c6-49b8-96d0-d6d028f8181f",
//		}, nil
//	}
//
//	return nil, nil
//}
//
//func GetFlatGyRFID(key string) (int, error) {
//	//TODO implement this method
//
//	//FIXME: stub flat
//	if key == "00000075BC01AD," {
//		return 20, nil
//	}
//
//	return 0, nil
//}
//
//func GetDomophone(ip string) (interface{}, error) {
//	/**
//	TODO: implement me
//	Example response
//	"domophone": map[string]interface{}{
//			"camera_id":             8,
//			"domophone_description": "✅ Подъезд Beward",
//			"domophone_id":          6,
//			"domophone_output":      0,
//			"entrance_id":           23,
//			"house_id":              11,
//		},
//	*/
//
//	return nil, nil
//}
