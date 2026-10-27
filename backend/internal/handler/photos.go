package handler

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// PhotoHandler handles manifest photo uploads captured from the camera.
type PhotoHandler struct{}

// NewPhotoHandler constructs a PhotoHandler.
func NewPhotoHandler() *PhotoHandler {
	return &PhotoHandler{}
}

// PhotoResponse represents a successfully saved photo.
type PhotoResponse struct {
	Filename string `json:"filename" example:"photo-1730000000000000000.jpg"`
	Path     string `json:"path" example:"photos/photo-1730000000000000000.jpg"`
	Size     int64  `json:"size" example:"183402"`
}

// maxPhotoBytes is the largest accepted photo upload (15 MB).
const maxPhotoBytes = 15 << 20

// safeExt matches a file extension we are willing to write to disk.
var safeExt = regexp.MustCompile(`^\.[a-z0-9]{2,5}$`)

// extensionByMediaType maps common image media types to a file extension.
var extensionByMediaType = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
	"image/heic": ".heic",
	"image/heif": ".heif",
}

// SavePhoto godoc
// @Summary Upload a photo taken with the camera
// @Description Saves a captured image into the `photos` folder at the project root.
// @Tags photos
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param image formData file true "Captured photo"
// @Success 200 {object} PhotoResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/photos [post]
func (h *PhotoHandler) SavePhoto(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "image is required: " + err.Error()})
		return
	}

	if file.Size > maxPhotoBytes {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "image is too large (max 15 MB)"})
		return
	}

	mediaType, _, err := mime.ParseMediaType(file.Header.Get("Content-Type"))
	if err != nil || !strings.HasPrefix(mediaType, "image/") {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file must be an image"})
		return
	}

	ext := extensionByMediaType[mediaType]
	if ext == "" {
		if candidate := strings.ToLower(filepath.Ext(file.Filename)); safeExt.MatchString(candidate) {
			ext = candidate
		} else {
			ext = ".jpg"
		}
	}

	dir := photosDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create photo folder: " + err.Error()})
		return
	}

	filename := "photo-" + time.Now().UTC().Format("20060102-150405.000000000") + ext
	dst := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save photo: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, PhotoResponse{
		Filename: filename,
		Path:     relativeToProjectRoot(dst),
		Size:     file.Size,
	})
}

// photosDir returns the folder photos are written to. It defaults to
// <project-root>/photos, resolved by walking up from the working directory
// until the directory holding both `backend` and `frontend` is found, so it
// lands in the project root even when the server runs from `backend/`.
// Override with the PHOTO_DIR environment variable.
func photosDir() string {
	if dir := strings.TrimSpace(os.Getenv("PHOTO_DIR")); dir != "" {
		return dir
	}

	wd, err := os.Getwd()
	if err != nil {
		return "photos"
	}

	dir := wd
	for i := 0; i < 8; i++ {
		if isProjectRoot(dir) {
			return filepath.Join(dir, "photos")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Join(wd, "photos")
}

// isProjectRoot reports whether dir looks like the repository root
// (the directory containing both the backend and the frontend).
func isProjectRoot(dir string) bool {
	for _, sub := range []string{"backend", "frontend"} {
		info, err := os.Stat(filepath.Join(dir, sub))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}

// projectRoot returns the repository root, or the working directory as a fallback.
func projectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 8; i++ {
		if isProjectRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return wd
}

// relativeToProjectRoot renders dst relative to the project root when possible.
func relativeToProjectRoot(dst string) string {
	root := projectRoot()
	if root == "" {
		return dst
	}
	rel, err := filepath.Rel(root, dst)
	if err != nil {
		return dst
	}
	return rel
}
