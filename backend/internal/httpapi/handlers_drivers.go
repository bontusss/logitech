package httpapi

import (
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bol-platform/backend/internal/auth"
	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

type driverProfileDTO struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	VehiclePlate string    `json:"vehicle_plate"`
	Email        string    `json:"email"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func toDriverProfileDTO(driver db.ListDriverProfilesRow) driverProfileDTO {
	return driverProfileDTO{
		ID:           driver.ID,
		FullName:     driver.FullName,
		Phone:        driver.Phone,
		VehiclePlate: driver.VehiclePlate,
		Email:        driver.Email,
		Status:       driver.Status,
		CreatedAt:    driver.CreatedAt,
	}
}

func (a *API) listDrivers(c *gin.Context) {
	tenant, _ := currentTenant(c)
	drivers, err := a.svc.Q.ListDriverProfiles(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]driverProfileDTO, 0, len(drivers))
	for _, driver := range drivers {
		result = append(result, toDriverProfileDTO(driver))
	}
	c.JSON(http.StatusOK, gin.H{"drivers": result})
}

func (a *API) createDriver(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req struct {
		FullName     string `json:"full_name" binding:"required,min=2,max=120"`
		Phone        string `json:"phone" binding:"required,min=7,max=30"`
		VehiclePlate string `json:"vehicle_plate" binding:"required,min=2,max=20"`
		Email        string `json:"email" binding:"required,email,max=254"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Enter the driver's name, phone, vehicle plate number and valid email.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		fail(c, http.StatusBadRequest, "invalid_email", "Enter a valid driver email address.")
		return
	}
	rawToken, tokenHash, err := auth.GenerateOpaqueToken()
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if company.Status != "verified" {
		fail(c, http.StatusForbidden, "provider_not_approved", "Driver portals are available after your company is verified.")
		return
	}
	driver, err := a.svc.Q.CreateDriverProfile(c.Request.Context(), db.CreateDriverProfileParams{
		CompanyID:       tenant.CompanyID,
		FullName:        strings.TrimSpace(req.FullName),
		Phone:           strings.TrimSpace(req.Phone),
		VehiclePlate:    strings.ToUpper(strings.TrimSpace(req.VehiclePlate)),
		Email:           email,
		PortalTokenHash: tokenHash,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "driver_exists", "A driver with this email is already registered for your company.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	portalURL := strings.TrimRight(a.cfg.AppURL, "/") + "/rider/" + rawToken
	emailSent := true
	if err := a.svc.Mailer.SendDriverWelcome(c.Request.Context(), driver.Email, driver.FullName, company.Name, portalURL); err != nil {
		emailSent = false
		a.log.Warn("could not send driver welcome email", "error", err, "company_id", tenant.CompanyID, "driver_id", driver.ID)
	}
	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.driver_created",
		EntityType: "driver_profile",
		EntityID:   &driver.ID,
		Metadata:   map[string]any{"email": driver.Email, "vehicle_plate": driver.VehiclePlate},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusCreated, gin.H{
		"driver":     driverProfileDTO{ID: driver.ID, FullName: driver.FullName, Phone: driver.Phone, VehiclePlate: driver.VehiclePlate, Email: driver.Email, Status: driver.Status, CreatedAt: driver.CreatedAt},
		"portal_url": portalURL,
		"email_sent": emailSent,
	})
}

func (a *API) setDriverStatus(c *gin.Context) {
	tenant, _ := currentTenant(c)
	driverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_driver", "That driver id is invalid.")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required,oneof=active suspended"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_status", "Driver status must be active or suspended.")
		return
	}
	driver, err := a.svc.Q.SetDriverProfileStatus(c.Request.Context(), db.SetDriverProfileStatusParams{
		ID:        driverID,
		CompanyID: tenant.CompanyID,
		Status:    req.Status,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "driver_not_found", "That driver was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.driver_status_updated",
		EntityType: "driver_profile",
		EntityID:   &driver.ID,
		Metadata:   map[string]any{"status": driver.Status},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"driver": driver})
}

func (a *API) driverPortal(c *gin.Context) {
	token := c.Param("token")
	if len(token) != 43 {
		fail(c, http.StatusNotFound, "driver_portal_not_found", "This driver portal link is invalid or inactive.")
		return
	}
	driver, err := a.svc.Q.GetDriverProfileForPortal(c.Request.Context(), auth.HashToken(token))
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "driver_portal_not_found", "This driver portal link is invalid or inactive.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Robots-Tag", "noindex, nofollow, noarchive")
	c.JSON(http.StatusOK, gin.H{
		"driver": gin.H{
			"full_name":     driver.FullName,
			"vehicle_plate": driver.VehiclePlate,
			"company_name":  driver.CompanyName,
		},
		"active_assignment": nil,
		"location_tracking": "Location sharing is requested only during an active assigned delivery.",
	})
}
