package handler

import (
	"database/sql"
	"net/http"
	"time"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/db/sqlc"
	"backend/internal/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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
type LoginRequest struct {
	StaffID  string `json:"staff_id" example:"STF-001"`
	Pin      string `json:"pin" example:"1001"`
	Username string `json:"username,omitempty" example:"admin"`
	Password string `json:"password,omitempty" example:"adminpassword123"`
}

// AdminLoginRequest defines credentials specifically for system admin login.
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"adminpassword123"`
}

// StaffProfile represents the public user info returned in auth responses.
type StaffProfile struct {
	ID        string `json:"id" example:"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"`
	StaffID   string `json:"staff_id" example:"STF-001"`
	FullName  string `json:"full_name" example:"Chinedu Okafor"`
	Role      string `json:"role" example:"staff"`
	BranchID  string `json:"branch_id,omitempty" example:"11111111-1111-4111-8111-111111111111"`
	Phone     string `json:"phone,omitempty" example:"08031110001"`
	Email     string `json:"email,omitempty" example:"staff@logitech.example"`
	Status    string `json:"status" example:"active"`
	CreatedAt string `json:"created_at,omitempty"`
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

// Login godoc
// @Summary Staff and Admin Login
// @Description Authenticate staff using staff_id and 4-digit alphanumeric PIN, or system admin using username and password.
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

	staff, err := h.queries.GetStaffByStaffID(c.Request.Context(), sql.NullString{String: req.StaffID, Valid: true})
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid staff_id or pin"})
		return
	}

	if staff.Status != sqlc.RecordStatusActive {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "account is inactive, please contact management"})
		return
	}

	// Verify PIN: first check plain pin match, or bcrypt check against password_hash
	pinMatches := staff.Pin == req.Pin
	if !pinMatches && staff.PasswordHash != "" {
		if bcryptErr := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(req.Pin)); bcryptErr == nil {
			pinMatches = true
		}
	}

	if !pinMatches {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid staff_id or pin"})
		return
	}

	ttl := 24 * time.Hour
	token, err := auth.GenerateToken(
		staff.ID.String(),
		staff.StaffID.String,
		staff.FullName,
		string(staff.Role),
		staff.BranchID.String(),
		h.cfg.JWTSecret,
		ttl,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int64(ttl.Seconds()),
		User: StaffProfile{
			ID:        staff.ID.String(),
			StaffID:   staff.StaffID.String,
			FullName:  staff.FullName,
			Role:      string(staff.Role),
			BranchID:  staff.BranchID.String(),
			Phone:     staff.Phone.String,
			Email:     staff.Email.String,
			Status:    string(staff.Status),
			CreatedAt: staff.CreatedAt.Format(time.RFC3339),
		},
	})
}

// AdminLogin godoc
// @Summary System Admin Login
// @Description Authenticate as system administrator using credentials configured in environment.
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
			ID:       "admin",
			StaffID:  "SYS-ADMIN",
			FullName: "System Administrator",
			Role:     auth.RoleSystemAdmin,
			Email:    h.cfg.AdminEmail,
			Status:   "active",
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
			ID:       "admin",
			StaffID:  "SYS-ADMIN",
			FullName: "System Administrator",
			Role:     auth.RoleSystemAdmin,
			Email:    h.cfg.AdminEmail,
			Status:   "active",
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

	c.JSON(http.StatusOK, StaffProfile{
		ID:        staff.ID.String(),
		StaffID:   staff.StaffID.String,
		FullName:  staff.FullName,
		Role:      string(staff.Role),
		BranchID:  staff.BranchID.String(),
		Phone:     staff.Phone.String,
		Email:     staff.Email.String,
		Status:    string(staff.Status),
		CreatedAt: staff.CreatedAt.Format(time.RFC3339),
	})
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
