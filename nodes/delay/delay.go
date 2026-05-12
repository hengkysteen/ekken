package delay

import (
	"ekken/internal/features/workflow/node"
	"fmt"
	"time"
)

type DelayNode struct {
	Action node.NodeAction
	Output any
}

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		NodeSpec: node.NodeSpec{
			NodeMetadata: node.NodeMetadata{
				Type:        "delay",
				Label:       "Delay",
				Icon:        "https://www.svgrepo.com/show/86792/sand-clock.svg",
				Tags:        []string{"System"},
				Description: "Adds a pause before continuing the workflow.",
			},

			DefaultAction: "seconds",
			Actions: []node.NodeAction{
				{
					Key:         "seconds",
					Label:       "Seconds",
					Description: "Pause execution for specified duration",
					Fields: []node.NodeField{
						{
							Key:      "duration",
							Type:     "number",
							Required: true,
							Label:    "Duration in seconds",
							Default:  1.0,
							Options:  map[string]any{"min": 0},
						},
					},
					AutoLayout: [][]node.AutoLayout{
						{{Key: "duration", Component: "number-s1"}},
					},
				},
			},
			Outputs: []node.HandleEdge{{Key: "success", Label: "Success"}}},

		ExecutorFactory: func(action node.NodeAction) node.NodeExecutor {
			return &DelayNode{Action: action}
		},
	})
}
func (n *DelayNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	seconds, ok := node.FieldValue(n.Action, "duration").(float64)
	if !ok {
		return node.NodeExecutionResult{}, fmt.Errorf("invalid duration format")
	}
	if seconds < 0 {
		return node.NodeExecutionResult{}, fmt.Errorf("duration cannot be negative")
	}
	select {
	case <-ctx.Stop:
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	case <-time.After(time.Duration(seconds * float64(time.Second))):
		return node.NodeExecutionResult{Handle: "success"}, nil
	}
}
