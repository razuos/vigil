package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HeartbeatRequest struct {
	Load   float64 `json:"load"`
	Uptime uint64  `json:"uptime"`
}

type HeartbeatResponse struct {
	Action string `json:"action"` // "sleep" or "stay_awake"
}

type WebControlClient struct {
	ControllerURL string
	Client        *http.Client
}

func NewWebControlClient(url string) *WebControlClient {
	return &WebControlClient{
		ControllerURL: url,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *WebControlClient) Report(ctx context.Context, loadVal float64) (string, error) {
	reqBody := HeartbeatRequest{
		Load: loadVal,
		// Uptime can be added later if needed
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal heartbeat: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.ControllerURL+"/api/heartbeat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("controller returned status: %d", resp.StatusCode)
	}

	var hbResp HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&hbResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return hbResp.Action, nil
}
