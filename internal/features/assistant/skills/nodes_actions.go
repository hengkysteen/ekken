package skills

import (
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

func fetchNodeCatalog() ([]map[string]interface{}, error) {
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
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %v", err)
	}

	return rawResp.Data, nil
}

func filterNodeActions(catalog []map[string]interface{}, requestedMap map[string][]string) []map[string]interface{} {
	// Filter only the requested nodes and actions
	var filtered []map[string]interface{}
	for _, node := range catalog {
		nodeType, _ := node["type"].(string)
		specificActions, requested := requestedMap[nodeType]
		if !requested {
			continue
		}

		cleanNode := map[string]interface{}{
			"type": nodeType,
		}

		if globalFields, ok := node["global_fields"].([]interface{}); ok {
			cleanGlobalFields := cleanNodeFields(globalFields)
			if len(cleanGlobalFields) > 0 {
				cleanNode["global_fields"] = cleanGlobalFields
			}
		}

		if actions, ok := node["actions"].([]interface{}); ok {
			var filteredActions []interface{}
			filterLookup := make(map[string]bool)
			for _, k := range specificActions {
				filterLookup[k] = true
			}

			for _, a := range actions {
				action, ok := a.(map[string]interface{})
				if !ok {
					continue
				}

				actionType, _ := action["type"].(string)
				if len(specificActions) > 0 && !filterLookup[actionType] {
					continue
				}

				filteredActions = append(filteredActions, cleanNodeAction(action, actionType))
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

func cleanNodeAction(action map[string]interface{}, actionType string) map[string]interface{} {
	cleanAction := map[string]interface{}{
		"type": actionType,
	}

	if desc, ok := action["description"].(string); ok && desc != "" {
		cleanAction["description"] = desc
	}
	if hr, ok := action["has_response"].(bool); ok && hr {
		cleanAction["has_response"] = true
		if rv, ok := action["response_var"].(string); ok && rv != "" {
			cleanAction["response_var"] = rv
		}
	}
	if fields, ok := action["fields"].([]interface{}); ok {
		cleanFields := cleanNodeFields(fields)
		if len(cleanFields) > 0 {
			cleanAction["fields"] = cleanFields
		}
	}

	return cleanAction
}

func cleanNodeFields(fields []interface{}) []interface{} {
	var cleanFields []interface{}
	for _, field := range fields {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			continue
		}

		cleanField := map[string]interface{}{
			"key":  fieldMap["key"],
			"type": fieldMap["type"],
		}
		if req, ok := fieldMap["required"].(bool); ok && req {
			cleanField["required"] = true
		}
		if def, ok := fieldMap["default"]; ok && def != nil {
			cleanField["default"] = def
		}

		cleanFields = append(cleanFields, cleanField)
	}
	return cleanFields
}

func init() {
	Register(&NodesActions{})
}
