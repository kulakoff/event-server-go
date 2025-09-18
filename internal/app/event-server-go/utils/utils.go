package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	BearerTokenExample = "fbc55e76-848f-417e-a0c8-809646a5a4f8"
)

type FRSFaceData struct {
	Screenshot string `json:"screenshot"`
	Left       int    `json:"left"`
	Top        int    `json:"top"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

type FRSBestQualityResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Data    *FRSFaceData `json:"data"`
}

func SendPostRequest(url string, headers map[string]string, payload interface{}) ([]byte, int, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("error marshalling payload: %v", err)
	}

	// make request
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, 0, fmt.Errorf("error creating request: %v", err)
	}

	// process headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// call request
	res, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing request: %v", err)
	}
	defer res.Body.Close()

	// process response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, fmt.Errorf("error reading response body: %v", err)
	}

	return body, res.StatusCode, nil
}

func GetBestQuality(streamId int, timestamp time.Time) (*FRSBestQualityResponse, error) {
	url := "http://rbt-demo.lanta.me:12345/frs/api/bestQuality"

	// make headers
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + BearerTokenExample,
	}

	// make payload
	payload := map[string]interface{}{
		"streamId": streamId,
		"date":     timestamp,
	}

	// call request
	response, statusCode, err := SendPostRequest(url, headers, payload)
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
		var bestQualityResp FRSBestQualityResponse
		err = json.Unmarshal(response, &bestQualityResp)
		if err != nil {
			fmt.Println("error decoding response", err)
			return nil, err
		}

		return &bestQualityResp, nil
	}

	return nil, fmt.Errorf("unexpected status code: %d", statusCode)
}

func DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading file: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	return body, nil
}

func SaveFile(fileName string, data []byte) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

// ToGUIDv4 - convert MongoDb ObjectId to GUIDv4 string
func ToGUIDv4(objectId string) string {
	// add prefix
	uuid := "10001000" + objectId

	// string to UUID v4
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		uuid[0:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:])
}

// FromGUIDv4 - convert GUIDv4 string to MongoDb ObjectId
func FromGUIDv4(guid string) (string, error) {
	uuid := strings.ReplaceAll(guid, "-", "")

	if len(uuid) != 32 {
		return "", fmt.Errorf("invalid GUID format")
	}

	return uuid[8:], nil
}

// ExtractRFIDKey - extract RFID key from syslog message
func ExtractRFIDKey(message string) string {
	rfidRegex := regexp.MustCompile(`\b([0-9A-Fa-f]{14})\b`)
	match := rfidRegex.FindStringSubmatch(message)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
