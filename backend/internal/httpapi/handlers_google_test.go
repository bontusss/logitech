package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bol-platform/backend/internal/auth"
	"github.com/bol-platform/backend/internal/config"
)

func TestSafeAuthNext(t *testing.T) {
	tests := []struct {
		name string
		next string
		want string
	}{
		{name: "empty", next: "", want: "/dashboard"},
		{name: "relative path", next: "/dashboard/deliveries?tab=active", want: "/dashboard/deliveries?tab=active"},
		{name: "protocol relative", next: "//evil.example", want: "/dashboard"},
		{name: "absolute URL", next: "https://evil.example", want: "/dashboard"},
		{name: "backslash", next: "/\\evil.example", want: "/dashboard"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := safeAuthNext(test.next); got != test.want {
				t.Fatalf("safeAuthNext(%q) = %q, want %q", test.next, got, test.want)
			}
		})
	}
}

func TestGoogleAuthStartBuildsGoogleRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &API{
		cfg: &config.Config{
			GoogleClientID:     "test-client",
			GoogleClientSecret: "test-secret",
			GoogleRedirectURI:  "http://localhost:8080/api/v1/auth/google/callback",
			AppURL:             "http://localhost:3000",
		},
		auth: auth.NewManager(strings.Repeat("s", 32), 15*time.Minute),
		log:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	router := gin.New()
	router.GET("/start", api.googleAuthStart)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/start?intent=provider&next=%2Fregister%3Fgoogle%3Dprovider", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", recorder.Code)
	}

	destination, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	if destination.Host != "accounts.google.com" {
		t.Fatalf("redirect host = %q, want accounts.google.com", destination.Host)
	}
	if destination.Query().Get("client_id") != "test-client" {
		t.Fatal("client_id missing from redirect")
	}
	if destination.Query().Get("state") == "" {
		t.Fatal("state missing from redirect")
	}

	var stateCookie *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == googleStateCookieName {
			stateCookie = cookie
		}
	}
	if stateCookie == nil {
		t.Fatal("oauth state cookie was not set")
	}
	if !stateCookie.HttpOnly || stateCookie.Path != googleStateCookiePath {
		t.Fatalf("unexpected state cookie: HttpOnly=%v path=%q", stateCookie.HttpOnly, stateCookie.Path)
	}
}

func TestGoogleAuthStartWithoutConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &API{
		cfg:  &config.Config{AppURL: "http://localhost:3000"},
		auth: auth.NewManager(strings.Repeat("s", 32), 15*time.Minute),
		log:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	router := gin.New()
	router.GET("/start", api.googleAuthStart)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/start?intent=login", nil))

	location := recorder.Header().Get("Location")
	if !strings.Contains(location, "google_error=not_configured") {
		t.Fatalf("location = %q, want google_error=not_configured", location)
	}
}

func TestGoogleEmailVerified(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{raw: `true`, want: true},
		{raw: `"true"`, want: true},
		{raw: `false`, want: false},
		{raw: `"false"`, want: false},
		{raw: ``, want: false},
	}

	for _, test := range tests {
		if got := googleEmailVerified(json.RawMessage(test.raw)); got != test.want {
			t.Fatalf("googleEmailVerified(%s) = %v, want %v", test.raw, got, test.want)
		}
	}
}
