package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/db/sqlc"
	"backend/internal/handler"
	"backend/internal/redis"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *sql.DB, *sqlc.Queries, *config.Config) {
	gin.SetMode(gin.TestMode)

	cfg := config.Load()
	pool, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		t.Skip("skipping integration test: database unavailable", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.Ping(ctx, pool); err != nil {
		t.Skip("skipping integration test: database ping failed", err)
	}

	queries := sqlc.New(pool)
	store := redis.NewStore(cfg.RedisURL)

	authHandler := handler.NewAuthHandler(cfg, queries, store)
	staffHandler := handler.NewStaffHandler(queries)
	branchHandler := handler.NewBranchHandler(queries)

	r := gin.New()

	api := r.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/admin/login", authHandler.AdminLogin)

		protected := api.Group("")
		protected.Use(auth.AuthRequired(cfg.JWTSecret, store))
		{
			protected.GET("/auth/me", authHandler.GetMe)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.POST("/staff", staffHandler.CreateStaff)
			protected.GET("/staff", staffHandler.ListStaff)
			protected.GET("/branches", branchHandler.ListBranches)
		}
	}

	return r, pool, queries, cfg
}

func TestAuthAndStaffCreationFlow(t *testing.T) {
	r, dbPool, _, cfg := setupTestRouter(t)
	defer dbPool.Close()

	// 1. We need admin token first to query branches
	adminLoginBody, _ := json.Marshal(map[string]string{
		"username": cfg.AdminUsername,
		"password": cfg.AdminPassword,
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/admin/login", bytes.NewReader(adminLoginBody)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected admin login 200, got %d: %s", w.Code, w.Body.String())
	}

	var adminAuth handler.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &adminAuth); err != nil {
		t.Fatalf("failed to decode admin auth response: %v", err)
	}
	if adminAuth.User.Role != auth.RoleSystemAdmin {
		t.Fatalf("expected role system_admin, got %s", adminAuth.User.Role)
	}

	// 2. Query branches using admin token
	branchReq := httptest.NewRequest(http.MethodGet, "/api/branches", nil)
	branchReq.Header.Set("Authorization", "Bearer "+adminAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, branchReq)
	if w.Code != http.StatusOK {
		t.Fatalf("expected branches 200, got %d: %s", w.Code, w.Body.String())
	}

	var branches []handler.BranchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &branches); err != nil || len(branches) == 0 {
		t.Fatalf("expected at least 1 branch, got: %v", branches)
	}
	testBranchID := branches[0].ID

	// 3. System admin tries to create "staff" role -> must be FORBIDDEN
	createStaffByAdminBody, _ := json.Marshal(map[string]string{
		"branch_id": testBranchID,
		"full_name": "Test Direct Staff",
		"role":      "staff",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/staff", bytes.NewReader(createStaffByAdminBody))
	req.Header.Set("Authorization", "Bearer "+adminAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when admin tries to create role 'staff', got %d: %s", w.Code, w.Body.String())
	}

	// 4. System admin creates a "manager" -> must be SUCCESSFUL (201)
	createManagerBody, _ := json.Marshal(map[string]string{
		"branch_id": testBranchID,
		"full_name": "Test Branch Manager",
		"role":      "manager",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/staff", bytes.NewReader(createManagerBody))
	req.Header.Set("Authorization", "Bearer "+adminAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for admin creating manager, got %d: %s", w.Code, w.Body.String())
	}

	var createdManager handler.CreateStaffResponse
	if err := json.Unmarshal(w.Body.Bytes(), &createdManager); err != nil {
		t.Fatalf("failed to decode created manager response: %v", err)
	}
	if createdManager.Role != "manager" {
		t.Errorf("expected role manager, got %s", createdManager.Role)
	}
	if createdManager.StaffID == "" {
		t.Errorf("expected auto-generated staff_id, got empty")
	}
	if len(createdManager.Pin) != 4 {
		t.Errorf("expected 4-digit alphanumeric PIN, got %q (len %d)", createdManager.Pin, len(createdManager.Pin))
	}

	// 5. Newly created Manager logs in with staff_id and PIN
	mgrLoginBody, _ := json.Marshal(map[string]string{
		"staff_id": createdManager.StaffID,
		"pin":      createdManager.Pin,
	})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(mgrLoginBody)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected manager login 200, got %d: %s", w.Code, w.Body.String())
	}

	var mgrAuth handler.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &mgrAuth); err != nil {
		t.Fatalf("failed to decode manager auth: %v", err)
	}
	if mgrAuth.User.Role != "manager" {
		t.Fatalf("expected logged-in role manager, got %s", mgrAuth.User.Role)
	}

	// 6. Manager attempts to create an "owner" -> must be FORBIDDEN
	badRoleByMgr, _ := json.Marshal(map[string]string{
		"branch_id": testBranchID,
		"full_name": "Unauthorized Owner",
		"role":      "owner",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/staff", bytes.NewReader(badRoleByMgr))
	req.Header.Set("Authorization", "Bearer "+mgrAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when manager attempts to create owner, got %d: %s", w.Code, w.Body.String())
	}

	// 7. Manager creates a "staff" -> must be SUCCESSFUL (201)
	createStaffByMgr, _ := json.Marshal(map[string]string{
		"branch_id": testBranchID,
		"full_name": "Test Front Desk Staff",
		"role":      "staff",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/staff", bytes.NewReader(createStaffByMgr))
	req.Header.Set("Authorization", "Bearer "+mgrAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created when manager creates staff, got %d: %s", w.Code, w.Body.String())
	}

	var createdStaff handler.CreateStaffResponse
	if err := json.Unmarshal(w.Body.Bytes(), &createdStaff); err != nil {
		t.Fatalf("failed to decode created staff: %v", err)
	}
	if createdStaff.Role != "staff" {
		t.Errorf("expected role staff, got %s", createdStaff.Role)
	}
	if len(createdStaff.Pin) != 4 {
		t.Errorf("expected 4-digit PIN, got %s", createdStaff.Pin)
	}

	// 8. Newly created Staff logs in with staff_id and PIN
	staffLoginBody, _ := json.Marshal(map[string]string{
		"staff_id": createdStaff.StaffID,
		"pin":      createdStaff.Pin,
	})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(staffLoginBody)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected staff login 200, got %d: %s", w.Code, w.Body.String())
	}

	var staffAuth handler.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &staffAuth); err != nil {
		t.Fatalf("failed to decode staff auth: %v", err)
	}
	if staffAuth.User.Role != "staff" {
		t.Fatalf("expected role staff, got %s", staffAuth.User.Role)
	}

	// 9. Staff attempts to create another staff -> must be FORBIDDEN
	createByStaff, _ := json.Marshal(map[string]string{
		"branch_id": testBranchID,
		"full_name": "Sub-staff",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/staff", bytes.NewReader(createByStaff))
	req.Header.Set("Authorization", "Bearer "+staffAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when staff tries to create staff, got %d: %s", w.Code, w.Body.String())
	}

	// 10. Logout and Token Revocation test
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+staffAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, logoutReq)
	if w.Code != http.StatusOK {
		t.Fatalf("expected logout 200, got %d: %s", w.Code, w.Body.String())
	}

	// Using the revoked token should now fail
	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+staffAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, meReq)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized after token revocation, got %d: %s", w.Code, w.Body.String())
	}
}
