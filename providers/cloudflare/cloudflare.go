package cloudflare

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

type CloudflareProvider struct {
	assistant.BaseProvider
}

func init() {
	assistant.Register("cloudflare", func() assistant.IProvider {
		return &CloudflareProvider{
			BaseProvider: assistant.BaseProvider{
				ID:           "cloudflare",
				Name:         "Workers AI",
				Logo:         "https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/cloudflare-workers/default.svg",
				BaseURL:      "https://api.cloudflare.com/client/v4/accounts/%s/ai/v1",
				OfficialURL:  "https://dash.cloudflare.com/",
				ConfigFields: []string{"ACCOUNT_ID", "API_KEY"},
			},
		}
	}, []assistant.ModelEntry{
		{Name: "kimi-k2.6", Origin: "@cf/moonshotai/kimi-k2.6", ContextWindow: 262144},
		{Name: "kimi-k2.5", Origin: "@cf/moonshotai/kimi-k2.5", ContextWindow: 256000},
		{Name: "glm-4.7-flash", Origin: "@cf/zai-org/glm-4.7-flash", ContextWindow: 131072},
		{Name: "nemotron-3-120b-a12b", Origin: "@cf/nvidia/nemotron-3-120b-a12b", ContextWindow: 256000},
		{Name: "gemma-4-26b-a4b-it", Origin: "@cf/google/gemma-4-26b-a4b-it", ContextWindow: 256000},
	})
}

func (p *CloudflareProvider) Chat(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	accountID := p.Config["ACCOUNT_ID"]
	apiKey := p.Config["API_KEY"]

	if accountID == "" || apiKey == "" {
		return assistant.MessageContent{}, fmt.Errorf("Cloudflare is not configured (ACCOUNT_ID or API_KEY missing)")
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/v1/chat/completions", accountID)

	payload := common.NewOpenAIPayload(req)
	jsonData, _ := json.Marshal(payload)
	apiReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
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
		return assistant.MessageContent{}, fmt.Errorf("cloudflare API error (%d): %s", resp.StatusCode, string(body))
	}

	if req.Stream {
		return common.CommonChatStream(ctx, resp.Body, listener)
	}

	return common.CommonChatNonStream(resp.Body)
}
