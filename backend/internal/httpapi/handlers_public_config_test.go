package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bol-platform/backend/internal/config"
)

func TestPublicConfigReturnsMapboxToken(t *testing.T) {
	api := &API{cfg: &config.Config{MapboxAccessToken: "pk.test-token"}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/config", nil)

	api.publicConfig(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}
	var response struct {
		MapboxAccessToken string `json:"mapbox_access_token"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode public config: %v", err)
	}
	if response.MapboxAccessToken != "pk.test-token" {
		t.Fatalf("got token %q, want %q", response.MapboxAccessToken, "pk.test-token")
	}
}
