package httpapi

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

func (a *API) getItemCategoryPolicy(c *gin.Context) {
	tenant, _ := currentTenant(c)
	policy, err := a.svc.Q.GetItemCategoryPolicy(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"accepted":   nonNilStrings(policy.AcceptedItemCategories),
		"prohibited": nonNilStrings(policy.ProhibitedItemCategories),
	})
}

func (a *API) updateItemCategoryPolicy(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req struct {
		Accepted   []string `json:"accepted"`
		Prohibited []string `json:"prohibited"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_item_policy", "Choose the categories you accept and prohibit.")
		return
	}
	accepted, ok := service.NormalizeItemCategoryList(req.Accepted)
	if !ok {
		fail(c, http.StatusUnprocessableEntity, "invalid_item_policy", "One of the accepted categories is not supported.")
		return
	}
	prohibited, ok := service.NormalizeItemCategoryList(req.Prohibited)
	if !ok {
		fail(c, http.StatusUnprocessableEntity, "invalid_item_policy", "One of the prohibited categories is not supported.")
		return
	}
	for _, category := range accepted {
		if slices.Contains(prohibited, category) {
			fail(c, http.StatusUnprocessableEntity, "invalid_item_policy", "A category cannot be both accepted and prohibited.")
			return
		}
	}

	policy, err := a.svc.Q.UpdateItemCategoryPolicy(c.Request.Context(), db.UpdateItemCategoryPolicyParams{
		ID:      tenant.CompanyID,
		Column2: nonNilStrings(accepted),
		Column3: nonNilStrings(prohibited),
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.item_categories_updated",
		EntityType: "company",
		EntityID:   &tenant.CompanyID,
		Metadata:   map[string]any{"accepted": accepted, "prohibited": prohibited},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{
		"accepted":   nonNilStrings(policy.AcceptedItemCategories),
		"prohibited": nonNilStrings(policy.ProhibitedItemCategories),
	})
}
