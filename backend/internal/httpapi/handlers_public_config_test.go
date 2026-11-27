package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bol-platform/backend/internal/config"
)

func TestPublicConfigReturnsMapProviderConfig(t *testing.T) {
	api := &API{cfg: &config.Config{
		MapProvider:      "geoapify",
		GeoapifyAPIKey:   "test-geoapify-key",
		GeoapifyMapStyle: "osm-bright",
	}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/config", nil)

	api.publicConfig(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}
	var response struct {
		Map struct {
			Provider string `json:"provider"`
			APIKey   string `json:"api_key"`
			Style    string `json:"style"`
		} `json:"map"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode public config: %v", err)
	}
	if response.Map.Provider != "geoapify" {
		t.Fatalf("got provider %q, want %q", response.Map.Provider, "geoapify")
	}
	if response.Map.APIKey != "test-geoapify-key" {
		t.Fatalf("got key %q, want %q", response.Map.APIKey, "test-geoapify-key")
	}
	if response.Map.Style != "osm-bright" {
		t.Fatalf("got style %q, want %q", response.Map.Style, "osm-bright")
	}
}
