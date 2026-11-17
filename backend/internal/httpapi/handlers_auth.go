package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bol-platform/backend/internal/service"
)

type registerRequest struct {
	FullName string `json:"full_name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Phone    string `json:"phone" binding:"omitempty,max=30"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"omitempty"`
}

type tokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type resetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

func (a *API) registerCustomer(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the name, email and password you entered.")
		return
	}

	session, err := a.svc.RegisterCustomer(
		c.Request.Context(), req.FullName, req.Email, req.Password, req.Phone,
		c.ClientIP(), c.Request.UserAgent(),
	)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.setSessionCookies(c, session)
	c.JSON(http.StatusCreated, toSessionDTO(session))
}

func (a *API) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Enter your email and password.")
		return
	}

	session, err := a.svc.Login(c.Request.Context(), req.Email, req.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.setSessionCookies(c, session)
	c.JSON(http.StatusOK, toSessionDTO(session))
}

func (a *API) refresh(c *gin.Context) {
	raw := ""
	if cookie, err := c.Cookie(refreshCookieName); err == nil {
		raw = cookie
	}
	if raw == "" {
		var req refreshRequest
		_ = c.ShouldBindJSON(&req)
		raw = strings.TrimSpace(req.RefreshToken)
	}
	if raw == "" {
		fail(c, http.StatusUnauthorized, "refresh_required", "No refresh token was provided.")
		return
	}

	session, err := a.svc.Refresh(c.Request.Context(), raw, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.setSessionCookies(c, session)
	c.JSON(http.StatusOK, toSessionDTO(session))
}

func (a *API) logout(c *gin.Context) {
	raw := ""
	if cookie, err := c.Cookie(refreshCookieName); err == nil {
		raw = cookie
	}
	if raw == "" {
		var req refreshRequest
		_ = c.ShouldBindJSON(&req)
		raw = strings.TrimSpace(req.RefreshToken)
	}

	// Best-effort access-token denylisting; an expired token needs no entry.
	if token := accessTokenFromRequest(c); token != "" {
		if claims, err := a.auth.Parse(token); err == nil {
			expiry := claims.ExpiresAt.Time
			a.svc.DenyAccessToken(c.Request.Context(), claims.ID, expiry)
		}
	}

	a.svc.Logout(c.Request.Context(), raw, c.ClientIP(), c.Request.UserAgent())
	a.clearSessionCookies(c)
	c.Status(http.StatusNoContent)
}

func (a *API) forgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Enter a valid email address.")
		return
	}

	devToken, err := a.svc.ForgotPassword(c.Request.Context(), req.Email, c.ClientIP())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	response := gin.H{"message": "If that email is registered, a reset link has been sent."}
	if devToken != "" {
		response["dev_reset_token"] = devToken
	}
	c.JSON(http.StatusOK, response)
}

func (a *API) resetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "The reset link or new password is invalid.")
		return
	}

	if err := a.svc.ResetPassword(c.Request.Context(), req.Token, req.Password, c.ClientIP()); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) verifyEmail(c *gin.Context) {
	var req tokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "A verification token is required.")
		return
	}
	if err := a.svc.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Email verified."})
}

var _ = service.ErrInvalidToken
