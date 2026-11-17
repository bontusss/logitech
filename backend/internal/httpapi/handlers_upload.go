package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

const maxImageBytes = 5 << 20 // 5 MiB

// uploadBrandingImage accepts logo, cover or favicon files and stores them under
// the configured upload directory, scoped by company id.
func (a *API) uploadBrandingImage(c *gin.Context) {
	tenant, _ := currentTenant(c)
	kind := strings.ToLower(strings.TrimSpace(c.PostForm("kind")))
	if kind != "logo" && kind != "cover" && kind != "favicon" {
		fail(c, http.StatusBadRequest, "invalid_kind", "Kind must be logo, cover or favicon.")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "file_required", "Choose an image to upload.")
		return
	}
	if fileHeader.Size > maxImageBytes {
		fail(c, http.StatusRequestEntityTooLarge, "file_too_large", "Images must be 5 MB or smaller.")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	defer file.Close()

	head := make([]byte, 512)
	read, _ := io.ReadFull(file, head)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		a.handleServiceError(c, err)
		return
	}

	contentType := http.DetectContentType(head[:read])
	extensions := map[string]string{
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/webp": ".webp",
	}
	extension, allowed := extensions[contentType]
	if !allowed {
		fail(c, http.StatusUnsupportedMediaType, "unsupported_image", "Use a PNG, JPEG or WebP image.")
		return
	}

	directory := filepath.Join(a.cfg.UploadDir, tenant.CompanyID.String())
	if err := os.MkdirAll(directory, 0o755); err != nil {
		a.handleServiceError(c, err)
		return
	}

	filename := fmt.Sprintf("%s-%s%s", kind, uuid.NewString(), extension)
	destination := filepath.Join(directory, filename)
	target, err := os.Create(destination)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	defer target.Close()

	written, err := io.Copy(target, io.LimitReader(file, maxImageBytes+1))
	if err != nil {
		_ = os.Remove(destination)
		a.handleServiceError(c, err)
		return
	}
	if written > maxImageBytes {
		_ = os.Remove(destination)
		fail(c, http.StatusRequestEntityTooLarge, "file_too_large", "Images must be 5 MB or smaller.")
		return
	}

	publicURL := strings.TrimRight(a.cfg.APIPublicURL, "/") + "/uploads/" +
		tenant.CompanyID.String() + "/" + filename

	current, err := a.svc.Q.GetCompanyBranding(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	logo, cover, favicon := deref(current.LogoUrl), deref(current.CoverImageUrl), deref(current.FaviconUrl)
	switch kind {
	case "logo":
		logo = publicURL
	case "cover":
		cover = publicURL
	case "favicon":
		favicon = publicURL
	}

	branding, err := a.svc.Q.UpdateCompanyBranding(c.Request.Context(), db.UpdateCompanyBrandingParams{
		CompanyID:      tenant.CompanyID,
		Column2:        logo,
		PrimaryColor:   current.PrimaryColor,
		SecondaryColor: current.SecondaryColor,
		Column5:        favicon,
		Column6:        cover,
		Column7:        deref(current.PublicName),
		Column8:        deref(current.Description),
		Column9:        deref(current.PwaShortName),
	})
	if err != nil {
		_ = os.Remove(destination)
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.branding_image_uploaded",
		EntityType: "company_branding",
		Metadata:   map[string]any{"kind": kind, "url": publicURL},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"branding": branding})
}
