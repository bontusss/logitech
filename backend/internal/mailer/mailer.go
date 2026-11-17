package mailer

import (
	"context"
	"fmt"
	"log/slog"
)

// Mailer is the boundary for transactional email. A provider-backed
// implementation can replace LogMailer without touching the auth service.
type Mailer interface {
	SendVerification(ctx context.Context, to, token string) error
	SendPasswordReset(ctx context.Context, to, token string) error
	SendInvitation(ctx context.Context, to, companyName, token string) error
}

// LogMailer writes the action and token to the server log. It is intentionally
// development-only: no email leaves the process.
type LogMailer struct {
	Logger *slog.Logger
	AppURL string
}

func (m *LogMailer) SendVerification(_ context.Context, to, token string) error {
	m.Logger.Info("email verification requested",
		"to", to,
		"link", fmt.Sprintf("%s/verify-email?token=%s", m.AppURL, token),
	)
	return nil
}

func (m *LogMailer) SendPasswordReset(_ context.Context, to, token string) error {
	m.Logger.Info("password reset requested",
		"to", to,
		"link", fmt.Sprintf("%s/forgot-password?token=%s", m.AppURL, token),
	)
	return nil
}

func (m *LogMailer) SendInvitation(_ context.Context, to, companyName, token string) error {
	m.Logger.Info("company invitation created",
		"to", to,
		"company", companyName,
		"token", token,
	)
	return nil
}
