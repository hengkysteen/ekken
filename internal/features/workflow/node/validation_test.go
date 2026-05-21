package node

import (
	"strings"
	"testing"
)

func makeIndex(types ...string) map[string]Spec {
	index := make(map[string]Spec)
	for _, t := range types {
		index[t] = Spec{Meta: Meta{Type: t}}
	}
	return index
}

// Should pass when dependency node exists and is in the correct order.
func TestValidateNodes_DependsOn_NodePresent(t *testing.T) {
	index := makeIndex("chrome", "click")
	index["click"] = Spec{
		Meta: Meta{
			Type:      "click",
			DependsOn: []DependsOn{{Node: "chrome"}},
		},
	}

	nodes := []Node{
		{Meta: Meta{Type: "chrome"}, ID: "n1"},
		{Meta: Meta{Type: "click"}, ID: "n2"},
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
	index["click"] = Spec{
		Meta: Meta{
			Type:      "click",
			DependsOn: []DependsOn{{Node: "chrome"}},
		},
	}

	nodes := []Node{
		{Meta: Meta{Type: "click"}, ID: "n1"},
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
			Meta: Meta{
				Type:      "click",
				DependsOn: []DependsOn{{Node: "chrome"}},
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
	nodes := []Node{{Meta: Meta{Type: "unknown_type"}, ID: "n1"}}

	issues := make([]string, 0)
	ValidateNodes(nodes, "workflow.nodes", index, &issues)

	if len(issues) == 0 || !strings.Contains(issues[0], "type is unknown") {
		t.Errorf("expected type is unknown issue, got: %v", issues)
	}
}

// Should fail when action type doesn't exist in catalog.
func TestValidateNodes_UnknownAction(t *testing.T) {
	index := makeIndex("chrome")
	index["chrome"] = Spec{
		Meta:    Meta{Type: "chrome"},
		Actions: []Action{{Type: "launch"}},
	}
	nodes := []Node{{
		Meta:   Meta{Type: "chrome"},
		Action: Action{Type: "navigate"}, // Invalid action
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
	index["chrome"] = Spec{
		Meta: Meta{Type: "chrome"},
		Actions: []Action{
			{
				Type:   "launch",
				Fields: []NodeField{{Key: "url", Required: true}},
			},
		},
	}
	nodes := []Node{{
		Meta:   Meta{Type: "chrome"},
		Action: Action{Type: "launch"}, // Missing URL
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
	index["chrome"] = Spec{
		Meta: Meta{Type: "chrome"},
		Actions: []Action{
			{
				Type:   "launch",
				Fields: []NodeField{{Key: "url", Required: true}},
			},
		},
	}
	nodes := []Node{{
		Meta: Meta{Type: "chrome"},
		Action: Action{
			Type:   "launch",
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
	index["chrome"] = Spec{
		Meta: Meta{Type: "chrome"},
		Actions: []Action{
			{
				Type:   "launch",
				Fields: []NodeField{{Key: "timeout", Required: true, Default: 60}},
			},
		},
	}
	nodes := []Node{{
		Meta: Meta{Type: "chrome"},
		Action: Action{
			Type:   "launch",
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

// Should pass when dependency matches both node type and action type.
func TestValidateNodes_DependsOn_Success(t *testing.T) {
	index := makeIndex("chrome", "click")
	index["chrome"] = Spec{
		Meta:    Meta{Type: "chrome"},
		Actions: []Action{{Type: "launch"}},
	}
	index["click"] = Spec{
		Meta: Meta{
			Type:      "click",
			DependsOn: []DependsOn{{Node: "chrome", Action: "launch"}},
		},
		Actions: []Action{{Type: "click_btn"}},
	}

	nodes := []Node{
		{Meta: Meta{Type: "chrome"}, Action: Action{Type: "launch"}},
		{Meta: Meta{Type: "click"}, Action: Action{Type: "click_btn"}},
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
	index["chrome"] = Spec{
		Meta: Meta{Type: "chrome"},
		Actions: []Action{
			{
				Type:   "launch",
				Fields: []NodeField{{Key: "url", Required: true}},
			},
		},
	}
	nodes := []Node{{
		Meta: Meta{Type: "chrome"},
		Action: Action{
			Type:   "launch",
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
	index["chrome"] = Spec{
		Meta: Meta{Type: "chrome"},
		Actions: []Action{
			{
				Type:   "launch",
				Fields: []NodeField{{Key: "url", Required: true}},
			},
		},
	}
	nodes := []Node{{
		Meta: Meta{Type: "chrome"},
		Action: Action{
			Type: "launch",
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
	index["chrome"] = Spec{
		Meta: Meta{Type: "chrome"},
		Actions: []Action{
			{
				Type:   "launch",
				Fields: []NodeField{{Key: "url", Type: "string", Required: true}},
			},
		},
	}
	nodes := []Node{{
		Meta: Meta{Type: "chrome"},
		Action: Action{
			Type: "launch",
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
	index["chrome"] = Spec{
		Meta:    Meta{Type: "chrome"},
		Actions: []Action{{Type: "launch"}},
	}
	index["click"] = Spec{
		Meta: Meta{
			Type:      "click",
			DependsOn: []DependsOn{{Node: "chrome", Action: "launch"}},
		},
		Actions: []Action{{Type: "click_btn"}},
	}

	nodes := []Node{
		// Node that needs dependency comes FIRST (Array Index 0)
		{Meta: Meta{Type: "click"}, Action: Action{Type: "click_btn"}},
		// Node that satisfies dependency comes LATER (Array Index 1)
		{Meta: Meta{Type: "chrome"}, Action: Action{Type: "launch"}},
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
