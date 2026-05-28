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
)

type NodesList struct{}

func (s *NodesList) GetID() string   { return "nodes" }
func (s *NodesList) GetName() string { return "Get Workflow Nodes List" }
func (s *NodesList) GetDescription() string {
	return "Retrieve the list of available automation blocks (Nodes) that can be used to build a workflow."
}

func (s *NodesList) Execute(args map[string]interface{}) (string, error) {
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
		return "", fmt.Errorf("failed to fetch node catalog: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Filter data for AI usage (save tokens)
	var rawResp struct {
		Data []node.Spec `json:"data"`
	}

	if err := json.Unmarshal(body, &rawResp); err != nil {
		return "", fmt.Errorf("failed to parse node catalog: %v", err)
	}

	var sb strings.Builder

	sb.WriteString(`
[INSTRUCTIONS]
1. This list is a simple index (Node & Action types only), NOT full field definitions.
2. CRITICAL: You MUST call the 'nodes_actions' skill to fetch the exact fields for each action before composing/modifying a workflow. NEVER assume or guess fields.
3. Only set 'response_var' for actions that have 'has_response: true' in their details.

Example:
~ekken skill nodes_actions
actions:
  - fs.write
  - shell.execute
ekken~


`)

	sb.WriteString("Available Nodes:\n")

	for _, n := range rawResp.Data {
		fmt.Fprintf(&sb, "- Type: %s\n", n.Type)
		fmt.Fprintf(&sb, "  Description: %s\n", n.Description)
		if len(n.DependsOn) > 0 {
			sb.WriteString("  Depends On:\n")
			for _, d := range n.DependsOn {
				fmt.Fprintf(&sb, "    - %s.%s\n", d.Node, d.Action)
			}
		}
		if len(n.Actions) > 0 {
			sb.WriteString("  Actions:\n")
			for _, a := range n.Actions {
				fmt.Fprintf(&sb, "    - %s: %s\n", a.Type, a.Description)
			}
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func init() {
	Register(&NodesList{})
}
