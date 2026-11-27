package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getCustomerOverview returns whole-account booking counts and payment totals,
// computed in SQL so the summary is not limited to the latest bookings list.
func (a *API) getCustomerOverview(c *gin.Context) {
	user, _ := currentUser(c)
	overview, err := a.svc.Q.GetCustomerBookingOverview(c.Request.Context(), &user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"overview": overview})
}
