package module

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	m := &PluginModule{}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RegisterRoutes panicked: %v", r)
		}
	}()

	m.RegisterRoutes(api)
}
