package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// Payload defines the structure of the data sent to the webhook.

type Payload struct {
	ID        string   `json:"id"`
	Content   string   `json:"content"`
	UserTags  []string `json:"userTags"`
	Note      string   `json:"note,omitempty"`
	CreatedAt string   `json:"createdAt"`
}

// sendToWebhook creates the payload and sends it to the configured webhook URL.
func sendToWebhook(content, tags, note string) error {
	webhookURL := viper.GetString("webhookUrl")
	if webhookURL == "" {
		return fmt.Errorf(errWebhookNotConfigured)
	}

	payload := Payload{
		ID:        uuid.New().String(),
		Content:   content,
		UserTags:  parseTags(tags),
		Note:      note,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error creating JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken := viper.GetString("authToken"); authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf(errSavingSnippet)
	}

	fmt.Println("Successfully saved to DevStash!")
	return nil
}

// parseTags splits a comma-separated string of tags into a slice of strings.
func parseTags(tagsStr string) []string {
	if strings.TrimSpace(tagsStr) == "" {
		return []string{}
	}
	tags := strings.Split(tagsStr, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}
