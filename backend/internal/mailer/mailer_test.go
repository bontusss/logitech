package mailer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlunkMailerUsesConfiguredSenderForEveryMessage(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		require.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
		require.Equal(t, "/v1/send", r.URL.Path)
		var payload struct {
			From struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"from"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "Nnanna from BOL", payload.From.Name)
		require.Equal(t, "hello@example.com", payload.From.Email)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	mailer := NewPlunk("secret", "hello@example.com", "https://bol.example")
	mailer.APIURL = server.URL + "/v1/send"
	mailer.Client = server.Client()
	ctx := context.Background()

	require.NoError(t, mailer.SendVerification(ctx, "person@example.com", "verify-token"))
	require.NoError(t, mailer.SendPasswordReset(ctx, "person@example.com", "reset-token"))
	require.NoError(t, mailer.SendInvitation(ctx, "person@example.com", "Aba Freight", "invite-token"))
	require.NoError(t, mailer.SendWelcome(ctx, "person@example.com", "Ada", false))
	require.NoError(t, mailer.SendWelcome(ctx, "person@example.com", "Ada", true))
	require.NoError(t, mailer.SendOnboardingReminder(ctx, "person@example.com", "Ada", "Aba Freight", 1))
	require.Equal(t, 6, requests)
}
