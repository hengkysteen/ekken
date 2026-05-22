package skills

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"ekken/internal/features/workflow"
	"ekken/internal/features/workflow/node"

	"github.com/goccy/go-yaml"
)

// YamlToJSON converts a YAML workflow string to a validated JSON workflow string conformant to the core engine.
func YamlToJSON(yamlStr string) (string, error) {
	var args map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &args); err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	var wfJson workflow.Workflow
	if err := ConvertToInternal(args, &wfJson); err != nil {
		return "", err
	}

	if err := ValidateWorkflow(wfJson); err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(wfJson, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// ConvertToInternal converts a parsed YAML map representation to the core engine's Workflow struct.
// It maps fields into a loose structure and delegates all strict SOT validation to the backend engine.
func ConvertToInternal(args map[string]interface{}, out *workflow.Workflow) error {
	// 1. Map Name
	if name, ok := args["name"].(string); ok {
		out.Name = name
	}

	// 2. Map Nodes
	nodesRaw, ok := args["nodes"].([]interface{})
	if !ok {
		return fmt.Errorf("nodes must be an array")
	}

	for idx, nRaw := range nodesRaw {
		nMap, ok := nRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("node at index %d must be an object", idx)
		}

		id, ok := nMap["id"].(string)
		if !ok || id == "" {
			return fmt.Errorf("node at index %d is missing required string 'id' field", idx)
		}

		actionRaw, ok := nMap["action"]
		if !ok {
			return fmt.Errorf("node '%s' is missing required 'action' field", id)
		}
		actionStr, ok := actionRaw.(string)
		if !ok || actionStr == "" {
			return fmt.Errorf("node '%s' field 'action' must be a non-empty string, got %T", id, actionRaw)
		}

		// Split action (e.g., timer.manual -> timer, manual)
		parts := strings.Split(actionStr, ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid action format: %s", actionStr)
		}
		nodeType, actionType := parts[0], parts[1]

		// Build engine node structure loosely
		engineNode := node.Node{
			ID: id,
		}
		engineNode.Type = nodeType
		engineNode.Action.Type = actionType

		// Map ResponseVar (from user YAML input)
		if respVar, ok := nMap["response_var"].(string); ok && respVar != "" {
			engineNode.Action.ResponseVar = respVar
		} else if respVar, ok := nMap["responseVar"].(string); ok && respVar != "" {
			engineNode.Action.ResponseVar = respVar
		}

		// Map dynamic fields (everything that is not metadata)
		for key, val := range nMap {
			if key == "id" || key == "action" || key == "response_var" || key == "responseVar" {
				continue
			}
			engineNode.Action.Fields = append(engineNode.Action.Fields, node.NodeField{
				Key:   key,
				Value: val,
			})
		}

		out.Nodes = append(out.Nodes, engineNode)
	}

	// 3. Map Edges
	if edgesRaw, ok := args["edges"]; ok {
		edgesList, ok := edgesRaw.([]interface{})
		if !ok {
			return fmt.Errorf("edges must be an array")
		}
		for idx, eRaw := range edgesList {
			eMap, ok := eRaw.(map[string]interface{})
			if !ok {
				return fmt.Errorf("edge at index %d must be an object", idx)
			}
			source, ok := eMap["source"]
			if !ok || source == nil || fmt.Sprint(source) == "" {
				return fmt.Errorf("edge at index %d is missing required 'source' field", idx)
			}
			target, ok := eMap["target"]
			if !ok || target == nil || fmt.Sprint(target) == "" {
				return fmt.Errorf("edge at index %d is missing required 'target' field", idx)
			}

			sourceHandle := ""
			if sh, ok := eMap["sourceHandle"]; ok && sh != nil {
				sourceHandle = fmt.Sprint(sh)
			}

			out.Edges = append(out.Edges, node.Edge{
				Source:       fmt.Sprint(source),
				SourceHandle: sourceHandle,
				Target:       fmt.Sprint(target),
			})
		}
	}

	return nil
}

// ValidateWorkflow validates a Workflow struct against the backend validation API.
func ValidateWorkflow(wf workflow.Workflow) error {
	payload, err := json.Marshal(wf)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow for validation: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(GetValidationAPIURL("/api/workflows/validate"), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("validation API call failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode validation response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("workflow validation failed: %s", result.Error)
	}

	return nil
}

// GetValidationAPIURL returns the full URL for a validation API path.
func GetValidationAPIURL(path string) string {
	host := os.Getenv("EKKENAPI_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("EKKENAPI_PORT")
	if port == "" {
		port = "11245"
	}
	return fmt.Sprintf("http://%s:%s%s", host, port, path)
}
