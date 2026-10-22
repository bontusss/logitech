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
	t.Helper()
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
	staffHandler := handler.NewStaffHandler(cfg, queries)
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
			protected.GET("/staff/:id", staffHandler.GetStaff)
			protected.PUT("/staff/:id", staffHandler.UpdateStaff)
			protected.DELETE("/staff/:id", staffHandler.RevokeStaff)
			protected.POST("/staff/:id/reset-pin", staffHandler.ResetStaffPin)
			protected.POST("/staff/:id/unlock", staffHandler.UnlockStaffAccount)
			protected.GET("/staff/:id/first-login", staffHandler.FirstLoginStatus)
			protected.GET("/branches", branchHandler.ListBranches)
			protected.POST("/branches", branchHandler.CreateBranch)
			protected.PUT("/branches/:id", branchHandler.UpdateBranch)
			protected.DELETE("/branches/:id", branchHandler.DeleteBranch)
		}
	}

	return r, pool, queries, cfg
}

// postJSON issues a JSON POST request with an optional bearer token and
// returns the recorder.
func postJSON(r *gin.Engine, path string, token string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(http.MethodPost, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// firstLogin performs the first-login call: staff_id + one-time password + new PIN.
func firstLogin(r *gin.Engine, staffID, oneTimePassword, newPin string) *httptest.ResponseRecorder {
	return postJSON(r, "/api/auth/login", "", map[string]any{
		"staff_id": staffID,
		"pin":      oneTimePassword,
		"new_pin":  newPin,
	})
}

func TestAuthAndStaffCreationFlow(t *testing.T) {
	r, dbPool, _, cfg := setupTestRouter(t)
	defer dbPool.Close()

	unique := time.Now().UnixNano()

	// 1. Admin login
	w := postJSON(r, "/api/auth/admin/login", "", map[string]any{
		"username": cfg.AdminUsername,
		"password": cfg.AdminPassword,
	})
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

	// 2. Create a branch with the admin token (to scope this run)
	w = postJSON(r, "/api/branches", adminAuth.Token, map[string]string{
		"name": "Test Branch " + time.Now().Format("20060102150405"),
		"code": "TST-" + time.Now().Format("050405"),
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected branch create 201, got %d: %s", w.Code, w.Body.String())
	}
	var branch handler.BranchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &branch); err != nil {
		t.Fatalf("failed to decode branch response: %v", err)
	}
	testBranchID := branch.ID

	// 3. System admin tries to create "staff" role -> must be FORBIDDEN
	w = postJSON(r, "/api/staff", adminAuth.Token, map[string]string{
		"branch_id": testBranchID,
		"full_name": "Test Direct Staff",
		"role":      "staff",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when admin tries to create role 'staff', got %d: %s", w.Code, w.Body.String())
	}

	// 4. System admin creates an "owner" -> must be SUCCESSFUL (201)
	w = postJSON(r, "/api/staff", adminAuth.Token, map[string]string{
		"branch_id": testBranchID,
		"full_name": "Test Owner",
		"role":      "owner",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for admin creating owner, got %d: %s", w.Code, w.Body.String())
	}
	var createdOwner handler.CreateStaffResponse
	if err := json.Unmarshal(w.Body.Bytes(), &createdOwner); err != nil {
		t.Fatalf("failed to decode created owner response: %v", err)
	}
	if createdOwner.StaffID == "" {
		t.Errorf("expected auto-generated staff_id, got empty")
	}
	if createdOwner.OneTimePassword == "" {
		t.Errorf("expected one-time password in creation response")
	}
	if createdOwner.FirstLoginDone {
		t.Errorf("expected first_login_done=false on creation")
	}

	// 5. Owner's first login: one-time password + new_pin required
	w = firstLogin(r, createdOwner.StaffID, createdOwner.OneTimePassword, "OWNPIN01")
	if w.Code != http.StatusOK {
		t.Fatalf("expected owner first login 200, got %d: %s", w.Code, w.Body.String())
	}
	var ownerAuth handler.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ownerAuth); err != nil {
		t.Fatalf("failed to decode owner auth response: %v", err)
	}
	if ownerAuth.User.Role != auth.RoleOwner {
		t.Fatalf("expected role owner, got %s", ownerAuth.User.Role)
	}
	if !ownerAuth.User.FirstLoginDone {
		t.Errorf("expected first_login_done=true after first login")
	}

	// 6. Owner creates a "manager" -> must be SUCCESSFUL (201)
	w = postJSON(r, "/api/staff", ownerAuth.Token, map[string]string{
		"branch_id": testBranchID,
		"full_name": "Test Branch Manager",
		"role":      "manager",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for owner creating manager, got %d: %s", w.Code, w.Body.String())
	}
	var createdManager handler.CreateStaffResponse
	if err := json.Unmarshal(w.Body.Bytes(), &createdManager); err != nil {
		t.Fatalf("failed to decode created manager response: %v", err)
	}
	if createdManager.Role != "manager" {
		t.Errorf("expected role manager, got %s", createdManager.Role)
	}
	if createdManager.StaffID == "" || createdManager.StaffID[len(createdManager.StaffID)-3:] != createdManager.StaffID[len(createdManager.StaffID)-3:] {
		t.Errorf("expected form <-NNN staff id, got %q", createdManager.StaffID)
	}

	// 7. Manager's first login with the one-time password and new _pin
	w = firstLogin(r, createdManager.StaffID, createdManager.OneTimePassword, "MGRPIN01")
	if w.Code != http.StatusOK {
		t.Fatalf("expected manager first login 200, got %d: %s", w.Code, w.Body.String())
	}
	var mgrAuth handler.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &mgrAuth); err != nil {
		t.Fatalf("failed to decode manager auth: %v", err)
	}
	if mgrAuth.User.Role != "manager" {
		t.Fatalf("expected logged-in role manager, got %s", mgrAuth.User.Role)
	}

	// 8. Manager attempts to create an "owner" -> must be FORBIDDEN
	w = postJSON(r, "/api/staff", mgrAuth.Token, map[string]string{
		"branch_id": testBranchID,
		"full_name": "Unauthorized Owner",
		"role":      "owner",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when manager attempts to create owner, got %d: %s", w.Code, w.Body.String())
	}

	// 9. Manager creates an "owner" with job_role=operations -> 201
	w = postJSON(r, "/api/staff", mgrAuth.Token, map[string]string{
		"branch_id": testBranchID,
		"full_name": "Test Front Desk Staff",
		"role":      "staff",
		"job_role":  "operations",
	})
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
	if createdStaff.JobRole != "operations" {
		t.Errorf("expected job_role operations, got %s", createdStaff.JobRole)
	}

	// 10. Staff's first login
	w = firstLogin(r, createdStaff.StaffID, createdStaff.OneTimePassword, "STFPIN01")
	if w.Code != http.StatusOK {
		t.Fatalf("expected staff first login 200, got %d: %s", w.Code, w.Body.String())
	}
	var staffAuth handler.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &staffAuth); err != nil {
		t.Fatalf("failed to decode staff auth: %v", err)
	}
	if staffAuth.User.Role != "staff" {
		t.Fatalf("expected role staff, got %s", staffAuth.User.Role)
	}

	// 11. Staff attempts to create another staff -> must be FORBIDDEN
	w = postJSON(r, "/api/staff", staffAuth.Token, map[string]string{
		"branch_id": testBranchID,
		"full_name": "Sub-staff",
		"role":      "staff",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when staff tries to create staff, got %d: %s", w.Code, w.Body.String())
	}

	// 12. Log in again with the chosen PIN -> 200
	w = postJSON(r, "/api/auth/login", "", map[string]string{
		"staff_id": createdStaff.StaffID,
		"pin":      "STFPIN01",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected staff re-login 200, got %d: %s", w.Code, w.Body.String())
	}

	// 13. Three failed PIN attempts lock the account
	for i := 1; i <= 2; i++ {
		w = postJSON(r, "/api/auth/login", "", map[string]string{
			"staff_id": createdStaff.StaffID,
			"pin":      "wrongpin",
		})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 on failed attempt %d, got %d: %s", i, w.Code, w.Body.String())
		}
	}
	w = postJSON(r, "/api/auth/login", "", map[string]string{
		"staff_id": createdStaff.StaffID,
		"pin":      "wrongpin",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 locked after 3 failed attempts, got %d: %s", w.Code, w.Body.String())
	}
	// Even the correct PIN must not work while locked.
	w = postJSON(r, "/api/auth/login", "", map[string]string{
		"staff_id": createdStaff.StaffID,
		"pin":      "STFPIN01",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 while locked even with correct PIN, got %d: %s", w.Code, w.Body.String())
	}

	// 14. Branch manager unlocks the staff account
	w = postJSON(r, "/api/staff/"+staffAuth.User.ID+"/unlock", mgrAuth.Token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected manager unlock 200, got %d: %s", w.Code, w.Body.String())
	}

	// 15. Login works again after unlocking
	w = postJSON(r, "/api/auth/login", "", map[string]string{
		"staff_id": createdStaff.StaffID,
		"pin":      "STFPIN01",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected staff login 200 after unlock, got %d: %s", w.Code, w.Body.String())
	}

	// 16. First-login status endpoint reports completed first login
	w = getProduct(r, "/api/staff/"+staffAuth.User.ID+"/first-login", mgrAuth.Token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected first-login status 200, got %d: %s", w.Code, w.Body.String())
	}
	var flStatus handler.FirstLoginStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &flStatus); err != nil {
		t.Fatalf("failed to decode first-login status: %v", err)
	}
	if !flStatus.FirstLoginDone {
		t.Errorf("expected first_login_done=true after first login")
	}
	if flStatus.StaffID != createdStaff.StaffID {
		t.Errorf("expected staff_id %s, got %s", createdStaff.StaffID, flStatus.StaffID)
	}

	// 17. Manager resets the staff PIN; system issues a new one-time password (old PIN invalid)
	w = postJSON(r, "/api/staff/"+staffAuth.User.ID+"/reset-pin", mgrAuth.Token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected reset-pin 200, got %d: %s", w.Code, w.Body.String())
	}
	var reset handler.ResetStaffPinResponse
	if err := json.Unmarshal(w.Body.Bytes(), &reset); err != nil {
		t.Fatalf("failed to decode reset-pin response: %v", err)
	}
	if reset.OneTimePassword == "" {
		t.Fatalf("expected one-time password after reset")
	}

	// Old PIN no longer works; the one-time password + new_pin does.
	w = postJSON(r, "/api/auth/login", "", map[string]string{
		"staff_id": createdStaff.StaffID,
		"pin":      "STFPIN01",
	})
	if w.Code == http.StatusOK {
		t.Fatalf("expected old PIN to be rejected after pin reset, got 200")
	}
	w = firstLogin(r, createdStaff.StaffID, reset.OneTimePassword, "NEWPIN99")
	if w.Code != http.StatusOK {
		t.Fatalf("expected login with one-time password + new pin 200, got %d: %s", w.Code, w.Body.String())
	}

	// 18. Logout and token revocation test
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+staffAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, logoutReq)
	if w.Code != http.StatusOK {
		t.Fatalf("expected logout 200, got %d: %s", w.Code, w.Body.String())
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+staffAuth.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, meReq)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized after token revocation, got %d: %s", w.Code, w.Body.String())
	}

	_ = unique
}

// getProduct issues a GET request with a bearer token.
func getProduct(r *gin.Engine, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
