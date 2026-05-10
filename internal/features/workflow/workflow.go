package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ekken/internal/features/workflow/node"
)

type WorkflowObserver interface {
	OnStatusUpdate(id, status string, iteration int)
	OnLog(id, level, message, raw string)
}

type Runner struct {
	observer WorkflowObserver
	registry node.NodeProvider
}

func New(observer WorkflowObserver, registry node.NodeProvider) *Runner {
	return &Runner{
		observer: observer,
		registry: registry,
	}
}

func (e *Runner) Run(ctx context.Context, wf Workflow) error {
	wfCtx := node.NewRunnerContext(&node.NodeContext{
		Stop:              ctx.Done(),
		Context:           ctx,
		Variables:         make(map[string]interface{}),
		InternalVariables: make(map[string]interface{}),
		Metadata:          make(map[string]interface{}),
		OnCleanup:         make([]func(), 0),
		WorkflowID:        wf.ID,
	})

	defer func() {
		for _, cleanup := range wfCtx.OnCleanup {
			if cleanup != nil {
				cleanup()
			}
		}
	}()

	for iteration := 0; ; iteration++ {
		wfCtx.IsLooping = false
		if err := checkContextDone(ctx); err != nil {
			return err
		}

		if err := e.executeGraph(wf, wfCtx, iteration); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, node.ErrNodeStopped) {
				return err
			}
			if errors.Is(err, node.ErrWorkflowComplete) {
				e.updateStatus(wf.ID, iteration, "done")
				return nil
			}
			e.updateStatus(wf.ID, iteration, "error")
			return err
		}

		if !wfCtx.IsLooping {
			e.updateStatus(wf.ID, iteration, "done")
			break
		}
	}

	return nil
}

func (e *Runner) updateStatus(name string, iteration int, status string) {
	if e.observer != nil {
		e.observer.OnStatusUpdate(name, status, iteration)
	}
}

func (e *Runner) executeGraph(wf Workflow, ctx *node.RunnerContext, iteration int) error {
	e.updateStatus(wf.ID, iteration, "running")
	ctx.Iteration = iteration
	e.logInfo(wf.ID, "[Graph] Starting iteration %d", iteration)

	nodeIndex := make(map[string]int)
	var currentID string
	for i, n := range wf.Nodes {
		id := n.ID
		if id == "" {
			id = fmt.Sprintf("node-%d", i)
		}
		nodeIndex[id] = i

		// Identify the trigger node as the entry point
		if currentID == "" {
			spec, ok := e.registry.GetSpec(n.Type)
			if ok {
				for _, tag := range spec.Tags {
					if strings.EqualFold(tag, "trigger") {
						currentID = id
						break
					}
				}
			}
		}
	}
	if currentID == "" {
		if len(wf.Nodes) > 0 {
			return fmt.Errorf("no trigger node found in workflow")
		}
		return nil
	}

	edgeMap := make(map[string]string)
	for _, edge := range wf.Edges {
		key := edge.Source + ":" + edge.SourceHandle
		edgeMap[key] = edge.Target
	}

	visited := make(map[string]int)
	maxSteps := len(wf.Nodes) * 10

	for step := 0; step < maxSteps; step++ {
		if err := checkContextDone(ctx.Context); err != nil {
			return err
		}
		if currentID == "" {
			e.logInfo(wf.ID, "[Finished] End of path.")
			break
		}

		idx, ok := nodeIndex[currentID]
		if !ok {
			return fmt.Errorf("Node ID '%s' not found in node registry", currentID)
		}

		nodeDef := &wf.Nodes[idx]

		visited[currentID]++
		if visited[currentID] > 100 {
			e.logInfo(wf.ID, "[Loop Protection] Node '%s' visited >100 times. Stopping to prevent infinite loop.", currentID)
			break
		}

		handle, response, err := e.executeSingleNode(wf.ID, nodeDef, ctx, iteration)
		if err != nil {
			if errors.Is(err, node.ErrNodeStopped) {
				e.logInfo(wf.ID, "[Node] %s: Stopped by user", nodeDef.Label)
				return err
			}
			if errors.Is(err, node.ErrWorkflowComplete) {
				e.logInfo(wf.ID, "[Node] %s: %v", nodeDef.Label, err)
				return err
			}
			if errors.Is(err, context.Canceled) {
				return err
			}
			errorKey := currentID + ":error"
			nextID, hasErrorEdge := edgeMap[errorKey]
			if !hasErrorEdge {
				errorKey = currentID + ":failure"
				nextID, hasErrorEdge = edgeMap[errorKey]
			}

			if hasErrorEdge {
				nextLabel := e.getNodeLabel(wf, nextID)
				e.logInfo(wf.ID, "[Recovery] Node '%s' error. Following recovery edge to '%s'.", nodeDef.Label, nextLabel)
				e.updateStatus(wf.ID, iteration, "running")
				ctx.OutputHandle = "error"
				e.saveNodeOutput(wf.ID, nodeDef, ctx.NodeContext, err.Error())
				currentID = nextID
				continue
			}

			onErrorAction := getOnErrorAction(nodeDef)
			if onErrorAction == "stop" {
				e.updateStatus(wf.ID, iteration, "error")
				return err
			}
			handle = "success"
		}

		ctx.OutputHandle = handle
		e.saveNodeOutput(wf.ID, nodeDef, ctx.NodeContext, response)

		key := currentID + ":" + handle
		nextID, hasEdge := edgeMap[key]
		if !hasEdge {
			key = currentID + ":next"
			nextID, hasEdge = edgeMap[key]
		}

		if !hasEdge {
			e.logInfo(wf.ID, "[Finished] No edges from '%s' (handle: %s). Path complete.", e.getNodeLabel(wf, currentID), handle)
			break
		}

		nextLabel := e.getNodeLabel(wf, nextID)
		e.logInfo(wf.ID, "[Transition] %s: '%s' -> '%s'", handle, nodeDef.Label, nextLabel)
		currentID = nextID
	}

	return nil
}

func (e *Runner) executeSingleNode(wfID string, nodeDef *node.Node, ctx *node.RunnerContext, iteration int) (string, interface{}, error) {
	nodeType := nodeDef.Type
	if nodeType == "" && len(nodeDef.Nodes) > 0 {
		nodeType = "wrapper"
	}

	spec, hasSpec := e.registry.GetSpec(nodeType)

	var allDeps []node.NodeDependency
	if hasSpec && len(spec.DependsOn) > 0 {
		allDeps = append(allDeps, spec.DependsOn...)
	}
	if len(nodeDef.DependsOn) > 0 {
		allDeps = append(allDeps, nodeDef.DependsOn...)
	}

	if len(allDeps) > 0 {
		if err := node.GlobalTracker.CheckDependsOn(wfID, allDeps); err != nil {
			e.logError(wfID, "Dependency check failed for node %s (%s): %v", nodeDef.ID, nodeType, err)
			return "error", nil, err
		}
	}

	e.updateStatus(wfID, iteration, "running")

	label := nodeDef.Label
	if label == "" {
		label = nodeType
	}
	action, _ := nodeDef.Config["action"].(string)
	display := label
	if action != "" {
		display = fmt.Sprintf("%s (%s)", label, action)
	}

	e.logInfo(wfID, "[Node] Executing: %s", display)

	if len(nodeDef.Nodes) > 0 {
		e.logInfo(wfID, "[Nested] Entering sub-workflow in node '%s' (Label: %s)", nodeDef.ID, nodeDef.Label)
		return e.executeNestedNodes(wfID, nodeDef, ctx, iteration)
	}

	executor := e.registry.GetExecutor(nodeDef.Type, nodeDef.Config, nodeDef.Nodes)
	if executor == nil {
		return "", nil, fmt.Errorf("unknown node type: %s", nodeDef.Type)
	}

	retryCount := 0
	if rc, ok := nodeDef.Config["retry_count"].(float64); ok {
		retryCount = int(rc)
	}
	retryDelay := 2.0
	if rd, ok := nodeDef.Config["retry_delay"].(float64); ok && rd > 0 {
		retryDelay = rd
	}

	var err error
	for attempt := 0; attempt <= retryCount; attempt++ {
		if attempt > 0 {
			e.logInfo(wfID, "[Retry] Attempt %d/%d (delay %.1fs) for node %s", attempt, retryCount, retryDelay, nodeDef.Type)
			e.updateStatus(wfID, iteration, fmt.Sprintf("Retry %d", attempt))
			select {
			case <-ctx.Context.Done():
				return "", nil, ctx.Context.Err()
			case <-time.After(time.Duration(retryDelay * float64(time.Second))):
			}
		}
		var result node.NodeExecutionResult
		result, err = safeExecute(executor, ctx.NodeContext)
		if err == nil {
			handle := result.Handle
			if handle == "" {
				handle = "success"
			}

			actionKey, _ := nodeDef.Config["action"].(string)
			if actionKey == "" {
				actionKey = e.getActionKeyFromSpec(nodeDef.Type)
			}
			node.GlobalTracker.RecordExecuted(wfID, nodeDef.Type, actionKey)
			e.updateStatus(wfID, iteration, "success")

			hasResponse := e.checkActionHasResponse(nodeDef.Type, actionKey)
			if hasResponse && result.Response != nil && result.Type == nil {
				return "", nil, fmt.Errorf("node '%s' action '%s': kontrak memiliki response, dan node memberikan response, tetapi response type (Mime/Charset) tidak diatur", nodeDef.Type, actionKey)
			}

			e.logInfo(wfID, "[Node] Output handle: %s", handle)
			return handle, result.Response, nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, node.ErrNodeStopped) || errors.Is(err, node.ErrWorkflowComplete) {
			return "", nil, err
		}
	}

	if err != nil {
		e.logError(wfID, "Error in Node %s: %v", display, err)
		e.logInfo(wfID, "[Node] Output handle: error")
		return "error", err.Error(), err
	}

	return "", nil, nil
}

func (e *Runner) executeNestedNodes(wfID string, parentNode *node.Node, ctx *node.RunnerContext, iteration int) (string, interface{}, error) {
	childWorkflow := Workflow{
		ID:        wfID,
		Name:      parentNode.Type,
		Nodes:     parentNode.Nodes,
		Edges:     parentNode.Edges,
		Positions: parentNode.Positions,
	}

	err := e.executeGraph(childWorkflow, ctx, iteration)
	if err != nil {
		return "error", nil, err
	}

	resultHandle := ctx.OutputHandle
	if resultHandle == "" {
		resultHandle = "success"
	}

	e.logInfo(wfID, "[Nested] Wrapper node '%s' finished with handle: %s", parentNode.ID, resultHandle)
	return resultHandle, nil, nil
}

func getOnErrorAction(nodeDef *node.Node) string {
	if nodeDef.OnError != "" {
		return nodeDef.OnError
	}
	if nodeDef.ContinueOnError {
		return "continue"
	}
	return "stop"
}

func (e *Runner) saveNodeOutput(wfID string, nodeDef *node.Node, ctx *node.NodeContext, response interface{}) {
	if nodeDef.ResponseVar == "" || response == nil {
		return
	}

	val := response

	// Normalize variable name: strip {{ }} if user typed {{varName}}, consistent with UI convention
	varName := strings.TrimSpace(nodeDef.ResponseVar)
	varName = strings.TrimPrefix(varName, "{{")
	varName = strings.TrimSuffix(varName, "}}")
	varName = strings.TrimSpace(varName)

	ctx.Variables[varName] = val

	// Prepare display name for the node
	label := nodeDef.Label
	if label == "" {
		label = nodeDef.Type
	}
	action, _ := nodeDef.Config["action"].(string)
	display := label
	if action != "" {
		display = fmt.Sprintf("%s (%s)", label, action)
	}

	fullVal, _ := json.MarshalIndent(val, "", "  ")
	e.logRaw(wfID, "debug", fmt.Sprintf("[Node] Response %s variable saved as '%s'", display, varName), string(fullVal))
}

func checkContextDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	default:
		return nil
	}
}

func (e *Runner) getActionKeyFromSpec(nodeType string) string {
	spec, ok := e.registry.GetSpec(nodeType)
	if !ok {
		return ""
	}
	if spec.DefaultAction != "" {
		return spec.DefaultAction
	}
	if len(spec.Actions) > 0 {
		return spec.Actions[0].Key
	}
	return ""
}

func (e *Runner) checkActionHasResponse(nodeType, actionKey string) bool {
	spec, ok := e.registry.GetSpec(nodeType)
	if !ok {
		return false
	}
	for _, action := range spec.Actions {
		if action.Key == actionKey {
			return action.HasResponse
		}
	}
	return false
}

func (e *Runner) getNodeLabel(wf Workflow, id string) string {
	for _, n := range wf.Nodes {
		if n.ID == id {
			label := n.Label
			if label == "" {
				label = n.Type
			}
			action, _ := n.Config["action"].(string)
			if action != "" {
				return fmt.Sprintf("%s (%s)", label, action)
			}
			return label
		}
	}
	return id
}

// safeExecute wraps an executor call with panic recovery.
// If the executor panics (e.g. nil pointer, slice out of bounds),
// the panic is caught and returned as a regular error instead of
// crashing the entire server.
func safeExecute(executor node.NodeExecutor, ctx *node.NodeContext) (result node.NodeExecutionResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("node panic recovered: %v", r)
		}
	}()
	return executor.Execute(ctx)
}
