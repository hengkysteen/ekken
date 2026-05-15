package node

import (
	"strings"
	"testing"
)

func makeIndex(types ...string) map[string]NodeSpec {
	index := make(map[string]NodeSpec)
	for _, t := range types {
		index[t] = NodeSpec{NodeMetadata: NodeMetadata{Type: t}}
	}
	return index
}

// Should pass when dependency node exists and is in the correct order.
func TestValidateNodes_DependsOn_NodePresent(t *testing.T) {
	index := makeIndex("chrome", "click")
	index["click"] = NodeSpec{
		NodeMetadata: NodeMetadata{
			Type:      "click",
			DependsOn: []NodeDependency{{Node: "chrome"}},
		},
	}

	nodes := []Node{
		{NodeMetadata: NodeMetadata{Type: "chrome"}, ID: "n1"},
		{NodeMetadata: NodeMetadata{Type: "click"}, ID: "n2"},
	}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) != 0 {
		t.Errorf("expected no issues, got: %v", issues)
	}
}

// Should fail when catalog dependency is missing in the workflow.
func TestValidateNodes_DependsOn_NodeMissing(t *testing.T) {
	index := makeIndex("click")
	index["click"] = NodeSpec{
		NodeMetadata: NodeMetadata{
			Type:      "click",
			DependsOn: []NodeDependency{{Node: "chrome"}},
		},
	}

	nodes := []Node{
		{NodeMetadata: NodeMetadata{Type: "click"}, ID: "n1"},
	}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) == 0 {
		t.Error("expected issue for missing dependsOn node, got none")
	}
}

// Should fail when custom node dependency is missing in the workflow.
func TestValidateNodes_DependsOn_NodeLevel(t *testing.T) {
	index := makeIndex("chrome", "click")

	nodes := []Node{
		{
			NodeMetadata: NodeMetadata{
				Type:      "click",
				DependsOn: []NodeDependency{{Node: "chrome"}},
			},
			ID: "n1",
		},
	}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) == 0 {
		t.Error("expected issue for missing dependsOn node at node level, got none")
	}
}

// Should fail when node type is empty.
func TestValidateNodes_MissingType(t *testing.T) {
	index := makeIndex("chrome")
	nodes := []Node{{ID: "n1"}} // No Type

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) == 0 || !strings.Contains(issues[0], "type is required") {
		t.Errorf("expected type is required issue, got: %v", issues)
	}
}

// Should fail when node type doesn't exist in catalog.
func TestValidateNodes_UnknownType(t *testing.T) {
	index := makeIndex("chrome")
	nodes := []Node{{NodeMetadata: NodeMetadata{Type: "unknown_type"}, ID: "n1"}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) == 0 || !strings.Contains(issues[0], "type is unknown") {
		t.Errorf("expected type is unknown issue, got: %v", issues)
	}
}

// Should fail when action key doesn't exist in catalog.
func TestValidateNodes_UnknownAction(t *testing.T) {
	index := makeIndex("chrome")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions:      []NodeAction{{Key: "launch"}},
	}
	nodes := []Node{{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Action:       NodeAction{Key: "navigate"}, // Invalid action
	}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) == 0 || !strings.Contains(issues[0], "action \"navigate\" is not valid") {
		t.Errorf("expected invalid action issue, got: %v", issues)
	}
}

// Should fail when a required field is completely missing.
func TestValidateNodes_MissingRequiredField(t *testing.T) {
	index := makeIndex("chrome")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions: []NodeAction{
			{
				Key:    "launch",
				Fields: []NodeField{{Key: "url", Required: true}},
			},
		},
	}
	nodes := []Node{{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Action:       NodeAction{Key: "launch"}, // Missing URL
	}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) == 0 || !strings.Contains(issues[0], "field \"url\" is required") {
		t.Errorf("expected missing required field issue, got: %v", issues)
	}
}

// Should pass when required field is provided with a valid value.
func TestValidateNodes_ValidRequiredField(t *testing.T) {
	index := makeIndex("chrome")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions: []NodeAction{
			{
				Key:    "launch",
				Fields: []NodeField{{Key: "url", Required: true}},
			},
		},
	}
	nodes := []Node{{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Action: NodeAction{
			Key:    "launch",
			Fields: []NodeField{{Key: "url", Value: "https://example.com"}},
		},
	}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) != 0 {
		t.Errorf("expected no issues, got: %v", issues)
	}
}

// Should automatically fill empty field with default catalog value.
func TestValidateNodes_DefaultValuePopulation(t *testing.T) {
	index := makeIndex("chrome")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions: []NodeAction{
			{
				Key:    "launch",
				Fields: []NodeField{{Key: "timeout", Required: true, Default: 60}},
			},
		},
	}
	nodes := []Node{{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Action: NodeAction{
			Key:    "launch",
			Fields: []NodeField{{Key: "timeout", Required: true, Value: nil}}, // Empty value
		},
	}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) != 0 {
		t.Errorf("expected no issues due to default population, got: %v", issues)
	}

	// Verify that the field was populated with the default value
	if nodes[0].Action.Fields[0].Value != 60 {
		t.Errorf("expected field value to be populated with 60, got: %v", nodes[0].Action.Fields[0].Value)
	}
}

// Should pass when dependency matches both node type and action key.
func TestValidateNodes_DependsOn_Success(t *testing.T) {
	index := makeIndex("chrome", "click")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions:      []NodeAction{{Key: "launch"}},
	}
	index["click"] = NodeSpec{
		NodeMetadata: NodeMetadata{
			Type:      "click",
			DependsOn: []NodeDependency{{Node: "chrome", Action: "launch"}},
		},
		Actions: []NodeAction{{Key: "click_btn"}},
	}

	nodes := []Node{
		{NodeMetadata: NodeMetadata{Type: "chrome"}, Action: NodeAction{Key: "launch"}},
		{NodeMetadata: NodeMetadata{Type: "click"}, Action: NodeAction{Key: "click_btn"}},
	}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) != 0 {
		t.Errorf("expected no issues for satisfied dependency, got: %v", issues)
	}
}

// Should fail when a required field has an empty string value.
func TestValidateNodes_EmptyRequiredFieldValue(t *testing.T) {
	index := makeIndex("chrome")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions: []NodeAction{
			{
				Key:    "launch",
				Fields: []NodeField{{Key: "url", Required: true}},
			},
		},
	}
	nodes := []Node{{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Action: NodeAction{
			Key:    "launch",
			Fields: []NodeField{{Key: "url", Value: ""}}, // Value is empty string
		},
	}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	foundIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "field \"url\" is required") {
			foundIssue = true
			break
		}
	}

	if !foundIssue {
		t.Errorf("expected missing required field issue, got: %v", issues)
	}
}

// Should ignore client-provided schema flags and validate against catalog.
func TestValidateNodes_AlteredSchema_RequiredFalse(t *testing.T) {
	index := makeIndex("chrome")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions: []NodeAction{
			{
				Key:    "launch",
				Fields: []NodeField{{Key: "url", Required: true}},
			},
		},
	}
	nodes := []Node{{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Action: NodeAction{
			Key: "launch",
			// Client/LLM illegally manipulates the schema
			Fields: []NodeField{{Key: "url", Required: false, Value: "https://example.com"}},
		},
	}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) != 0 {
		t.Errorf("expected no issues because catalog schema is authoritative, got: %v", issues)
	}
}

// Should fail when payload contains a field that is not defined in the catalog.
func TestValidateNodes_UnknownField(t *testing.T) {
	index := makeIndex("chrome")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions: []NodeAction{
			{
				Key:    "launch",
				Fields: []NodeField{{Key: "url", Type: "string", Required: true}},
			},
		},
	}
	nodes := []Node{{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Action: NodeAction{
			Key: "launch",
			Fields: []NodeField{
				{Key: "url", Value: "https://example.com"},
				{Key: "unexpected", Value: "x"},
			},
		},
	}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	foundIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "unknown field") {
			foundIssue = true
			break
		}
	}
	if !foundIssue {
		t.Errorf("expected unknown field issue, got: %v", issues)
	}
}

// Should fail when dependency node is placed below the node that needs it.
func TestValidateNodes_DependsOn_PositionBelow(t *testing.T) {
	index := makeIndex("chrome", "click")
	index["chrome"] = NodeSpec{
		NodeMetadata: NodeMetadata{Type: "chrome"},
		Actions:      []NodeAction{{Key: "launch"}},
	}
	index["click"] = NodeSpec{
		NodeMetadata: NodeMetadata{
			Type:      "click",
			DependsOn: []NodeDependency{{Node: "chrome", Action: "launch"}},
		},
		Actions: []NodeAction{{Key: "click_btn"}},
	}

	nodes := []Node{
		// Node that needs dependency comes FIRST (Array Index 0)
		{NodeMetadata: NodeMetadata{Type: "click"}, Action: NodeAction{Key: "click_btn"}},
		// Node that satisfies dependency comes LATER (Array Index 1)
		{NodeMetadata: NodeMetadata{Type: "chrome"}, Action: NodeAction{Key: "launch"}},
	}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	foundIssue := false
	for _, issue := range issues {
		if strings.Contains(issue, "depends on Node \"chrome\" (\"launch\") to be executed before it") {
			foundIssue = true
			break
		}
	}

	if !foundIssue {
		t.Errorf("expected issue for dependency positioned below, got: %v", issues)
	}
}
