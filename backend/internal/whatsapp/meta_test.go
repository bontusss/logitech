package whatsapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetaClientSendsTemplateCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v23.0/123456/messages", r.URL.Path)
		require.Equal(t, "Bearer access-token", r.Header.Get("Authorization"))
		var payload struct {
			To       string `json:"to"`
			Template struct {
				Name     string `json:"name"`
				Language struct {
					Code string `json:"code"`
				} `json:"language"`
				Components []struct {
					Type       string `json:"type"`
					Parameters []struct {
						Text string `json:"text"`
					} `json:"parameters"`
				} `json:"components"`
			} `json:"template"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "2348012345678", payload.To)
		require.Equal(t, "bol_phone_verification", payload.Template.Name)
		require.Equal(t, "en_US", payload.Template.Language.Code)
		require.Equal(t, "123456", payload.Template.Components[0].Parameters[0].Text)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &MetaClient{
		AccessToken:   "access-token",
		PhoneNumberID: "123456",
		APIVersion:    "v23.0",
		Template:      "bol_phone_verification",
		Language:      "en_US",
		GraphAPIURL:   server.URL,
		Client:        server.Client(),
	}
	require.NoError(t, client.SendVerificationCode(context.Background(), "+2348012345678", "123456"))
}
