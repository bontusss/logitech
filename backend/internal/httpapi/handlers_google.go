package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bol-platform/backend/internal/service"
)

const (
	googleAuthEndpoint      = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenEndpoint     = "https://oauth2.googleapis.com/token"
	googleTokenInfoEndpoint = "https://oauth2.googleapis.com/tokeninfo"
	googleStateCookieName   = "bol_google_oauth_state"
	googleStateCookiePath   = "/api/v1/auth/google"
)

var googleHTTPClient = &http.Client{Timeout: 10 * time.Second}

func (a *API) googleConfigured() bool {
	return a.cfg.GoogleClientID != "" &&
		a.cfg.GoogleClientSecret != "" &&
		a.cfg.GoogleRedirectURI != ""
}

func (a *API) googleAuthStart(c *gin.Context) {
	intent := strings.ToLower(strings.TrimSpace(c.Query("intent")))
	switch intent {
	case "login", "customer", "provider":
	default:
		intent = "login"
	}
	next := safeAuthNext(c.Query("next"))

	if !a.googleConfigured() {
		a.redirectGoogleError(c, "not_configured")
		return
	}

	state, err := a.auth.SignOAuthState(intent, next)
	if err != nil {
		a.log.Error("could not sign google oauth state", "error", err)
		a.redirectGoogleError(c, "google_failed")
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     googleStateCookieName,
		Value:    state,
		Path:     googleStateCookiePath,
		MaxAge:   600,
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	params := url.Values{
		"client_id":     {a.cfg.GoogleClientID},
		"redirect_uri":  {a.cfg.GoogleRedirectURI},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"online"},
		"prompt":        {"select_account"},
	}

	c.Redirect(http.StatusFound, googleAuthEndpoint+"?"+params.Encode())
}

func (a *API) googleAuthCallback(c *gin.Context) {
	stateCookie, cookieErr := c.Cookie(googleStateCookieName)
	a.clearGoogleStateCookie(c)

	state := strings.TrimSpace(c.Query("state"))
	if cookieErr != nil || state == "" ||
		subtle.ConstantTimeCompare([]byte(stateCookie), []byte(state)) != 1 {
		a.redirectGoogleError(c, "state")
		return
	}

	claims, err := a.auth.ParseOAuthState(stateCookie)
	if err != nil {
		a.redirectGoogleError(c, "state")
		return
	}

	if c.Query("error") != "" {
		a.redirectGoogleError(c, "cancelled")
		return
	}
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		a.redirectGoogleError(c, "google_failed")
		return
	}
	if !a.googleConfigured() {
		a.redirectGoogleError(c, "not_configured")
		return
	}

	profile, err := a.fetchGoogleProfile(c.Request.Context(), code)
	if err != nil {
		a.log.Warn("google oauth profile lookup failed", "error", err)
		a.redirectGoogleError(c, "google_failed")
		return
	}

	result, err := a.svc.GoogleSignIn(
		c.Request.Context(), profile, claims.Intent,
		c.ClientIP(), c.Request.UserAgent(),
	)
	if err != nil {
		a.log.Warn("google oauth sign-in rejected", "error", err, "intent", claims.Intent)
		a.redirectGoogleError(c, googleErrorCode(err))
		return
	}

	a.setSessionCookies(c, result.Session)

	target := safeAuthNext(claims.Next)
	if result.NeedsCompanySetup {
		target = "/register?google=provider"
	}
	c.Redirect(http.StatusFound, strings.TrimRight(a.cfg.AppURL, "/")+target)
}

func (a *API) clearGoogleStateCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     googleStateCookieName,
		Value:    "",
		Path:     googleStateCookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *API) redirectGoogleError(c *gin.Context, code string) {
	destination := strings.TrimRight(a.cfg.AppURL, "/") + "/login?google_error=" + url.QueryEscape(code)
	c.Redirect(http.StatusFound, destination)
}

func googleErrorCode(err error) string {
	switch {
	case errors.Is(err, service.ErrGoogleAccountRestricted):
		return "admin_blocked"
	case errors.Is(err, service.ErrGoogleAccountNotFound):
		return "no_account"
	case errors.Is(err, service.ErrGoogleAccountConflict):
		return "email_exists"
	case errors.Is(err, service.ErrEmailTaken):
		return "email_exists"
	case errors.Is(err, service.ErrUserSuspended):
		return "suspended"
	default:
		return "google_failed"
	}
}

func safeAuthNext(next string) string {
	next = strings.TrimSpace(next)
	if next == "" ||
		!strings.HasPrefix(next, "/") ||
		strings.HasPrefix(next, "//") ||
		strings.Contains(next, "\\") {
		return "/dashboard"
	}
	return next
}

type googleIDTokenInfo struct {
	Sub           string          `json:"sub"`
	Email         string          `json:"email"`
	EmailVerified json.RawMessage `json:"email_verified"`
	Name          string          `json:"name"`
	Audience      string          `json:"aud"`
	Issuer        string          `json:"iss"`
}

func (a *API) fetchGoogleProfile(ctx context.Context, code string) (service.GoogleProfile, error) {
	form := url.Values{
		"code":          {code},
		"client_id":     {a.cfg.GoogleClientID},
		"client_secret": {a.cfg.GoogleClientSecret},
		"redirect_uri":  {a.cfg.GoogleRedirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return service.GoogleProfile{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := googleHTTPClient.Do(req)
	if err != nil {
		return service.GoogleProfile{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return service.GoogleProfile{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return service.GoogleProfile{}, fmt.Errorf("google token exchange returned %d", resp.StatusCode)
	}

	var tokenResponse struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil || tokenResponse.IDToken == "" {
		return service.GoogleProfile{}, errors.New("google token response did not include an id_token")
	}

	info, err := a.verifyGoogleIDToken(ctx, tokenResponse.IDToken)
	if err != nil {
		return service.GoogleProfile{}, err
	}

	return service.GoogleProfile{
		Subject:       info.Sub,
		Email:         info.Email,
		EmailVerified: googleEmailVerified(info.EmailVerified),
		FullName:      info.Name,
	}, nil
}

func (a *API) verifyGoogleIDToken(ctx context.Context, idToken string) (googleIDTokenInfo, error) {
	params := url.Values{"id_token": {idToken}}
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, googleTokenInfoEndpoint+"?"+params.Encode(), nil,
	)
	if err != nil {
		return googleIDTokenInfo{}, err
	}

	resp, err := googleHTTPClient.Do(req)
	if err != nil {
		return googleIDTokenInfo{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return googleIDTokenInfo{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return googleIDTokenInfo{}, fmt.Errorf("google token verification returned %d", resp.StatusCode)
	}

	var info googleIDTokenInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return googleIDTokenInfo{}, err
	}
	if info.Audience != a.cfg.GoogleClientID {
		return googleIDTokenInfo{}, errors.New("google id token audience mismatch")
	}
	if info.Issuer != "accounts.google.com" && info.Issuer != "https://accounts.google.com" {
		return googleIDTokenInfo{}, errors.New("google id token issuer mismatch")
	}
	if info.Sub == "" || info.Email == "" || !googleEmailVerified(info.EmailVerified) {
		return googleIDTokenInfo{}, errors.New("google account email is not verified")
	}

	return info, nil
}

func googleEmailVerified(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var boolean bool
	if err := json.Unmarshal(raw, &boolean); err == nil {
		return boolean
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text == "true"
	}
	return false
}
