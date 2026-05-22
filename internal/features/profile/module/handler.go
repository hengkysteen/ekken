package module

import (
	"net/http"

	"ekken/internal/api"
	"ekken/internal/features/profile"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	service profile.Servicer
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	item, err := h.service.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: item})
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var req profile.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	item, err := h.service.Update(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: item})
}

func (h *ProfileHandler) VerifyPIN(c *gin.Context) {
	var req profile.VerifyPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	valid, err := h.service.VerifyPIN(req.PIN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	if !valid {
		c.JSON(http.StatusUnauthorized, api.Response{OK: false, Error: "invalid pin"})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: profile.VerifyPINResponse{Valid: valid}})
}

func (h *ProfileHandler) ResetPIN(c *gin.Context) {
	var req profile.ResetPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}

	success, err := h.service.ResetPIN(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	if !success {
		c.JSON(http.StatusUnauthorized, api.Response{OK: false, Error: "invalid security answer"})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: true})
}
