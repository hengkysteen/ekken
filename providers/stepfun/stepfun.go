package stepfun

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

type StepFunProvider struct {
	assistant.BaseProvider
}

func init() {
	assistant.Register("stepfun", func() assistant.IProvider {
		return &StepFunProvider{
			BaseProvider: assistant.BaseProvider{
				ID:           "stepfun",
				Name:         "StepFun",
				Logo:         "https://cdn.jsdelivr.net/gh/glincker/thesvg@main/public/icons/stepfun/default.svg",
				BaseURL:      "https://api.stepfun.ai/v1",
				OfficialURL:  "https://platform.stepfun.ai",
				ConfigFields: []string{"API_KEY"},
			},
		}
	}, []assistant.ModelEntry{
		{Name: "Step 3.5 Flash", Origin: "step-3.5-flash", ContextWindow: 262144},
		{Name: "Step 3.7 Flash", Origin: "step-3.7-flash", ContextWindow: 262144},
	})
}

func (p *StepFunProvider) Chat(ctx context.Context, req assistant.ChatRequest, listener assistant.StreamListener) (assistant.MessageContent, error) {
	apiKey := p.Config["API_KEY"]
	if apiKey == "" {
		return assistant.MessageContent{}, fmt.Errorf("StepFun is not configured (API_KEY missing)")
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
		return assistant.MessageContent{}, fmt.Errorf("StepFun API Error (%d): %s", resp.StatusCode, string(body))
	}

	if req.Stream {
		return common.CommonChatStream(ctx, resp.Body, listener)
	}

	return common.CommonChatNonStream(resp.Body)
}
