package embed

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServeEmbeddedSkipsRoutesOutsideProduction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	ServeEmbedded(engine, "development")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
