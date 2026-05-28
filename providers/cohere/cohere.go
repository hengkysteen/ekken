package cohere

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ekken/internal/features/assistant"
	"ekken/providers/common"
)

type CohereProvider struct {
	assistant.BaseProvider
}

func init() {
	assistant.Register("cohere", func() assistant.IProvider {
		return &CohereProvider{
			BaseProvider: assistant.BaseProvider{
				ID:           "cohere",
				Name:         "Cohere",
				Logo:         "https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/cohere/default.svg",
				BaseURL:      "https://api.cohere.com/v2",
				OfficialURL:  "https://dashboard.cohere.com/api-keys",
				ConfigFields: []string{"API_KEY"},
			},
		}
	}, []assistant.ModelEntry{
		{Name: "Command A 03-2025", Origin: "command-a-03-2025", ContextWindow: 262144},
		{Name: "Command R plus 08-2024", Origin: "command-r-plus-08-2024", ContextWindow: 128000},
		{Name: "Aya Expanse 32b", Origin: "c4ai-aya-expanse-32b", ContextWindow: 128000},
	})
}

func (p *CohereProvider) Chat(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	if req.Stream {
		return p.ChatStream(ctx, req, listener)
	}

	apiKey := p.Config["API_KEY"]
	if apiKey == "" {
		return assistant.MessageContent{}, fmt.Errorf("Cohere is not configured (API_KEY missing)")
	}

	payload := p.mapPayload(req, false)
	jsonData, _ := json.Marshal(payload)

	apiReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return assistant.MessageContent{}, err
	}
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	apiReq.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(apiReq)
	if err != nil {
		return assistant.MessageContent{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return assistant.MessageContent{}, fmt.Errorf("Cohere API error (%d): %s", resp.StatusCode, string(body))
	}

	var cohereResp struct {
		Message struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"message"`
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &cohereResp); err != nil {
		return assistant.MessageContent{}, fmt.Errorf("failed to unmarshal cohere response: %w", err)
	}

	content := ""
	if len(cohereResp.Message.Content) > 0 {
		content = cohereResp.Message.Content[0].Text
	}

	return assistant.MessageContent{Role: "assistant", Content: content}, nil
}

func (p *CohereProvider) ChatStream(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	apiKey := p.Config["API_KEY"]
	if apiKey == "" {
		return assistant.MessageContent{}, fmt.Errorf("Cohere is not configured (API_KEY missing)")
	}

	payload := p.mapPayload(req, true)
	jsonData, _ := json.Marshal(payload)

	apiReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return assistant.MessageContent{}, err
	}
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	apiReq.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(apiReq)
	if err != nil {
		return assistant.MessageContent{}, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return assistant.MessageContent{}, fmt.Errorf("Cohere API error (%d): %s", resp.StatusCode, string(body))
	}

	thinkingIndexes := map[int]bool{}

	return p.Stream(ctx, resp.Body, func(line string) (*assistant.StreamChunk, error) {
		lineStr := strings.TrimSpace(line)
		if lineStr == "" || !strings.HasPrefix(lineStr, "data:") {
			return nil, nil
		}

		data := strings.TrimSpace(strings.TrimPrefix(lineStr, "data:"))

		var event struct {
			Type  string `json:"type"`
			Index int    `json:"index"`
			Delta struct {
				FinishReason string `json:"finish_reason"`
				Error        string `json:"error"`
				Message      struct {
					Content struct {
						Type     string `json:"type"`
						Text     string `json:"text"`
						Thinking string `json:"thinking"`
					} `json:"content"`
				} `json:"message"`
			} `json:"delta"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, nil
		}

		// fmt.Println("[cohere chunk]", event.Type, data)

		switch event.Type {
		case "content-start":
			if event.Delta.Message.Content.Type == "thinking" {
				thinkingIndexes[event.Index] = true
			}
		case "content-delta":
			if thinkingIndexes[event.Index] {
				return &assistant.StreamChunk{Reasoning: event.Delta.Message.Content.Thinking}, nil
			}
			return &assistant.StreamChunk{Content: event.Delta.Message.Content.Text}, nil
		case "tool-call-start":
			return nil, fmt.Errorf("model attempted a unknown call, please follow the platform instructions and respond directly")
		case "message-end":
			if event.Delta.FinishReason == "ERROR" {
				return nil, fmt.Errorf("cohere: %s", event.Delta.Error)
			}
			return &assistant.StreamChunk{Done: true}, nil
		}

		return nil, nil
	}, listener)
}

func (p *CohereProvider) mapPayload(req assistant.ChatRequest, stream bool) map[string]any {
	return map[string]any{
		"model":    req.Model,
		"messages": common.MapMessagesWithRoles(req, nil),
		"stream":   stream,
	}
}
