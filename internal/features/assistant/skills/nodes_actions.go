package skills

import (
	"ekken/internal/features/workflow/node"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

type NodesActions struct{}

func (s *NodesActions) GetID() string   { return "nodes_actions" }
func (s *NodesActions) GetName() string { return "Get Node Actions" }
func (s *NodesActions) GetDescription() string {
	return "Retrieve the full detailed configuration (including fields) for specific actions within nodes."
}

func (s *NodesActions) Execute(args map[string]interface{}) (string, error) {
	requestedMap, err := parseRequestedNodeActions(args)
	if err != nil {
		return "", err
	}

	catalog, err := fetchNodeCatalog()
	if err != nil {
		return "", err
	}

	filtered := filterNodeActions(catalog, requestedMap)
	if len(filtered) == 0 {
		return "", fmt.Errorf("node action not found in catalog")
	}

	res, err := yaml.Marshal(filtered)
	if err != nil {
		return "", fmt.Errorf("failed to format node actions: %v", err)
	}
	return string(res), nil
}

func parseRequestedNodeActions(args map[string]interface{}) (map[string][]string, error) {
	actionsRaw, ok := args["actions"]
	if !ok {
		return nil, fmt.Errorf("actions must be an array of strings")
	}

	var actions []string
	switch values := actionsRaw.(type) {
	case []interface{}:
		for _, raw := range values {
			str, ok := raw.(string)
			if !ok {
				return nil, fmt.Errorf("actions must be an array of strings")
			}
			actions = append(actions, str)
		}
	case []string:
		actions = append(actions, values...)
	default:
		return nil, fmt.Errorf("actions must be an array of strings")
	}

	// nodeType -> list of action types
	requestedMap := make(map[string][]string)
	for _, str := range actions {
		str = strings.TrimSpace(str)
		if str == "" {
			continue
		}

		idx := strings.LastIndex(str, ".")
		if idx == -1 {
			requestedMap[str] = nil
			continue
		}
		if idx == 0 || idx == len(str)-1 {
			return nil, fmt.Errorf("format error: use node_type or node_type.action_type")
		}

		nodeType := str[:idx]
		actionType := str[idx+1:]
		requestedMap[nodeType] = append(requestedMap[nodeType], actionType)
	}

	if len(requestedMap) == 0 {
		return nil, fmt.Errorf("actions must contain at least one node_type or node_type.action_type")
	}

	return requestedMap, nil
}

func fetchNodeCatalog() ([]node.Spec, error) {
	host := os.Getenv("EKKENAPI_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("EKKENAPI_PORT")
	if port == "" {
		port = "11245"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:%s/api/nodes/catalog", host, port))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node catalog: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var rawResp struct {
		Data []node.Spec `json:"data"`
	}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %v", err)
	}

	return rawResp.Data, nil
}

func filterNodeActions(catalog []node.Spec, requestedMap map[string][]string) []map[string]interface{} {
	// Filter only the requested nodes and actions
	var filtered []map[string]interface{}
	for _, n := range catalog {
		nodeType := n.Type
		specificActions, requested := requestedMap[nodeType]
		if !requested {
			continue
		}

		cleanNode := map[string]interface{}{
			"type": nodeType,
		}

		if len(n.GlobalFields) > 0 {
			cleanNode["global_fields"] = cleanNodeFields(n.GlobalFields)
		}

		if len(n.OutputHandles) > 0 {
			cleanNode["output_handles"] = n.OutputHandles
		}

		if len(n.Actions) > 0 {
			var filteredActions []interface{}
			filterLookup := make(map[string]bool)
			for _, k := range specificActions {
				filterLookup[k] = true
			}

			for _, a := range n.Actions {
				actionType := a.Type
				if len(specificActions) > 0 && !filterLookup[actionType] {
					continue
				}

				filteredActions = append(filteredActions, cleanNodeAction(a))
			}
			if len(filteredActions) > 0 {
				cleanNode["actions"] = filteredActions
			} else {
				continue
			}
		}

		filtered = append(filtered, cleanNode)
	}

	return filtered
}

func cleanNodeAction(action node.Action) map[string]interface{} {
	cleanAction := map[string]interface{}{
		"type": action.Type,
	}

	if action.Description != "" {
		cleanAction["description"] = action.Description
	}
	if action.HasResponse {
		cleanAction["has_response"] = true
		if action.ResponseType != nil {
			cleanAction["response_type"] = action.ResponseType
		}
		if action.ResponseVar != "" {
			cleanAction["response_var"] = action.ResponseVar
		}
	}
	if len(action.Fields) > 0 {
		cleanAction["fields"] = cleanNodeFields(action.Fields)
	}

	return cleanAction
}

func cleanNodeFields(fields []node.NodeField) []interface{} {
	var cleanFields []interface{}
	for _, field := range fields {
		cleanField := map[string]interface{}{
			"key":  field.Key,
			"type": field.Type,
		}
		if field.Required {
			cleanField["required"] = true
		}
		if field.Default != nil {
			cleanField["default"] = field.Default
		}
		if field.Options != nil {
			if optSlice, ok := field.Options.([]interface{}); ok {
				var simplifiedOpts []interface{}
				hasValueKey := false
				for _, opt := range optSlice {
					if optMap, ok := opt.(map[string]interface{}); ok {
						if val, exists := optMap["value"]; exists {
							simplifiedOpts = append(simplifiedOpts, val)
							hasValueKey = true
						} else {
							simplifiedOpts = append(simplifiedOpts, opt)
						}
					} else {
						simplifiedOpts = append(simplifiedOpts, opt)
					}
				}
				if hasValueKey {
					cleanField["options"] = simplifiedOpts
				} else {
					cleanField["options"] = field.Options
				}
			} else {
				cleanField["options"] = field.Options
			}
		}

		cleanFields = append(cleanFields, cleanField)
	}
	return cleanFields
}

func init() {
	Register(&NodesActions{})
}
