package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

type publicBranchDTO struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	City           string    `json:"city"`
	State          string    `json:"state"`
	Phone          *string   `json:"phone"`
	Email          string    `json:"email"`
	OperatingHours any       `json:"operating_hours"`
}

func publicBranchFromRow(row db.ListPublicCompanyBranchesRow) publicBranchDTO {
	var hours any
	if len(row.OperatingHours) > 0 {
		_ = json.Unmarshal(row.OperatingHours, &hours)
	}
	if hours == nil {
		hours = gin.H{}
	}
	return publicBranchDTO{
		ID:             row.ID,
		Name:           row.Name,
		Address:        row.Address,
		City:           row.City,
		State:          row.State,
		Phone:          row.Phone,
		Email:          row.Email,
		OperatingHours: hours,
	}
}

type publicReviewDTO struct {
	ID           uuid.UUID `json:"id"`
	ReviewerName string    `json:"reviewer_name"`
	Rating       int16     `json:"rating"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"created_at"`
}

func publicReviewFromRow(row db.ListCompanyReviewsRow) publicReviewDTO {
	return publicReviewDTO{
		ID:           row.ID,
		ReviewerName: row.ReviewerName,
		Rating:       row.Rating,
		Comment:      row.Comment,
		CreatedAt:    row.CreatedAt,
	}
}

func optionalInt32(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	parsed := value.Int32
	return &parsed
}

func (a *API) submitPublicReview(c *gin.Context) {
	slug, reason := validateBookingPageSlug(c.Param("slug"))
	if reason != "" {
		fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
		return
	}
	page, err := a.svc.Q.GetPublicBookingPage(c.Request.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_page_not_found", "That booking page was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	var req struct {
		ReviewerName string `json:"reviewer_name" binding:"omitempty,max=120"`
		Rating       int16  `json:"rating" binding:"required,min=1,max=5"`
		Comment      string `json:"comment" binding:"max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_review", "Choose a rating from 1 to 5 and keep your review under 1,000 characters.")
		return
	}

	reviewerName := strings.TrimSpace(req.ReviewerName)
	var customerID *uuid.UUID
	if user, ok := currentUser(c); ok {
		customerID = &user.ID
		if reviewerName == "" {
			reviewerName = strings.TrimSpace(user.FullName)
		}
	}
	if len(reviewerName) < 2 {
		fail(c, http.StatusUnprocessableEntity, "reviewer_name_required", "Add your name so the provider knows who left the review.")
		return
	}

	review, err := a.svc.Q.CreateCompanyReview(c.Request.Context(), db.CreateCompanyReviewParams{
		CompanyID:    page.CompanyID,
		ReviewerName: reviewerName,
		Rating:       req.Rating,
		Comment:      strings.TrimSpace(req.Comment),
		CustomerID:   customerID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	summary, err := a.svc.Q.GetCompanyReviewSummary(c.Request.Context(), page.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    customerID,
		CompanyID:  &page.CompanyID,
		Action:     "company.review_submitted",
		EntityType: "company_review",
		EntityID:   &review.ID,
		Metadata:   map[string]any{"rating": review.Rating},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"review": publicReviewDTO{
			ID:           review.ID,
			ReviewerName: review.ReviewerName,
			Rating:       review.Rating,
			Comment:      review.Comment,
			CreatedAt:    review.CreatedAt,
		},
		"review_summary": gin.H{
			"review_count":   summary.ReviewCount,
			"average_rating": summary.AverageRating,
		},
	})
}
