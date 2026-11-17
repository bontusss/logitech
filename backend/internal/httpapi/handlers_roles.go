package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

type roleDTO struct {
	ID          uuid.UUID  `json:"id"`
	CompanyID   *uuid.UUID `json:"company_id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	IsSystem    bool       `json:"is_system"`
	Permissions []string   `json:"permissions"`
}

// requesterPermissions returns the permissions the current actor may grant.
// Platform administrators can grant any catalogued permission.
func (a *API) requesterPermissions(c *gin.Context) map[string]bool {
	result := map[string]bool{}
	user, ok := currentUser(c)
	if !ok {
		return result
	}
	tenant, _ := currentTenant(c)

	if tenant.IsPlatform && user.PlatformRole != nil {
		permissions, _ := a.svc.Q.ListPermissions(c.Request.Context())
		for _, permission := range permissions {
			result[permission.Key] = true
		}
		return result
	}

	keys, _ := a.svc.Q.ListUserPermissions(c.Request.Context(), db.ListUserPermissionsParams{
		UserID:    user.ID,
		CompanyID: tenant.CompanyID,
	})
	for _, key := range keys {
		result[key] = true
	}
	return result
}

func (a *API) canGrantRole(c *gin.Context, roleID uuid.UUID) bool {
	keys, err := a.svc.Q.ListRolePermissionKeys(c.Request.Context(), roleID)
	if err != nil {
		return false
	}
	have := a.requesterPermissions(c)
	for _, key := range keys {
		if !have[key] {
			return false
		}
	}
	return true
}

func (a *API) ensureRoleGrantable(c *gin.Context, role db.Role) bool {
	if role.Key == "owner" {
		fail(c, http.StatusForbidden, "owner_role_forbidden", "The owner role cannot be granted to another member.")
		return false
	}
	if !a.canGrantRole(c, role.ID) {
		fail(c, http.StatusForbidden, "role_exceeds_authority", "You cannot grant permissions you do not hold yourself.")
		return false
	}
	return true
}

// resolveAssignableRole finds a system or company-owned role by key and
// enforces that the actor may grant its permissions.
func (a *API) resolveAssignableRole(c *gin.Context, key string) (db.Role, bool) {
	tenant, _ := currentTenant(c)
	role, err := a.svc.Q.GetRoleByCompanyAndKey(c.Request.Context(), db.GetRoleByCompanyAndKeyParams{
		Key:       key,
		CompanyID: &tenant.CompanyID,
	})
	if err != nil {
		fail(c, http.StatusUnprocessableEntity, "invalid_role", "That role does not exist for this company.")
		return db.Role{}, false
	}
	if !a.ensureRoleGrantable(c, role) {
		return db.Role{}, false
	}
	return role, true
}

func (a *API) validatePermissionKeys(c *gin.Context, requested []string) ([]string, bool) {
	catalog, err := a.svc.Q.ListPermissions(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return nil, false
	}
	known := make(map[string]bool, len(catalog))
	for _, permission := range catalog {
		known[permission.Key] = true
	}
	have := a.requesterPermissions(c)

	unique := make([]string, 0, len(requested))
	seen := map[string]bool{}
	for _, key := range requested {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			continue
		}
		if !known[key] {
			fail(c, http.StatusUnprocessableEntity, "unknown_permission", "One of the selected permissions does not exist.")
			return nil, false
		}
		if !have[key] {
			fail(c, http.StatusForbidden, "permission_exceeds_authority", "You cannot grant a permission you do not hold yourself.")
			return nil, false
		}
		seen[key] = true
		unique = append(unique, key)
	}
	if len(unique) == 0 {
		fail(c, http.StatusUnprocessableEntity, "permissions_required", "Select at least one permission for the role.")
		return nil, false
	}
	return unique, true
}

func (a *API) listRoles(c *gin.Context) {
	tenant, _ := currentTenant(c)
	rows, err := a.svc.Q.ListRolesForCompany(c.Request.Context(), &tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	roles := make([]roleDTO, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, roleDTO{
			ID:          row.ID,
			CompanyID:   row.CompanyID,
			Key:         row.Key,
			Name:        row.Name,
			Description: row.Description,
			IsSystem:    row.IsSystem,
			Permissions: nonNilStrings(row.Permissions),
		})
	}

	permissions, err := a.svc.Q.ListPermissions(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles, "permissions": permissions})
}

func (a *API) createRole(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req struct {
		Name        string   `json:"name" binding:"required,min=2,max=80"`
		Description string   `json:"description" binding:"omitempty,max=300"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Give the role a name and choose its permissions.")
		return
	}

	permissions, ok := a.validatePermissionKeys(c, req.Permissions)
	if !ok {
		return
	}

	key := "custom_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
	var role db.Role
	err := pgx.BeginFunc(c.Request.Context(), a.svc.DB, func(tx pgx.Tx) error {
		q := a.svc.Q.WithTx(tx)
		created, err := q.CreateCompanyRole(c.Request.Context(), db.CreateCompanyRoleParams{
			CompanyID: &tenant.CompanyID,
			Key:       key,
			Name:      strings.TrimSpace(req.Name),
			Column4:   strings.TrimSpace(req.Description),
		})
		if err != nil {
			return err
		}
		role = created
		for _, permission := range permissions {
			if err := q.AddRolePermission(c.Request.Context(), db.AddRolePermissionParams{
				RoleID: created.ID,
				Key:    permission,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.role_created",
		EntityType: "role",
		EntityID:   &role.ID,
		Metadata:   map[string]any{"name": role.Name, "permissions": permissions},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusCreated, gin.H{"role": roleDTO{
		ID: role.ID, CompanyID: role.CompanyID, Key: role.Key, Name: role.Name,
		Description: role.Description, IsSystem: role.IsSystem, Permissions: permissions,
	}})
}

func (a *API) updateRole(c *gin.Context) {
	tenant, _ := currentTenant(c)
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_role", "That role id is invalid.")
		return
	}

	role, err := a.svc.Q.GetCompanyRole(c.Request.Context(), db.GetCompanyRoleParams{
		ID:        roleID,
		CompanyID: &tenant.CompanyID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "role_not_found", "That role was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	if role.IsSystem || role.CompanyID == nil || *role.CompanyID != tenant.CompanyID {
		fail(c, http.StatusForbidden, "system_role_readonly", "System roles cannot be edited.")
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required,min=2,max=80"`
		Description string   `json:"description" binding:"omitempty,max=300"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Give the role a name and choose its permissions.")
		return
	}

	permissions, ok := a.validatePermissionKeys(c, req.Permissions)
	if !ok {
		return
	}

	err = pgx.BeginFunc(c.Request.Context(), a.svc.DB, func(tx pgx.Tx) error {
		q := a.svc.Q.WithTx(tx)
		updated, err := q.UpdateCompanyRole(c.Request.Context(), db.UpdateCompanyRoleParams{
			ID:        role.ID,
			CompanyID: &tenant.CompanyID,
			Name:      strings.TrimSpace(req.Name),
			Column4:   strings.TrimSpace(req.Description),
		})
		if err != nil {
			return err
		}
		role = updated
		if err := q.DeleteRolePermissions(c.Request.Context(), role.ID); err != nil {
			return err
		}
		for _, permission := range permissions {
			if err := q.AddRolePermission(c.Request.Context(), db.AddRolePermissionParams{
				RoleID: role.ID,
				Key:    permission,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.role_updated",
		EntityType: "role",
		EntityID:   &role.ID,
		Metadata:   map[string]any{"permissions": permissions},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"role": roleDTO{
		ID: role.ID, CompanyID: role.CompanyID, Key: role.Key, Name: role.Name,
		Description: role.Description, IsSystem: role.IsSystem, Permissions: permissions,
	}})
}

func (a *API) deleteRole(c *gin.Context) {
	tenant, _ := currentTenant(c)
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_role", "That role id is invalid.")
		return
	}

	role, err := a.svc.Q.GetCompanyRole(c.Request.Context(), db.GetCompanyRoleParams{
		ID:        roleID,
		CompanyID: &tenant.CompanyID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "role_not_found", "That role was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	if role.IsSystem || role.CompanyID == nil || *role.CompanyID != tenant.CompanyID {
		fail(c, http.StatusForbidden, "system_role_readonly", "System roles cannot be deleted.")
		return
	}

	assignments, err := a.svc.Q.CountRoleAssignments(c.Request.Context(), role.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if assignments > 0 {
		fail(c, http.StatusConflict, "role_in_use", "Move members off this role before deleting it.")
		return
	}

	if err := a.svc.Q.DeleteCompanyRole(c.Request.Context(), db.DeleteCompanyRoleParams{
		ID:        role.ID,
		CompanyID: &tenant.CompanyID,
	}); err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.role_deleted",
		EntityType: "role",
		EntityID:   &role.ID,
		Metadata:   map[string]any{"name": role.Name},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.Status(http.StatusNoContent)
}
