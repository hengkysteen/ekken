package nvidia

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

type NvidiaProvider struct {
	assistant.BaseProvider
}

func init() {
	assistant.Register("nvidia", func() assistant.IProvider {
		return &NvidiaProvider{
			BaseProvider: assistant.BaseProvider{
				ID:           "nvidia",
				Name:         "NVIDIA NIM",
				Logo:         "https://www.nvidia.com/content/dam/en-zz/Solutions/about-nvidia/logo-and-brand/01-nvidia-logo-vert-500x200-2c50-p@2x.png",
				BaseURL:      "https://integrate.api.nvidia.com/v1",
				OfficialURL:  "https://build.nvidia.com/",
				ConfigFields: []string{"API_KEY"},
			},
		}
	}, []assistant.ModelEntry{

		{Name: "Step-3.5-Flash", Origin: "stepfun-ai/step-3.5-flash", ContextWindow: 128000},
		{Name: "gpt-oss-120b", Origin: "openai/gpt-oss-120b", ContextWindow: 131072},
		{Name: "Qwen3-Next-80B-A3B-Instruct", Origin: "qwen/qwen3-next-80b-a3b-instruct", ContextWindow: 256000},
	})
}

func (p *NvidiaProvider) Chat(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	apiKey := p.Config["API_KEY"]
	if apiKey == "" {
		return assistant.MessageContent{}, fmt.Errorf("NVIDIA NIM is not configured (API_KEY missing)")
	}

	payload := common.NewOpenAIPayload(req)
	payload["max_tokens"] = 4096

	jsonData, _ := json.Marshal(payload)
	url := p.BaseURL + "/chat/completions"

	apiReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return assistant.MessageContent{}, err
	}

	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("Accept", "application/json")

	resp, err := p.HTTPClient.Do(apiReq)
	if err != nil {
		return assistant.MessageContent{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return assistant.MessageContent{}, fmt.Errorf("NVIDIA API Error (%d): %s", resp.StatusCode, string(body))
	}

	if req.Stream {
		return common.CommonChatStream(ctx, resp.Body, listener)
	}

	return common.CommonChatNonStream(resp.Body)
}
