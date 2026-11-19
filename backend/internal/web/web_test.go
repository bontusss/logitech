package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestHandlerServesGeneratedFilesAndFallsBackForAppRoutes(t *testing.T) {
	handler := newHandler(fstest.MapFS{
		"index.html":       &fstest.MapFile{Data: []byte("home")},
		"200.html":         &fstest.MapFile{Data: []byte("spa shell")},
		"_nuxt/app.js":     &fstest.MapFile{Data: []byte("asset")},
		"login/index.html": &fstest.MapFile{Data: []byte("login")},
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{name: "root page", path: "/", wantStatus: http.StatusOK, wantBody: "home"},
		{name: "generated page", path: "/login/", wantStatus: http.StatusOK, wantBody: "login"},
		{name: "frontend asset", path: "/_nuxt/app.js", wantStatus: http.StatusOK, wantBody: "asset"},
		{name: "client route", path: "/invitations/token", wantStatus: http.StatusOK, wantBody: "spa shell"},
		{name: "dynamic booking route", path: "/book/2seconds", wantStatus: http.StatusOK, wantBody: "spa shell"},
		{name: "missing asset", path: "/sw.js", wantStatus: http.StatusNotFound, wantBody: "404 page not found\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))
			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Equal(t, tt.wantBody, recorder.Body.String())
		})
	}
}

func TestHandlerFallsBackToHomeWhenSpaFallbackIsMissing(t *testing.T) {
	handler := newHandler(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("home")},
	})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/client-route", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "home", recorder.Body.String())
}

func TestHandlerRejectsNonReadMethods(t *testing.T) {
	handler := newHandler(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("home")},
	})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", nil))
	require.Equal(t, http.StatusNotFound, recorder.Code)
}
