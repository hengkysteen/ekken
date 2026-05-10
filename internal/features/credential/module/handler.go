package module

import (
	"net/http"

	"ekken/internal/api"
	"ekken/internal/features/credential"

	"github.com/gin-gonic/gin"
)

type CredentialHandler struct {
	service credential.Servicer
}

// ListCredentials returns all credentials (value field excluded).
func (h *CredentialHandler) ListCredentials(c *gin.Context) {
	items, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: items})
}

// GetCredential returns a single credential by ID including the decrypted value.
func (h *CredentialHandler) GetCredential(c *gin.Context) {
	item, err := h.service.Get(c.Param("id"))
	if err != nil {
		code := http.StatusInternalServerError
		if isNotFound(err) {
			code = http.StatusNotFound
		}
		c.JSON(code, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: item})
}

// CreateCredential stores a new credential with the value encrypted at rest.
func (h *CredentialHandler) CreateCredential(c *gin.Context) {
	var req credential.Credential
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	// Return without the plaintext value for security
	item.Value = ""
	c.JSON(http.StatusCreated, api.Response{OK: true, Data: item})
}

// UpdateCredential replaces the fields of an existing credential.
func (h *CredentialHandler) UpdateCredential(c *gin.Context) {
	var req credential.Credential
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Response{OK: false, Error: err.Error()})
		return
	}
	item, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		code := http.StatusBadRequest
		if isNotFound(err) {
			code = http.StatusNotFound
		}
		c.JSON(code, api.Response{OK: false, Error: err.Error()})
		return
	}
	item.Value = ""
	c.JSON(http.StatusOK, api.Response{OK: true, Data: item})
}

// DeleteCredential removes a credential by ID.
func (h *CredentialHandler) DeleteCredential(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		code := http.StatusInternalServerError
		if isNotFound(err) {
			code = http.StatusNotFound
		}
		c.JSON(code, api.Response{OK: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.Response{OK: true, Data: gin.H{"deleted": c.Param("id")}})
}

// isNotFound is a helper to check if an error message contains "not found".
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for i := 0; i < len(msg)-8; i++ {
		if msg[i:i+9] == "not found" {
			return true
		}
	}
	return false
}
