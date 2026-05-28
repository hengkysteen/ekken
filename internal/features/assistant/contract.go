package assistant

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

// IProvider defines the interface that all AI providers must implement.
// It handles both metadata, configuration, and execution.
type IProvider interface {
	Info() ProviderType
	Configure(config map[string]string)

	// Chat is the main execution method. The model to use is specified within ChatRequest.
	// listener can be nil for non-streaming requests.
	Chat(ctx context.Context, req ChatRequest, listener StreamListener) (MessageContent, error)
}

// BaseProvider implements the common fields and boilerplate for IProvider.
type BaseProvider struct {
	ID           string
	Name         string
	Logo         string
	BaseURL      string
	OfficialURL  string
	ConfigFields []string
	Config       map[string]string // The resolved runtime config
	HTTPClient   *http.Client
}

func (b *BaseProvider) Info() ProviderType {
	return ProviderType{
		ID:           b.ID,
		Name:         b.Name,
		Logo:         b.Logo,
		BaseURL:      b.BaseURL,
		OfficialURL:  b.OfficialURL,
		ConfigFields: b.ConfigFields,
	}
}

func (b *BaseProvider) Configure(config map[string]string) {
	b.Config = config
	if b.HTTPClient == nil {
		b.HTTPClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     false,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableKeepAlives:     false,
			},
		}
	}
}


// Stream is a helper for providers to handle the unified streaming logic.
func (b *BaseProvider) Stream(ctx context.Context, body io.Reader, mapper ChunkMapper, listener StreamListener) (MessageContent, error) {
	return StreamHandler(ctx, body, mapper, listener)
}
