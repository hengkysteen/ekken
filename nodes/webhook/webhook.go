package webhook

import (
	"fmt"
	"strings"

	"ekken/internal/features/workflow/node"
)

type WebhookNode struct {
	action node.Action
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "webhook",
				Tags:        []string{"Trigger"},
				Label:       "Webhook",
				Icon:        "https://www.svgrepo.com/show/451444/webhook.svg",
				Description: "Receive HTTP events from external apps.",
			},
			DefaultAction: "on_request",
			Actions: []node.Action{
				{
					Type:         "on_request",
					Label:        "On Request",
					Description:  "Trigger workflow when an HTTP request is received.",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{Key: "webhook_id", Type: "string", Required: true, Label: "Webhook ID"},
						{Key: "methods", Type: "array", Default: []string{"POST"}, Label: "Allowed Methods"},
						{Key: "secret", Type: "string", Label: "Secret"},
						{Key: "secret_location", Type: "string", Default: "header", Label: "Secret Location"},
						{Key: "secret_key", Type: "string", Default: "x-ekken-secret", Label: "Secret Key"},
						{Key: "public_url", Type: "string", Label: "Public URL"},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "webhook_id", Component: "input", Flex: 12, Options: map[string]any{"placeholder": "my-webhook"}},
							{Key: "methods", Component: "select", Flex: 12, Options: map[string]any{"multiple": true, "options": []string{"GET", "POST", "PUT", "PATCH", "DELETE"}}},
						},
						{
							{Key: "secret", Component: "input", Flex: 10, Options: map[string]any{"placeholder": "optional"}},
							{Key: "secret_location", Component: "select", Flex: 6, Options: map[string]any{"options": []string{"header", "query"}}},
							{Key: "secret_key", Component: "input", Flex: 8, Options: map[string]any{"placeholder": "x-ekken-secret"}},
						},
						{
							{Key: "public_url", Component: "input", Flex: 24, Options: map[string]any{"placeholder": "Optional public tunnel URL"}},
						},
					},
				},
			},
			OutputHandles: []string{"success", "error"},
		},
		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &WebhookNode{action: action}
		},
	})
}

func (n *WebhookNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	if n.action.Type != "" && n.action.Type != "on_request" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("unknown action: %s", n.action.Type)
	}

	webhookID, _ := node.FieldValue(n.action, "webhook_id").(string)
	webhookID = strings.TrimSpace(webhookID)
	if webhookID == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("webhook id is required")
	}

	listener, cleanup, err := node.GlobalEventListeners.Register(webhookID, n.action, ctx.WorkflowID)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, err
	}
	defer cleanup()

	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	case payload := <-listener.Payload:
		return node.NodeExecutionResult{
			Handle:   "success",
			Response: payload,
			Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
		}, nil
	}
}
