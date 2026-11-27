package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getCompanyOverview returns whole-company delivery and collection totals. The
// sums and counts are computed in SQL (bucketed by the company timezone) so
// they are not limited to the most recent deliveries like the list endpoint.
func (a *API) getCompanyOverview(c *gin.Context) {
	tenant, _ := currentTenant(c)
	ctx := c.Request.Context()

	overview, err := a.svc.Q.GetCompanyDeliveryOverview(ctx, tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	dailyRows, err := a.svc.Q.ListCompanyDeliveryDailyCounts(ctx, tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	daily := make([]gin.H, 0, len(dailyRows))
	for _, row := range dailyRows {
		day := ""
		if row.Day.Valid {
			day = row.Day.Time.Format("2006-01-02")
		}
		daily = append(daily, gin.H{"date": day, "bookings": row.Bookings})
	}

	c.JSON(http.StatusOK, gin.H{"overview": gin.H{
		"financial": gin.H{
			"earned_today_minor":      overview.EarnedTodayMinor,
			"collected_today_minor":   overview.CollectedTodayMinor,
			"amount_owed_minor":       overview.AmountOwedMinor,
			"cash_with_drivers_minor": overview.CashWithDriversMinor,
		},
		"bookings": gin.H{
			"created_today":       overview.BookingsToday,
			"in_progress":         overview.InProgress,
			"waiting_for_drivers": overview.WaitingForDrivers,
			"requests":            overview.BookingRequests,
			"total":               overview.TotalBookings,
		},
		"daily": daily,
	}})
}
