package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ekken/internal/features/assistant"
	"ekken/providers/common"
)

type OllamaProvider struct {
	assistant.BaseProvider
}

func init() {
	assistant.Register("ollama", func() assistant.IProvider {
		return &OllamaProvider{
			BaseProvider: assistant.BaseProvider{
				ID:           "ollama",
				Name:         "Ollama",
				Logo:         "https://ollama.com/public/ollama.png",
				BaseURL:      "http://localhost:11434/v1",
				OfficialURL:  "ollama.com",
				ConfigFields: []string{"URL"},
			},
		}
	}, []assistant.ModelEntry{
		{Name: "qwen3.5:0.8b", Origin: "qwen3.5:0.8b", ContextWindow: 262144},
		{Name: "qwen3.5:cloud", Origin: "qwen3.5:cloud", ContextWindow: 262144},
		{Name: "qwen3:4b-instruct", Origin: "qwen3:4b-instruct", ContextWindow: 262144},
		{Name: "gemma4:31b-cloud", Origin: "gemma4:31b-cloud", ContextWindow: 262144},
		{Name: "qwen3.5:2b", Origin: "qwen3.5:2b", ContextWindow: 262144},
	})
}

func (p *OllamaProvider) Chat(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	baseURL := p.BaseURL
	if customURL := p.Config["URL"]; customURL != "" {
		baseURL = customURL
	}

	payload := common.NewOpenAIPayload(req)

	jsonData, _ := json.Marshal(payload)
	apiReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
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
		return assistant.MessageContent{}, fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, string(body))
	}

	if req.Stream {
		return common.CommonChatStream(ctx, resp.Body, listener)
	}

	return common.CommonChatNonStream(resp.Body)
}
