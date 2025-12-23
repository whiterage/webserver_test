package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// payload for webhook
type WebhookPayload struct {
	UserID    string         `json:"user_id"`
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Incidents []IncidentInfo `json:"incidents"`
	CheckedAt time.Time      `json:"checked_at"`
}

type IncidentInfo struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Radius      float64 `json:"radius"`
}

// sender sends webhooks with retry mechanism
type Sender struct {
	client        *http.Client
	webhookURL    string
	retryAttempts int
	retryDelay    time.Duration
}

// new sender for webhooks
func NewSender(webhookURL string, retryAttempts int, retryDelay time.Duration) *Sender {
	return &Sender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		webhookURL:    webhookURL,
		retryAttempts: retryAttempts,
		retryDelay:    retryDelay,
	}
}

// send webhook with exponential backoff
func (s *Sender) Send(ctx context.Context, payload *WebhookPayload) error {
	var lastErr error

	for attempt := 0; attempt < s.retryAttempts; attempt++ {
		if attempt > 0 {
			delay := s.retryDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := s.sendRequest(ctx, payload)
		if err == nil {
			return nil // successfully sent
		}

		lastErr = err
		fmt.Printf("Webhook send attempt %d failed: %v\n", attempt+1, err)
	}

	return fmt.Errorf("webhook send failed after %d attempts: %w", s.retryAttempts, lastErr)
}

// otpravlyaem zayavku http zapros
func (s *Sender) sendRequest(ctx context.Context, payload *WebhookPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
