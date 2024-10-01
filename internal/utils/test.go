package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
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

func GetBestQuality(streamId int, timestamp time.Time) (*GetBestQualityResponse, error) {
	url := "http://rbt-demo.lanta.me:9051/api/bestQuality"

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
