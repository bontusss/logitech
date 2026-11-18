package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type MetaClient struct {
	AccessToken   string
	PhoneNumberID string
	APIVersion    string
	Template      string
	Language      string
	GraphAPIURL   string
	Client        *http.Client
}

func (m *MetaClient) SendVerificationCode(ctx context.Context, phone, code string) error {
	if m.AccessToken == "" || m.PhoneNumberID == "" {
		return fmt.Errorf("whatsapp Cloud API is not configured")
	}

	payload := struct {
		MessagingProduct string `json:"messaging_product"`
		RecipientType    string `json:"recipient_type"`
		To               string `json:"to"`
		Type             string `json:"type"`
		Template         struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
			Components []struct {
				Type       string `json:"type"`
				Parameters []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"parameters"`
			} `json:"components"`
		} `json:"template"`
	}{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               strings.TrimPrefix(phone, "+"),
		Type:             "template",
	}
	payload.Template.Name = m.Template
	payload.Template.Language.Code = m.Language
	payload.Template.Components = []struct {
		Type       string `json:"type"`
		Parameters []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"parameters"`
	}{{Type: "body", Parameters: []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}{{Type: "text", Text: code}}}}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("whatsapp: encode message: %w", err)
	}
	version := strings.Trim(m.APIVersion, "/")
	if version == "" {
		version = "v23.0"
	}
	baseURL := strings.TrimRight(m.GraphAPIURL, "/")
	if baseURL == "" {
		baseURL = "https://graph.facebook.com"
	}
	url := fmt.Sprintf("%s/%s/%s/messages", baseURL, version, m.PhoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("whatsapp: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	client := m.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp: send message: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("whatsapp: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	return nil
}
