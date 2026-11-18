package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	db "github.com/bol-platform/backend/internal/db/sqlc"
)

func (a *API) getMe(c *gin.Context) {
	user, _ := currentUser(c)
	memberships, err := a.svc.Q.ListUserMemberships(c.Request.Context(), user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	userResponse := toUserDTO(user)
	userResponse.WhatsAppVerified, err = a.svc.IsWhatsAppVerified(c.Request.Context(), user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":      userResponse,
		"companies": toMembershipDTOs(memberships),
	})
}

func (a *API) updateMe(c *gin.Context) {
	user, _ := currentUser(c)
	var req struct {
		FullName *string `json:"full_name" binding:"omitempty,min=2,max=120"`
		Phone    *string `json:"phone" binding:"omitempty,max=30"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the profile details you entered.")
		return
	}

	fullName := valueOr(req.FullName, user.FullName)
	phone := valueOr(req.Phone, deref(user.Phone))
	if strings.TrimSpace(fullName) == "" {
		fail(c, http.StatusBadRequest, "invalid_input", "Your name cannot be empty.")
		return
	}

	updated, err := a.svc.Q.UpdateUserProfile(c.Request.Context(), db.UpdateUserProfileParams{
		ID:       user.ID,
		FullName: strings.TrimSpace(fullName),
		Column3:  strings.TrimSpace(phone),
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": toUserDTO(updated)})
}
