package currency

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ekken/internal/features/workflow/node"
)

type CurrencyNode struct {
	Action node.Action
}

type frankfurterResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "currency",
				Label:       "Currency",
				Icon:        "https://www.svgrepo.com/show/515736/currency-dollar.svg",
				Tags:        []string{"System"},
				Description: "Convert values between different currencies using Frankfurter API.",
			},

			DefaultAction: "convert",
			Actions: []node.Action{
				{
					Type:         "convert",
					Label:        "Convert",
					Description:  "Convert amount from one currency to another",
					HasResponse:  true,
					ResponseType: &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
					Fields: []node.NodeField{
						{
							Key:      "amount",
							Type:     "string",
							Required: true,
							Label:    "Amount",
							Default:  "1.0",
						},
						{
							Key:      "from",
							Type:     "string",
							Required: true,
							Label:    "From",
							Default:  "USD",
							Options:  currencyOptions(),
						},
						{
							Key:      "to",
							Type:     "string",
							Required: true,
							Label:    "To",
							Default:  "IDR",
							Options:  currencyOptions(),
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{Key: "amount", Component: "input", Flex: 12, Options: map[string]any{"placeholder": "e.g. 100 or {{my_amount}}"}},
							{Key: "from", Component: "select", Flex: 6},
							{Key: "to", Component: "select", Flex: 6},
						},
					},
				},
			},
			OutputHandles: []string{"success"},
		},

		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &CurrencyNode{Action: action}
		},
	})
}

func (n *CurrencyNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	default:
	}

	if n.Action.Type != "convert" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("unknown action: %s", n.Action.Type)
	}

	return n.executeConvert(ctx)
}

func (n *CurrencyNode) executeConvert(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	var amountStr string
	amountVal := node.FieldValue(n.Action, "amount")
	switch v := amountVal.(type) {
	case string:
		amountStr = strings.TrimSpace(node.ParseTemplate(v, ctx.Variables))
	case float64:
		amountStr = strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		amountStr = strconv.Itoa(v)
	case int64:
		amountStr = strconv.FormatInt(v, 10)
	default:
		if amountVal != nil {
			amountStr = fmt.Sprintf("%v", amountVal)
		}
	}

	fromRaw, _ := node.FieldValue(n.Action, "from").(string)
	toRaw, _ := node.FieldValue(n.Action, "to").(string)

	from := strings.ToUpper(strings.TrimSpace(node.ParseTemplate(fromRaw, ctx.Variables)))
	to := strings.ToUpper(strings.TrimSpace(node.ParseTemplate(toRaw, ctx.Variables)))

	if amountStr == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("amount is required")
	}
	if from == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("from currency code is required")
	}
	if to == "" {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("to currency code is required")
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("invalid amount: %w", err)
	}

	if from == to {
		return node.NodeExecutionResult{
			Handle:   "success",
			Response: amount,
			Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
		}, nil
	}

	params := url.Values{}
	params.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	params.Add("from", from)
	params.Add("to", to)

	apiUrl := "https://api.frankfurter.app/latest?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx.Context, "GET", apiUrl, nil)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("request to Frankfurter API failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		if err := json.Unmarshal(body, &errResp); err == nil {
			if msg, ok := errResp["message"].(string); ok {
				return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("Frankfurter API error: %s", msg)
			}
		}
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("Frankfurter API returned status %d", resp.StatusCode)
	}

	var data frankfurterResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	convertedValue, exists := data.Rates[to]
	if !exists {
		return node.NodeExecutionResult{Handle: "error"}, fmt.Errorf("converted rates not found in response for %s", to)
	}

	return node.NodeExecutionResult{
		Handle:   "success",
		Response: convertedValue,
		Type:     &node.NodeResponseType{Mime: "application/json", Charset: "utf-8"},
	}, nil
}

func currencyOptions() []map[string]string {
	return []map[string]string{
		{"label": "USD - US Dollar", "value": "USD"},
		{"label": "IDR - Indonesian Rupiah", "value": "IDR"},
		{"label": "EUR - Euro", "value": "EUR"},
		{"label": "SGD - Singapore Dollar", "value": "SGD"},
		{"label": "MYR - Malaysian Ringgit", "value": "MYR"},
		{"label": "GBP - British Pound", "value": "GBP"},
		{"label": "JPY - Japanese Yen", "value": "JPY"},
		{"label": "AUD - Australian Dollar", "value": "AUD"},
		{"label": "CAD - Canadian Dollar", "value": "CAD"},
		{"label": "CHF - Swiss Franc", "value": "CHF"},
		{"label": "CNY - Chinese Renminbi", "value": "CNY"},
		{"label": "HKD - Hong Kong Dollar", "value": "HKD"},
		{"label": "INR - Indian Rupee", "value": "INR"},
		{"label": "KRW - South Korean Won", "value": "KRW"},
		{"label": "PHP - Philippine Peso", "value": "PHP"},
		{"label": "THB - Thai Baht", "value": "THB"},
		{"label": "TRY - Turkish Lira", "value": "TRY"},
		{"label": "NZD - New Zealand Dollar", "value": "NZD"},
		{"label": "BRL - Brazilian Real", "value": "BRL"},
		{"label": "BGN - Bulgarian Lev", "value": "BGN"},
		{"label": "CZK - Czech Koruna", "value": "CZK"},
		{"label": "DKK - Danish Krone", "value": "DKK"},
		{"label": "HUF - Hungarian Forint", "value": "HUF"},
		{"label": "ILS - Israeli New Shekel", "value": "ILS"},
		{"label": "ISK - Icelandic Króna", "value": "ISK"},
		{"label": "MXN - Mexican Peso", "value": "MXN"},
		{"label": "NOK - Norwegian Krone", "value": "NOK"},
		{"label": "PLN - Polish Zloty", "value": "PLN"},
		{"label": "RON - Romanian Leu", "value": "RON"},
		{"label": "SEK - Swedish Krona", "value": "SEK"},
		{"label": "ZAR - South African Rand", "value": "ZAR"},
	}
}
