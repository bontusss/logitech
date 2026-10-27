package handler

import (
	"database/sql"
	"net/http"
	"time"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/db/sqlc"
	"backend/internal/generator"
	"backend/internal/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// maxFailedLoginAttempts is the number of failed login attempts before the
// account is locked. A locked account can be reopened by the manager of the
// staff's branch, or by the owner across all branches.
const maxFailedLoginAttempts = 3

// AuthHandler handles authentication requests.
type AuthHandler struct {
	cfg     *config.Config
	queries *sqlc.Queries
	store   *redis.Store
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(cfg *config.Config, queries *sqlc.Queries, store *redis.Store) *AuthHandler {
	return &AuthHandler{
		cfg:     cfg,
		queries: queries,
		store:   store,
	}
}

// LoginRequest defines credentials for staff or admin login.
//
// For regular users, "pin" holds either their personal PIN or, before the PIN
// has been set (first login / after a PIN reset), the one-time password issued
// by the system. On first login "new_pin" is required; it becomes the user's
// personal PIN (max 10 chars).
type LoginRequest struct {
	StaffID string `json:"staff_id" example:"GAL-001"`
	Pin     string `json:"pin"      example:"demo1234"`
	// NewPin is required on first login (and after a PIN reset) to create the personal PIN.
	NewPin   string `json:"new_pin,omitempty" example:"newpin99"`
	Username string `json:"username,omitempty" example:"admin"`
	Password string `json:"password,omitempty" example:"adminpassword123"`
}

type ChangePinRequest struct {
	CurrentPin string `json:"current_pin" binding:"required"`
	NewPin     string `json:"new_pin" binding:"required"`
}

// AdminLoginRequest defines credentials specifically for system admin login.
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"adminpassword123"`
}

// StaffProfile represents the public user info returned in auth responses.
type StaffProfile struct {
	ID             string `json:"id" example:"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"`
	StaffID        string `json:"staff_id" example:"GAL-001"`
	FullName       string `json:"full_name" example:"Chinedu Okafor"`
	Role           string `json:"role" example:"staff"`
	JobRole        string `json:"job_role,omitempty" example:"operations"`
	BranchID       string `json:"branch_id,omitempty" example:"11111111-1111-4111-8111-111111111111"`
	Phone          string `json:"phone,omitempty" example:"08031110001"`
	Email          string `json:"email,omitempty" example:"staff@logitech.example"`
	Status         string `json:"status" example:"active"`
	FirstLoginDone bool   `json:"first_login_done" example:"true"`
	Locked         bool   `json:"locked" example:"false"`
	CreatedAt      string `json:"created_at,omitempty"`
}

// AuthResponse represents the response with the generated JWT token and user info.
type AuthResponse struct {
	Token     string       `json:"token"`
	TokenType string       `json:"token_type" example:"Bearer"`
	ExpiresIn int64        `json:"expires_in" example:"86400"`
	User      StaffProfile `json:"user"`
}

// ErrorResponse represents an error response payload.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}

// newStaffProfile maps a staff row onto the public profile representation.
// Owners are not attached to any single branch, so their BranchID is left empty
// even when the underlying row carries one (e.g. from seed data).
func newStaffProfile(s sqlc.Staff) StaffProfile {
	branchID := s.BranchID.String()
	if s.Role == sqlc.StaffRoleOwner {
		branchID = ""
	}
	return StaffProfile{
		ID:             s.ID.String(),
		StaffID:        s.StaffID,
		FullName:       s.FullName,
		Role:           string(s.Role),
		JobRole:        jobRoleString(s.JobRole),
		BranchID:       branchID,
		Phone:          s.Phone.String,
		Email:          s.Email.String,
		Status:         string(s.Status),
		FirstLoginDone: s.FirstLoginDone,
		Locked:         s.LockedAt.Valid,
		CreatedAt:      s.CreatedAt.Format(time.RFC3339),
	}
}

// Login godoc
// @Summary Staff Login
// @Description Authenticate a staff member using their `staff_id` and `pin`.
// @Description
// @Description ### Normal login (PIN already set)
// @Description Send `staff_id` + `pin` only.
// @Description
// @Description ### First login / after a PIN reset
// @Description 1. Call this endpoint with `staff_id` + `pin` (the one-time password).
// @Description    The API responds 400 with `"provide new_pin"` in the error message.
// @Description 2. Call again with `staff_id` + `pin` (OTP) + `new_pin` (4-10 chars).
// @Description    The OTP is consumed; `new_pin` becomes the permanent PIN.
// @Description
// @Description Account locks after **3** failed attempts; a branch manager or owner must unlock it.
// @Description
// @Description ---
// @Description ### Seed / demo credentials (PIN = `demo1234` for all)
// @Description
// @Description | staff_id | Full name      | Role    |
// @Description |----------|----------------|---------|
// @Description | GAL-001  | Chinedu Okafor | owner   |
// @Description | GAL-002  | Ada Nwosu      | manager |
// @Description | GAL-003  | Tunde Bello    | staff   |
// @Description | GAL-004  | Mary James     | staff   |
// @Description
// @Description GAL-014 and GAL-015 have `first_login_done = false` (first-login flow required).
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	// 1. Check if system admin credentials were provided
	if (req.Username != "" || req.StaffID == h.cfg.AdminUsername) && req.Password != "" {
		if (req.Username == h.cfg.AdminUsername || req.Username == h.cfg.AdminEmail || req.StaffID == h.cfg.AdminUsername) &&
			req.Password == h.cfg.AdminPassword {
			h.issueAdminToken(c)
			return
		}
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid admin credentials"})
		return
	}

	// 2. Staff login requires staff_id and pin
	if req.StaffID == "" || req.Pin == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "staff_id and pin are required for staff login"})
		return
	}

	staff, err := h.queries.GetStaffByStaffID(c.Request.Context(), req.StaffID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid staff_id or pin"})
		return
	}

	if staff.Status != sqlc.RecordStatusActive {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "account is inactive, please contact management"})
		return
	}

	// A locked account cannot log in; it must be reopened by the branch
	// manager or the owner.
	if staff.LockedAt.Valid {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "account is locked after too many failed login attempts; contact your branch manager or the owner to unlock it"})
		return
	}

	// 3. Pin-reset pending (first login or after reset): the provided secret
	// must match the outstanding one-time password, and a new PIN is required.
	otpPending := !staff.FirstLoginDone || staff.OtpHash.Valid
	if otpPending {
		if !compareOtpHash(staff.OtpHash, req.Pin) {
			c.JSON(h.handleFailedLogin(c, staff.ID), ErrorResponse{Error: "invalid staff_id or one-time password"})
			return
		}

		// OTP verified. If no new_pin was supplied (sign-in page), ask for it.
		if req.NewPin == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "your one-time password was accepted for a first login; provide new_pin (4-10 characters) to create your PIN"})
			return
		}

		if !generator.ValidatePin(req.NewPin) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "new_pin must be between 4 and 10 characters"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPin), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to hash credentials"})
			return
		}

		updated, err := h.queries.CompleteFirstLogin(c.Request.Context(), sqlc.CompleteFirstLoginParams{
			ID:      staff.ID,
			PinHash: sql.NullString{String: string(hash), Valid: true},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save your PIN: " + err.Error()})
			return
		}
		staff = updated
	} else {
		// 4. Normal login: verify against the saved PIN hash.
		if staff.PinHash.Valid && bcrypt.CompareHashAndPassword([]byte(staff.PinHash.String), []byte(req.Pin)) == nil {
			if staff.FailedLoginAttempts > 0 {
				_, _ = h.queries.ResetLoginState(c.Request.Context(), staff.ID)
			}
		} else if staff.OtpHash.Valid && compareOtpHash(staff.OtpHash, req.Pin) {
			// Valid one-time password after a PIN reset that already had a first login.
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "your one-time password was accepted; provide new_pin (4-10 characters) to create a new PIN"})
			return
		} else {
			c.JSON(h.handleFailedLogin(c, staff.ID), ErrorResponse{Error: "invalid staff_id or pin"})
			return
		}
	}

	token, err := auth.GenerateToken(
		staff.ID.String(),
		staff.StaffID,
		staff.FullName,
		string(staff.Role),
		jobRoleString(staff.JobRole),
		staff.BranchID.String(),
		h.cfg.JWTSecret,
		24*time.Hour,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int64((24 * time.Hour).Seconds()),
		User:      newStaffProfile(staff),
	})
}

// jobRoleString renders an optional staff job role as a plain string
// (empty for owner/manager accounts).
func jobRoleString(nr sqlc.NullStaffJobRole) string {
	if !nr.Valid {
		return ""
	}
	return string(nr.StaffJobRole)
}

// compareOtpHash compares a clear-text one-time password against its bcrypt hash.
func compareOtpHash(hash sql.NullString, secret string) bool {
	if !hash.Valid || hash.String == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(secret)) == nil
}

// handleFailedLogin records a failed attempt and locks the account after 3
// failures. It returns the HTTP status to respond with.
func (h *AuthHandler) handleFailedLogin(c *gin.Context, staffID uuid.UUID) int {
	row, err := h.queries.RecordFailedLogin(c.Request.Context(), sqlc.RecordFailedLoginParams{
		ID:                  staffID,
		FailedLoginAttempts: maxFailedLoginAttempts,
	})
	if err != nil {
		return http.StatusUnauthorized
	}
	if row.LockedAt.Valid {
		return http.StatusForbidden
	}
	return http.StatusUnauthorized
}

// AdminLogin godoc
// @Summary System Admin Login
// @Description Authenticate as system administrator using credentials configured in environment.
// @Description
// @Description ### Demo credentials
// @Description
// @Description | Field    | Value            |
// @Description |----------|------------------|
// @Description | username | admin            |
// @Description | password | adminpassword123 |
// @Tags auth
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "Admin Credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/admin/login [post]
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	if (req.Username == h.cfg.AdminUsername || req.Username == h.cfg.AdminEmail) && req.Password == h.cfg.AdminPassword {
		h.issueAdminToken(c)
		return
	}

	c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid admin username or password"})
}

func (h *AuthHandler) issueAdminToken(c *gin.Context) {
	ttl := 24 * time.Hour
	token, err := auth.GenerateToken(
		"admin",
		"SYS-ADMIN",
		"System Administrator",
		auth.RoleSystemAdmin,
		"",
		"",
		h.cfg.JWTSecret,
		ttl,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate admin token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int64(ttl.Seconds()),
		User: StaffProfile{
			ID:             "admin",
			StaffID:        "SYS-ADMIN",
			FullName:       "System Administrator",
			Role:           auth.RoleSystemAdmin,
			Email:          h.cfg.AdminEmail,
			Status:         "active",
			FirstLoginDone: true,
		},
	})
}

// GetMe godoc
// @Summary Get Current Authenticated User Profile
// @Description Returns the profile and role details of the currently authenticated user.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} StaffProfile
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	if claims.Role == auth.RoleSystemAdmin {
		c.JSON(http.StatusOK, StaffProfile{
			ID:             "admin",
			StaffID:        "SYS-ADMIN",
			FullName:       "System Administrator",
			Role:           auth.RoleSystemAdmin,
			Email:          h.cfg.AdminEmail,
			Status:         "active",
			FirstLoginDone: true,
		})
		return
	}

	staffUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id in token claims"})
		return
	}

	staff, err := h.queries.GetStaffByID(c.Request.Context(), staffUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "staff record not found"})
		return
	}

	c.JSON(http.StatusOK, newStaffProfile(staff))
}

// ChangePin godoc
// @Summary Change the authenticated user's PIN
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePinRequest true "Current and new PIN"
// @Success 200 {object} map[string]string
// @Router /api/auth/change-pin [post]
func (h *AuthHandler) ChangePin(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "system administrator credentials cannot be changed here"})
		return
	}
	var req ChangePinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "current_pin and new_pin are required"})
		return
	}
	if !generator.ValidatePin(req.NewPin) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "new_pin must be between 4 and 10 characters"})
		return
	}
	if req.CurrentPin == req.NewPin {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "new_pin must be different from current_pin"})
		return
	}

	staff, err := h.queries.GetStaffByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "account not found"})
		return
	}
	if staff.Status != sqlc.RecordStatusActive || staff.LockedAt.Valid {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "inactive or locked accounts cannot change their PIN"})
		return
	}
	if !staff.FirstLoginDone || !staff.PinHash.Valid || staff.OtpHash.Valid {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "complete the one-time password sign-in flow before changing your PIN"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(staff.PinHash.String), []byte(req.CurrentPin)) != nil {
		c.JSON(h.handleFailedLogin(c, staff.ID), ErrorResponse{Error: "current PIN is incorrect"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPin), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to hash new PIN"})
		return
	}
	if _, err := h.queries.ChangeOwnPin(c.Request.Context(), sqlc.ChangeOwnPinParams{
		ID:      staff.ID,
		PinHash: sql.NullString{String: string(hash), Valid: true},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to change PIN: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "PIN updated"})
}

// Logout godoc
// @Summary User Logout
// @Description Invalidate the current session token using Redis token revocation.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token, ok := auth.GetToken(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	if h.store != nil {
		_ = h.store.BlacklistToken(c.Request.Context(), token, 24*time.Hour)
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}
