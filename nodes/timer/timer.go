package timer

import (
	"ekken/internal/features/workflow/node"
	"ekken/internal/logger"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type TimerNode struct {
	action node.Action
}

var (
	cronParser = cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
)

func init() {
	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec: node.Spec{
			Meta: node.Meta{
				Type:        "timer",
				Tags:        []string{"Trigger"},
				Label:       "Timer",
				Icon:        "https://www.svgrepo.com/show/522683/timer.svg",
				Description: "Workflow entry point (manual, interval, or cron)",
			},
			DefaultAction: "manual",
			Actions: []node.Action{
				{
					Type:        "manual",
					Label:       "Manual",
					Description: "Trigger workflow manually (runs once)",
					HasResponse: false,
					Fields:      []node.NodeField{},
					AutoLayout: [][]node.AutoLayout{
						{
							{
								Key:       "info_1",
								Component: "text",
								Flex:      24,
								Options: map[string]any{
									"text": "This workflow will only run when manually triggered via the Run button.",
								},
							},
						},
					},
				},
				{
					Type:        "interval",
					Label:       "Interval",
					Description: "Trigger workflow at regular intervals",
					HasResponse: false,
					Fields: []node.NodeField{
						{Key: "interval", Type: "number", Label: "Interval", Default: 10, Required: true},
						{Key: "count", Type: "number", Label: "Iterations", Default: 1, Required: true},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{
								Key:       "interval",
								Component: "number",
								Flex:      12,
								Options:   map[string]any{"min": 1, "helper": "Interval in seconds"},
							},

							{
								Key:       "count",
								Component: "number",
								Flex:      12,
								Options:   map[string]any{"min": 0, "helper": "How many times to repeat (0 = run forever)"},
							},
						},
					},
				},
				{
					Type:        "cron",
					Label:       "Cron",
					Description: "Trigger workflow based on cron expression",
					HasResponse: false,
					Fields: []node.NodeField{
						{Key: "cron", Type: "string", Required: true, Label: "Cron expression"},
						{Key: "count", Type: "number", Label: "Iterations", Default: 0, Required: true},
					},
					AutoLayout: [][]node.AutoLayout{
						{
							{
								Key:       "cron",
								Component: "input",
								Flex:      14,
								Options:   map[string]any{"helper": "Format: second minute hour day month weekday", "placeholder": "*/30 * * * * *"},
							},

							{
								Key:       "count",
								Component: "number",
								Flex:      10,
								Options:   map[string]any{"min": 0, "helper": "Total scheduled runs (0 = run forever)"},
							}},
					},
				},
			},

			Outputs: []node.HandleEdge{{Key: "success", Label: "Success", Tone: "success"}},
		},
		ExecutorFactory: func(action node.Action) node.NodeExecutor {
			return &TimerNode{action: action}
		},
	})
}

func (n *TimerNode) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	// Validasi dasar
	if count := node.FieldValue(n.action, "count"); count != nil {
		if f, ok := count.(float64); ok && f < 0 {
			return node.NodeExecutionResult{}, fmt.Errorf("count cannot be negative")
		}
	}
	if interval := node.FieldValue(n.action, "interval"); interval != nil {
		if f, ok := interval.(float64); ok && f < 0 {
			return node.NodeExecutionResult{}, fmt.Errorf("interval cannot be negative")
		}
	}

	count := 1
	if c, ok := node.FieldValue(n.action, "count").(float64); ok {
		count = int(c)
	}
	if count > 0 && ctx.Iteration >= count {
		return node.NodeExecutionResult{}, node.ErrWorkflowComplete
	}

	switch n.action.Type {
	case "interval":
		return n.executeInterval(ctx, count)
	case "cron":
		return n.executeCron(ctx, count)
	default:
		return n.executeManual(ctx)
	}
}

func (n *TimerNode) executeManual(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	if ctx.Iteration > 0 {
		return node.NodeExecutionResult{}, node.ErrWorkflowComplete
	}
	return node.NodeExecutionResult{Handle: "success"}, nil
}

func (n *TimerNode) executeInterval(ctx *node.NodeContext, count int) (node.NodeExecutionResult, error) {
	interval := 10
	if v, ok := node.FieldValue(n.action, "interval").(float64); ok {
		interval = int(v)
	}

	logger.Info(fmt.Sprintf("[Timer] Interval trigger: waiting %d seconds...", interval))
	timer := time.NewTimer(time.Duration(interval) * time.Second)
	return n.waitTimer(ctx, timer, count)
}

func (n *TimerNode) executeCron(ctx *node.NodeContext, count int) (node.NodeExecutionResult, error) {
	cronExpr, _ := node.FieldValue(n.action, "cron").(string)
	if cronExpr == "" {
		return node.NodeExecutionResult{}, fmt.Errorf("cron expression is required")
	}

	schedule, err := cronParser.Parse(cronExpr)
	if err != nil {
		return node.NodeExecutionResult{}, fmt.Errorf("invalid cron expression: %v", err)
	}

	next := schedule.Next(time.Now())
	duration := time.Until(next)
	logger.Info(fmt.Sprintf("[Timer] Cron trigger [%s]: next run at %v (waiting %v)", cronExpr, next, duration))

	timer := time.NewTimer(duration)
	return n.waitTimer(ctx, timer, count)
}

func (n *TimerNode) waitTimer(ctx *node.NodeContext, timer *time.Timer, count int) (node.NodeExecutionResult, error) {
	if count <= 0 || ctx.Iteration+1 < count {
		ctx.IsLooping = true
	}

	select {
	case <-timer.C:
		// Timer fired, channel already drained
		return node.NodeExecutionResult{Handle: "success"}, nil
	case <-ctx.Stop:
		// Stop timer & drain channel if already fired (anti-leak pattern)
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		return node.NodeExecutionResult{}, node.ErrNodeStopped
	}
}
