package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Mailer interface {
	SendVerification(ctx context.Context, to, token string) error
	SendPasswordReset(ctx context.Context, to, token string) error
	SendInvitation(ctx context.Context, to, companyName, token string) error
	SendWelcome(ctx context.Context, to, fullName string, provider bool) error
	SendOnboardingReminder(ctx context.Context, to, fullName, companyName string, reminder int) error
	SendProviderApproved(ctx context.Context, to, fullName, companyName string) error
	SendDriverWelcome(ctx context.Context, to, fullName, companyName, portalURL string) error
}

type LogMailer struct {
	Logger *slog.Logger
	AppURL string
}

func (m *LogMailer) SendVerification(_ context.Context, to, token string) error {
	m.Logger.Info("email verification requested",
		"to", to,
		"link", fmt.Sprintf("%s/verify-email?token=%s", strings.TrimRight(m.AppURL, "/"), token),
	)
	return nil
}

func (m *LogMailer) SendPasswordReset(_ context.Context, to, token string) error {
	m.Logger.Info("password reset requested",
		"to", to,
		"link", fmt.Sprintf("%s/forgot-password?token=%s", strings.TrimRight(m.AppURL, "/"), token),
	)
	return nil
}

func (m *LogMailer) SendInvitation(_ context.Context, to, companyName, token string) error {
	m.Logger.Info("company invitation created", "to", to, "company", companyName, "token", token)
	return nil
}

func (m *LogMailer) SendWelcome(_ context.Context, to, fullName string, provider bool) error {
	m.Logger.Info("welcome email scheduled", "to", to, "name", fullName, "provider", provider)
	return nil
}

func (m *LogMailer) SendOnboardingReminder(_ context.Context, to, fullName, companyName string, reminder int) error {
	m.Logger.Info("provider onboarding reminder scheduled",
		"to", to, "name", fullName, "company", companyName, "reminder", reminder)
	return nil
}

func (m *LogMailer) SendProviderApproved(_ context.Context, to, fullName, companyName string) error {
	m.Logger.Info("provider approval email", "to", to, "name", fullName, "company", companyName)
	return nil
}

func (m *LogMailer) SendDriverWelcome(_ context.Context, to, fullName, companyName, _ string) error {
	m.Logger.Info("driver welcome email requested", "to", to, "name", fullName, "company", companyName)
	return nil
}

type PlunkMailer struct {
	SecretKey string
	FromEmail string
	AppURL    string
	APIURL    string
	Client    *http.Client
}

func NewPlunk(secretKey, fromEmail, appURL string) *PlunkMailer {
	return &PlunkMailer{
		SecretKey: secretKey,
		FromEmail: fromEmail,
		AppURL:    strings.TrimRight(appURL, "/"),
		APIURL:    "https://next-api.useplunk.com/v1/send",
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *PlunkMailer) SendVerification(ctx context.Context, to, token string) error {
	link := m.AppURL + "/verify-email?token=" + token
	return m.send(ctx, to, "Confirm your BOL email address",
		fmt.Sprintf(`<p>Hi,</p><p>Confirm your email address to finish setting up your BOL account.</p><p><a href="%s">Verify email address</a></p><p>If you did not create this account, you can ignore this email.</p>`, html.EscapeString(link)))
}

func (m *PlunkMailer) SendPasswordReset(ctx context.Context, to, token string) error {
	link := m.AppURL + "/forgot-password?token=" + token
	return m.send(ctx, to, "Reset your BOL password",
		fmt.Sprintf(`<p>We received a request to reset your BOL password.</p><p><a href="%s">Choose a new password</a></p><p>If you did not request this, you can ignore this email.</p>`, html.EscapeString(link)))
}

func (m *PlunkMailer) SendInvitation(ctx context.Context, to, companyName, token string) error {
	link := m.AppURL + "/invitations/" + token
	return m.send(ctx, to, "You are invited to join "+companyName+" on BOL",
		fmt.Sprintf(`<p>You have been invited to join <strong>%s</strong> on BOL.</p><p><a href="%s">Accept invitation</a></p>`, html.EscapeString(companyName), html.EscapeString(link)))
}

func (m *PlunkMailer) SendWelcome(ctx context.Context, to, fullName string, provider bool) error {
	name := html.EscapeString(fullName)
	if provider {
		return m.send(ctx, to, "Welcome to BOL — grow your logistics business",
			fmt.Sprintf(`<p>Hi %s,</p><p>Welcome to BOL. Your provider account is ready.</p><p>Complete your company profile, verify your WhatsApp number, and submit your business for review so shippers can discover your services.</p><p><a href="%s/onboarding">Continue provider onboarding</a></p><p>— Nnanna from BOL</p>`, name, m.AppURL))
	}
	return m.send(ctx, to, "Welcome to BOL — shipping made simpler",
		fmt.Sprintf(`<p>Hi %s,</p><p>Welcome to BOL. Compare trusted logistics providers and find a shipping option that works for you.</p><p><a href="%s/dashboard">Explore BOL</a></p><p>— Nnanna from BOL</p>`, name, m.AppURL))
}

func (m *PlunkMailer) SendOnboardingReminder(ctx context.Context, to, fullName, companyName string, reminder int) error {
	name := html.EscapeString(fullName)
	company := html.EscapeString(companyName)
	subject := "Your BOL provider profile is waiting"
	body := fmt.Sprintf(`<p>Hi %s,</p><p>Your provider profile for <strong>%s</strong> is not complete yet. Finish onboarding to help shippers find your services.</p><p><a href="%s/onboarding">Continue onboarding</a></p><p>Need help? Reply to this email and we will be glad to assist.</p><p>— Nnanna from BOL</p>`, name, company, m.AppURL)
	if reminder > 1 {
		subject = "Need a hand finishing your BOL provider profile?"
		body = fmt.Sprintf(`<p>Hi %s,</p><p>Your provider profile for <strong>%s</strong> is still waiting for a few details. You can pick up where you left off whenever you are ready.</p><p><a href="%s/onboarding">Finish provider onboarding</a></p><p>— Nnanna from BOL</p>`, name, company, m.AppURL)
	}
	return m.send(ctx, to, subject, body)
}

func (m *PlunkMailer) SendProviderApproved(ctx context.Context, to, fullName, companyName string) error {
	name := html.EscapeString(fullName)
	company := html.EscapeString(companyName)
	return m.send(ctx, to, company+" is approved on BOL",
		fmt.Sprintf(`<p>Hi %s,</p><p>Good news — <strong>%s</strong> has been verified and is now live on the BOL marketplace.</p><p>Shippers can find your company, and you can publish rates and receive bookings from your workspace.</p><p><a href="%s/dashboard/company">Open your provider dashboard</a></p><p>— Nnanna from BOL</p>`, name, company, m.AppURL))
}

func (m *PlunkMailer) SendDriverWelcome(ctx context.Context, to, fullName, companyName, portalURL string) error {
	name := html.EscapeString(fullName)
	company := html.EscapeString(companyName)
	link := html.EscapeString(portalURL)
	return m.send(ctx, to, "Welcome to "+company+" — your driver portal is ready",
		fmt.Sprintf(`<p>Hi %s,</p><p><strong>%s</strong> has added you as a driver on BOL.</p><p>Use your personal driver portal to view delivery assignments. Your browser location will only be requested while you have an active delivery assignment.</p><p><a href="%s">Open your driver portal</a></p><p>Keep this link private. If you have questions, contact your company.</p><p>— Nnanna from BOL</p>`, name, company, link))
}

func (m *PlunkMailer) send(ctx context.Context, to, subject, body string) error {
	payload := struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		From    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"from"`
	}{
		To:      to,
		Subject: subject,
		Body:    body,
	}
	payload.From.Name = "Nnanna from BOL"
	payload.From.Email = m.FromEmail

	bodyReader, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	apiURL := m.APIURL
	if apiURL == "" {
		apiURL = "https://next-api.useplunk.com/v1/send"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyReader))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client().Do(req)
	if err != nil {
		return fmt.Errorf("plunk send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("plunk send: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	return nil
}

func (m *PlunkMailer) client() *http.Client {
	if m.Client != nil {
		return m.Client
	}
	return http.DefaultClient
}
