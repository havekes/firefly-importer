package parser

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"firefly-importer/models"
)

type visionRequest struct {
	Model    string        `json:"model"`
	Messages []interface{} `json:"messages"`
}

type textMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type imageContent struct {
	Type     string            `json:"type"`
	ImageURL map[string]string `json:"image_url"`
}

type visionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ParseImage sends an image to a Vision API and extracts transaction data.
func ParseImage(r io.Reader, visionAPIURL, visionAPIKey, visionModel string) ([]models.Transaction, error) {
	if visionAPIURL == "" {
		return nil, errors.New("vision API URL is required")
	}

	// Read image into memory
	imageBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	// Construct OpenAI-compatible payload
	prompt := `Extract bank transactions from this image. Return ONLY a JSON array with objects containing: 
	"date" (YYYY-MM-DD), "description" (string), "amount" (float, absolute value), and "type" (string: "withdrawal" or "deposit").
	Do not include markdown blocks like ` + "```json" + ` or any other text.`

	payload := visionRequest{
		Model: visionModel,
		Messages: []interface{}{
			textMessage{
				Role: "user",
				Content: []interface{}{
					textContent{Type: "text", Text: prompt},
					imageContent{
						Type: "image_url",
						ImageURL: map[string]string{
							"url": fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
						},
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode vision payload: %w", err)
	}

	endpoint := strings.TrimRight(visionAPIURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create vision request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if visionAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+visionAPIKey)
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vision API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vision API returned non-200 status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var vResp visionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vResp); err != nil {
		return nil, fmt.Errorf("failed to decode vision response: %w", err)
	}

	if len(vResp.Choices) == 0 {
		return nil, errors.New("no content parsed by vision API")
	}

	contentStr := vResp.Choices[0].Message.Content

	var transactions []models.Transaction
	if err := json.Unmarshal([]byte(contentStr), &transactions); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from vision response: %w, raw content: %s", err, contentStr)
	}

	// Set status for all parsed
	for i := range transactions {
		transactions[i].Status = models.StatusPending
	}

	return transactions, nil
}
