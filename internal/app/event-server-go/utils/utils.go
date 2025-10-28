package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/config"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const PUSH_SERVICE_URL = "https://isdn.lanta.me/isdn_api.php"
const PUSH_SERVICE_TOKEN = "7e381dfb9d293290d06f0b050b24a7b2"

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

func GetBestQuality(frsApi *config.FrsApi, streamId int, timestamp time.Time) (*FRSBestQualityResponse, error) {
	url := frsApi.URL + "/frs/api/bestQuality"

	// make headers
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + frsApi.Token,
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

func GetBestQualityByEvent(frsApi *config.FrsApi, streamId int, frsEventId string) (*FRSBestQualityResponse, error) {
	url := frsApi.URL + "/frs/api/bestQuality"

	// make headers
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + frsApi.Token,
	}

	// make payload
	payload := map[string]interface{}{
		"streamId": streamId,
		"eventId":  frsEventId,
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

// SendGetRequest -
func SendGetRequest(url string, headers map[string]string) (int, string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем заголовки если есть
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response: %w", err)
	}

	return resp.StatusCode, string(body), nil
}

// example push
func SendPush(hashImage, title, message, deviceToken string, tokenType, platform int) error {
	apiToken := os.Getenv("PUSH_SERVICE_TOKEN")
	if apiToken == "" {
		apiToken = PUSH_SERVICE_TOKEN // Provide a default value
	}
	var devicePlatform string
	if platform == 0 {
		devicePlatform = "android"
	}
	if platform == 1 {
		devicePlatform = "iphone"
	}
	baseURL := PUSH_SERVICE_URL
	queryParams := url.Values{}

	queryParams.Add("action", "push")
	queryParams.Add("secret", apiToken) // FIXME: isdn auth
	queryParams.Add("token", deviceToken)
	queryParams.Add("type", strconv.Itoa(tokenType))
	queryParams.Add("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	queryParams.Add("ttl", "30")
	queryParams.Add("platform", devicePlatform)
	queryParams.Add("title", title)
	queryParams.Add("msg", message)
	queryParams.Add("sound", "default")
	queryParams.Add("pushAction", "paranoid")
	queryParams.Add("hash", hashImage)

	fullURL := baseURL + "?" + queryParams.Encode()

	slog.Debug("sendPush", "url", fullURL)
	startTime := time.Now()
	statusCode, body, err := SendGetRequest(fullURL, nil)
	duration := time.Since(startTime)

	if err != nil {
		slog.Error("FAILED to send debug push", "error", err, "duration", duration)
		return err
	}

	response := strings.TrimSpace(body)
	slog.Info("✅ Push server response", "status", statusCode, "response", response, "duration", duration.Seconds())

	if statusCode != 200 || response != "success" {
		return fmt.Errorf("push failed: %s", response)
	}

	return nil
}
