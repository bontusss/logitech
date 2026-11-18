package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (a *API) resendEmailVerification(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "unauthenticated", "Sign in to continue.")
		return
	}
	verified, err := a.svc.ResendEmailVerification(c.Request.Context(), user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if verified {
		c.JSON(http.StatusOK, gin.H{"message": "Your email is already verified."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "A verification email has been sent."})
}

func (a *API) sendWhatsAppCode(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "unauthenticated", "Sign in to continue.")
		return
	}
	var req struct {
		Phone string `json:"phone" binding:"required,max=30"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Enter a WhatsApp number in international format.")
		return
	}
	verified, err := a.svc.SendWhatsAppCode(c.Request.Context(), user.ID, strings.TrimSpace(req.Phone))
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if verified {
		c.JSON(http.StatusOK, gin.H{"verified": true, "message": "This WhatsApp number is already verified."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"verified": false, "message": "A verification code has been sent on WhatsApp."})
}

func (a *API) verifyWhatsAppCode(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "unauthenticated", "Sign in to continue.")
		return
	}
	var req struct {
		Code string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Enter the six-digit code sent on WhatsApp.")
		return
	}
	if err := a.svc.VerifyWhatsAppCode(c.Request.Context(), user.ID, req.Code); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"verified": true, "message": "WhatsApp number verified."})
}
