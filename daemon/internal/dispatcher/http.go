package dispatcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gourish-mokashi/watchdog/daemon/pkg/models"
)

func SendAlerts(alerts models.SecEvent, backendURL string) error {
	jsonData, err := json.Marshal(alerts)
	if err != nil {
		return fmt.Errorf("failed to pack JSON: %w", err)
	}

	req, err := http.NewRequest("POST", backendURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("backend responded with status: %s", resp.Status)
	}

	return nil

}
