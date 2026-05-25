package currency

import (
	"context"
	"strings"
	"testing"

	"ekken/internal/features/workflow/node"
)

func TestCurrencyNode_ShortCircuit(t *testing.T) {
	n := &CurrencyNode{
		Action: node.Action{
			Type: "convert",
			Fields: []node.NodeField{
				{Key: "amount", Value: "150.50"},
				{Key: "from", Value: "usd"},
				{Key: "to", Value: "USD"},
			},
		},
	}

	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]any),
		Stop:      make(chan struct{}),
	}

	res, err := n.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Handle != "success" {
		t.Fatalf("expected success, got %s", res.Handle)
	}

	val, ok := res.Response.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", res.Response)
	}

	if val != 150.50 {
		t.Errorf("expected 150.50, got %f", val)
	}
}

func TestCurrencyNode_RealAPI(t *testing.T) {
	n := &CurrencyNode{
		Action: node.Action{
			Type: "convert",
			Fields: []node.NodeField{
				{Key: "amount", Value: "1.0"},
				{Key: "from", Value: "USD"},
				{Key: "to", Value: "EUR"},
			},
		},
	}

	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]any),
		Stop:      make(chan struct{}),
	}

	res, err := n.Execute(ctx)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "timeout") {
			t.Skipf("Skipping real API test due to network/internet issues: %v", err)
			return
		}
		t.Fatalf("failed to execute real API conversion: %v", err)
	}

	if res.Handle != "success" {
		t.Fatalf("expected success, got %s", res.Handle)
	}

	val, ok := res.Response.(float64)
	if !ok {
		t.Fatalf("expected float64 response, got %T", res.Response)
	}

	if val <= 0 {
		t.Errorf("expected conversion rate to be positive, got %f", val)
	}
}

func TestCurrencyNode_InvalidCode(t *testing.T) {
	n := &CurrencyNode{
		Action: node.Action{
			Type: "convert",
			Fields: []node.NodeField{
				{Key: "amount", Value: "10"},
				{Key: "from", Value: "USD"},
				{Key: "to", Value: "XYZ"},
			},
		},
	}

	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]any),
		Stop:      make(chan struct{}),
	}

	_, err := n.Execute(ctx)
	if err == nil {
		t.Fatalf("expected error for invalid currency code XYZ, got nil")
	}

	errStr := err.Error()
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "timeout") {
		t.Skip("Skipping test due to network issues")
		return
	}

	if !strings.Contains(err.Error(), "error") && !strings.Contains(err.Error(), "status 4") && !strings.Contains(err.Error(), "status 5") {
		t.Errorf("expected Frankfurter API rejection error, got: %v", err)
	}
}
