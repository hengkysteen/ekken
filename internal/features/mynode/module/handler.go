package module

import (
	"errors"
	"net/http"

	"ekken/internal/api"
	"ekken/internal/features/mynode"

	"github.com/gin-gonic/gin"
)

var ErrMyNodesItemNotFound = errors.New("my nodes item not found")

type MyNodeHandler struct {
	service mynode.MyNodesServicer
}

func (h *MyNodeHandler) ListMyNodes(c *gin.Context) {
	items, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: items})
}

func (h *MyNodeHandler) SaveMyNodes(c *gin.Context) {
	var req mynode.MyNodesItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	item, err := h.service.Save(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, api.Response{OK: true, Data: item})
}

func (h *MyNodeHandler) DeleteMyNodes(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, ErrMyNodesItemNotFound) {
			code = http.StatusNotFound
		}
		c.JSON(code, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: gin.H{"deleted": c.Param("id")}})
}

func (h *MyNodeHandler) UpdateMyNodes(c *gin.Context) {
	var req mynode.MyNodesItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	item, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, ErrMyNodesItemNotFound) {
			code = http.StatusNotFound
		}
		c.JSON(code, api.Response{OK: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.Response{OK: true, Data: item})
}
