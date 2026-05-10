package module

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ekken/internal/api"
	"ekken/internal/features/workflow"
	"ekken/internal/features/workflow/node"

	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
	workflows workflow.WorkflowServicer
	runtime   workflow.RuntimeServicer
	sse       workflow.SSEServicer
}

func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	items, err := h.workflows.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: items})
}

func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")
	wf, raw, err := h.workflows.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, api.Response{OK: false, Error: err.Error()})
		return
	}

	if c.Query("raw") == "true" {
		c.Data(http.StatusOK, "application/json", raw)
		return
	}

	c.JSON(http.StatusOK, api.Response{OK: true, Data: wf})
}

func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var wf workflow.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	item, _, err := h.workflows.Create(wf)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, api.Response{OK: true, Data: item})
}

func (h *WorkflowHandler) ValidateWorkflow(c *gin.Context) {
	var wf workflow.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	result := h.workflows.Validate(wf)
	if !result.Valid {
		c.JSON(http.StatusOK, api.Response{OK: false, Error: strings.Join(result.Errors, "\n")})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: result})
}

func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id := c.Param("id")
	var wf workflow.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	item, _, err := h.workflows.Update(id, wf)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: item})
}

func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")
	if err := h.workflows.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *WorkflowHandler) DeleteAllWorkflows(c *gin.Context) {
	if err := h.workflows.DeleteAll(); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *WorkflowHandler) RunWorkflow(c *gin.Context) {
	id := c.Param("id")
	if err := h.runtime.RunByID(id); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *WorkflowHandler) RunWorkflowPayload(c *gin.Context) {
	var wf workflow.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	// Validate before running
	result := h.workflows.ValidateForRun(wf)
	if !result.Valid {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: fmt.Sprintf("Workflow validation failed: %v", result.Errors)})
		return
	}

	if err := h.runtime.RunWorkflow(wf); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *WorkflowHandler) StopWorkflow(c *gin.Context) {
	id := c.Param("id")
	if err := h.runtime.Stop(id); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *WorkflowHandler) WorkflowStatus(c *gin.Context) {
	id := c.Param("id")
	status := h.runtime.Status(id)
	c.JSON(http.StatusOK, api.Response{OK: true, Data: status})
}

func (h *WorkflowHandler) WorkflowsStatus(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	subID, ch := h.sse.SubscribeGlobal()
	defer func() {
		h.sse.UnsubscribeGlobal(subID)
	}()

	// Initial sync: Send current status of all running workflows
	for _, r := range h.runtime.Running() {
		data, _ := json.Marshal(r)
		c.SSEvent("status_update", string(data))
	}

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-ch:
			if !ok {
				return false
			}
			data, _ := json.Marshal(msg.Data)
			c.SSEvent(msg.Type, string(data))
			w.Write([]byte("\n"))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

func (h *WorkflowHandler) WorkflowLogs(c *gin.Context) {
	id := c.Param("id")
	logs := h.runtime.Logs(id)
	c.JSON(http.StatusOK, api.Response{OK: true, Data: logs})
}

func (h *WorkflowHandler) DeleteWorkflowLogs(c *gin.Context) {
	id := c.Param("id")
	if err := h.runtime.DeleteLogs(id); err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true})
}

func (h *WorkflowHandler) SSEStream(c *gin.Context) {
	id := c.Param("id")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	subID, ch := h.sse.Subscribe(id)
	defer func() {
		h.sse.Unsubscribe(id, subID)
	}()

	// Initial sync: Send current status for this specific workflow
	status := h.runtime.Status(id)
	data, _ := json.Marshal(status)
	c.SSEvent("status_update", string(data))

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-ch:
			if !ok {
				return false
			}
			data, _ := json.Marshal(msg.Data)
			c.SSEvent(msg.Type, string(data))
			w.Write([]byte("\n"))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

func (h *WorkflowHandler) NodeCatalog(c *gin.Context) {
	parent := c.Query("parent")
	regs := node.GlobalRegistry.AllSpecs()

	filtered := make([]node.NodeSpec, 0)
	for _, r := range regs {
		if parent != "" && r.Parent != parent {
			continue
		}
		if parent == "" && r.Parent != "" {
			// Skip children if they're not explicitly asked for
			continue
		}
		filtered = append(filtered, r)
	}

	c.JSON(http.StatusOK, api.Response{OK: true, Data: filtered})
}
