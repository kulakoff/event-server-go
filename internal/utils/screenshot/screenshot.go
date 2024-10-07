package screenshot

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func DownloadVideoScreenshot(urlStr, filepath string) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		return fmt.Errorf("error downloading video: %w", err)
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return nil
}

func ExtractFrame(videoFile, outputImage string) error {
	cmd := exec.Command("ffmpeg", "-i", videoFile, "-vf", "thumbnail", "-frames:v", "1", outputImage)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error extracting frame: %v", err)
	}
	return nil
}
