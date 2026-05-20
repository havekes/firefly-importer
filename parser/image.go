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
	"time"

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
func ParseImage(r io.Reader, fileDate, visionAPIURL, visionAPIKey, visionModel string) ([]models.Transaction, error) {
	if visionAPIURL == "" {
		return nil, errors.New("vision API URL is required")
	}

	// Read image into memory
	imageBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	currentDate := fileDate
	if currentDate == "" {
		currentDate = time.Now().Format("2006-01-02")
	} else {
		// Validate fileDate format to prevent prompt injection
		if _, err := time.Parse("2006-01-02", currentDate); err != nil {
			return nil, fmt.Errorf("invalid file date format (expected YYYY-MM-DD): %w", err)
		}
	}
	currentYear := currentDate[:4]

	// Construct OpenAI-compatible payload
	prompt := `Extract bank transactions from this image. Return ONLY a JSON array with objects containing:
	"date" (YYYY-MM-DD), "description" (string), "amount" (float, absolute value), and "type" (string: "withdrawal" or "deposit").
	Description should only contain transaction title, not the full transaction details.
	Assume the year is ` + currentYear + ` if not provided in the image.
	Today's date is ` + currentDate + `, use this to resolve relative dates like "today" or "yesterday".
	Do not include markdown blocks like ` + "```json" + `, ` + "```" + `, or any other text.`

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

	contentStr := cleanVisionResponse(vResp.Choices[0].Message.Content)

	var transactions []models.Transaction
	if err := json.Unmarshal([]byte(contentStr), &transactions); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from vision response: %w, raw content: %s", err, contentStr)
	}

	// Set status for all parsed and capture original description
	for i := range transactions {
		transactions[i].OriginalDescription = transactions[i].Description
		transactions[i].Status = models.StatusPending
	}

	return transactions, nil
}

// cleanVisionResponse removes thinking blocks and markdown formatting from the response.
func cleanVisionResponse(content string) string {
	// Remove <thought>...</thought> blocks (case-insensitive)
	for {
		lowerContent := strings.ToLower(content)
		startIdx := strings.Index(lowerContent, "<thought>")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(lowerContent[startIdx:], "</thought>")
		if endIdx == -1 {
			// If unclosed, strip everything from <thought> to the end
			content = content[:startIdx]
			break
		}
		// endIdx is relative to content[startIdx:]
		endIdx += startIdx + len("</thought>")
		content = content[:startIdx] + content[endIdx:]
	}

	// Remove <think>...</think> blocks (case-insensitive)
	for {
		lowerContent := strings.ToLower(content)
		startIdx := strings.Index(lowerContent, "<think>")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(lowerContent[startIdx:], "</think>")
		if endIdx == -1 {
			// If unclosed, strip everything from <think> to the end
			content = content[:startIdx]
			break
		}
		// endIdx is relative to content[startIdx:]
		endIdx += startIdx + len("</think>")
		content = content[:startIdx] + content[endIdx:]
	}

	// Remove markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		// Find first newline to skip ```json or ```
		if nlIdx := strings.Index(content, "\n"); nlIdx != -1 {
			content = content[nlIdx+1:]
		}
		// Remove trailing ```
		if strings.HasSuffix(content, "```") {
			content = content[:len(content)-3]
		}
	}
	return strings.TrimSpace(content)
}
