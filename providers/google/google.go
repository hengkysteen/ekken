package google

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

type GoogleProvider struct {
	assistant.BaseProvider
}

func init() {
	assistant.Register("google", func() assistant.IProvider {
		return &GoogleProvider{
			BaseProvider: assistant.BaseProvider{
				ID:           "google",
				Name:         "Google",
				Logo:         "https://upload.wikimedia.org/wikipedia/commons/1/1d/Google_Gemini_icon_2025.svg",
				BaseURL:      "https://generativelanguage.googleapis.com/v1beta",
				OfficialURL:  "https://aistudio.google.com/app/apikey",
				ConfigFields: []string{"API_KEY"},
			},
		}
	}, []assistant.ModelEntry{
		{Name: "Gemini 3.5 Flash", Origin: "gemini-3.5-flash", ContextWindow: 1048576},
		{Name: "Gemini 3.1 Pro Preview", Origin: "gemini-3.1-pro-preview", ContextWindow: 1048576},
		{Name: "Gemini 3.1 Flash Lite", Origin: "gemini-3.1-flash-lite", ContextWindow: 1048576},
	})
}

func (p *GoogleProvider) Chat(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	if req.Stream {
		return p.ChatStream(ctx, req, listener)
	}

	apiKey := p.Config["API_KEY"]
	if apiKey == "" {
		return assistant.MessageContent{}, fmt.Errorf("Google AI Studio is not configured (API_KEY missing)")
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.BaseURL, req.Model, apiKey)

	payload := p.mapPayload(req)
	jsonData, _ := json.Marshal(payload)

	apiReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return assistant.MessageContent{}, err
	}
	apiReq.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(apiReq)
	if err != nil {
		return assistant.MessageContent{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return assistant.MessageContent{}, fmt.Errorf("Google API error (%d): %s", resp.StatusCode, string(body))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return assistant.MessageContent{}, fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	content := ""
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		content = geminiResp.Candidates[0].Content.Parts[0].Text
	}

	return assistant.MessageContent{Role: "assistant", Content: content}, nil
}

func (p *GoogleProvider) ChatStream(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	apiKey := p.Config["API_KEY"]
	if apiKey == "" {
		return assistant.MessageContent{}, fmt.Errorf("Google AI Studio is not configured (API_KEY missing)")
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", p.BaseURL, req.Model, apiKey)

	payload := p.mapPayload(req)
	jsonData, _ := json.Marshal(payload)

	apiReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return assistant.MessageContent{}, err
	}
	apiReq.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(apiReq)
	if err != nil {
		return assistant.MessageContent{}, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return assistant.MessageContent{}, fmt.Errorf("Google API error (%d): %s", resp.StatusCode, string(body))
	}

	return p.Stream(ctx, resp.Body, func(line string) (*assistant.StreamChunk, error) {
		lineStr := strings.TrimSpace(line)
		if lineStr == "" || strings.HasPrefix(lineStr, ":") {
			return nil, nil
		}
		if !strings.HasPrefix(lineStr, "data:") {
			return nil, nil
		}

		data := strings.TrimPrefix(lineStr, "data:")
		data = strings.TrimSpace(data)
		if strings.HasPrefix(data, "[") && strings.HasSuffix(data, "]") {
			data = data[1 : len(data)-1]
		}

		var geminiResp struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}

		if err := json.Unmarshal([]byte(data), &geminiResp); err != nil {
			return nil, nil
		}

		if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
			return &assistant.StreamChunk{
				Content: geminiResp.Candidates[0].Content.Parts[0].Text,
			}, nil
		}

		return nil, nil
	}, listener)
}

func (p *GoogleProvider) mapPayload(req assistant.ChatRequest) map[string]any {
	payload := map[string]any{
		"generationConfig": map[string]any{
			"maxOutputTokens": 4096,
		},
	}

	contents := make([]map[string]any, 0)
	var systemInstruction string

	mappedMessages := common.MapMessagesWithRoles(req, map[string]string{
		"assistant": "model",
	})

	for _, msg := range mappedMessages {
		role := msg["role"].(string)
		content := msg["content"].(string)

		if role == "system" {
			systemInstruction += content + "\n"
			continue
		}

		contents = append(contents, map[string]any{
			"role": role,
			"parts": []map[string]any{
				{"text": content},
			},
		})
	}

	if systemInstruction != "" {
		payload["system_instruction"] = map[string]any{
			"parts": []map[string]any{
				{"text": strings.TrimSpace(systemInstruction)},
			},
		}
	}

	payload["contents"] = contents
	return payload
}
