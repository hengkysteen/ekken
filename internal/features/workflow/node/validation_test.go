package node

import (
	"testing"
)

func makeIndex(types ...string) map[string]NodeSpec {
	index := make(map[string]NodeSpec)
	for _, t := range types {
		index[t] = NodeSpec{NodeMetadata: NodeMetadata{Type: t}}
	}
	return index
}

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
