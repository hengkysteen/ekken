package skills

import (
	"ekken/internal/features/workflow"
	_ "ekken/nodes"
	"encoding/json"
	"testing"
)

func TestConvertToInternal_AutoFillResponseVar(t *testing.T) {
	yamlStr := `
name: "Test workflow"
nodes:
  - id: trigger
    action: timer.manual
  - id: run_shell
    action: shell.execute
    fields:
      command: echo "hello"
  - id: check_status
    action: ifelse.if_else
    fields:
      operand_1: "{{shell.execute_}}"
      operand_2: "hello"
      operator: contains
`
	jsonStr, err := YamlToJSON(yamlStr)
	if err != nil {
		t.Fatalf("failed to convert YAML to JSON: %v", err)
	}

	t.Logf("Generated JSON:\n%s", jsonStr)

	var wfJson struct {
		Nodes []struct {
			ID     string `json:"id"`
			Action struct {
				Type        string `json:"type"`
				ResponseVar string `json:"response_var"`
				Fields      []struct {
					Key   string `json:"key"`
					Value any    `json:"value"`
				} `json:"fields"`
			} `json:"action"`
		} `json:"nodes"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &wfJson); err != nil {
		t.Fatalf("failed to unmarshal generated JSON: %v", err)
	}

	if len(wfJson.Nodes) < 3 {
		t.Fatalf("expected nodes, got %d", len(wfJson.Nodes))
	}

	// Verify run_shell node has response_var and command field nested
	runShellNode := wfJson.Nodes[1]
	if runShellNode.ID != "run_shell" {
		t.Errorf("expected ID run_shell, got %s", runShellNode.ID)
	}
	if runShellNode.Action.ResponseVar != "shell.execute_" {
		t.Errorf("expected response_var 'shell.execute_', got '%s'", runShellNode.Action.ResponseVar)
	}
	if len(runShellNode.Action.Fields) == 0 {
		t.Fatalf("expected fields, got 0")
	}
	if runShellNode.Action.Fields[0].Key != "command" || runShellNode.Action.Fields[0].Value != "echo \"hello\"" {
		t.Errorf("expected field command: 'echo \"hello\"', got key '%s', value '%v'", runShellNode.Action.Fields[0].Key, runShellNode.Action.Fields[0].Value)
	}
}

func TestConvertToInternal_DefaultPositionsWrapAfterFiveNodes(t *testing.T) {
	args := map[string]interface{}{
		"name": "Position test",
		"nodes": []interface{}{
			map[string]interface{}{"id": "hbt71", "action": "fs.write"},
			map[string]interface{}{"id": "y7tuf", "action": "fs.write"},
			map[string]interface{}{"id": "pcdlo", "action": "fs.write"},
			map[string]interface{}{"id": "ajk51", "action": "fs.write"},
			map[string]interface{}{"id": "h5vke", "action": "fs.write"},
			map[string]interface{}{"id": "kuyi4", "action": "fs.write"},
		},
	}

	var wf workflow.Workflow
	if err := ConvertToInternal(args, &wf); err != nil {
		t.Fatalf("failed to convert workflow: %v", err)
	}

	if len(wf.Positions) != 6 {
		t.Fatalf("expected 6 positions, got %d", len(wf.Positions))
	}

	if got := wf.Positions["hbt71"]; got.X != 39 || got.Y != 84.5 {
		t.Fatalf("unexpected first node position: %#v", got)
	}
	if got := wf.Positions["h5vke"]; got.X != 1239 || got.Y != 84.5 {
		t.Fatalf("unexpected fifth node position: %#v", got)
	}
	if got := wf.Positions["kuyi4"]; got.X != 39 || got.Y != 304.5 {
		t.Fatalf("unexpected sixth node position: %#v", got)
	}
}
