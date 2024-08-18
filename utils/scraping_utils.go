package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ethanhosier/mia-backend-go/types"
)

func GetPageScreenshot(url string) (string, error) {

	resp, err := http.Get(ScreenshotUrl + "?url=" + url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response types.ScreenshotScraperResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.ScreenshotBase64, nil
}
