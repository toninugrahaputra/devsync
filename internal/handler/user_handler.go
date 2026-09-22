package handler

import (
	"database/sql"
	"devsync/internal/model"
	"devsync/internal/repository"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var usernameRE = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

const maxAvatarSize = 5 << 20 // 5MB

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) Search(c *gin.Context) {
	users, err := h.userRepo.Search(c.Query("q"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := currentUserID(c)

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	if req.Username != nil {
		trimmed := strings.TrimSpace(*req.Username)
		if trimmed == "" {
			user.Username = nil
		} else {
			if !usernameRE.MatchString(trimmed) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "username must be 3-30 letters, numbers, or underscores"})
				return
			}
			if existing, err := h.userRepo.GetByUsername(trimmed); err == nil && existing.ID != userID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "that username is already taken"})
				return
			} else if err != nil && err != sql.ErrNoRows {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			user.Username = &trimmed
		}
	}

	user.Name = req.Name
	if err := h.userRepo.UpdateProfile(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := currentUserID(c)

	previous, err := h.userRepo.GetByID(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if file.Size > maxAvatarSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is too large (max 5MB)"})
		return
	}

	dir := "./uploads/avatars"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not prepare upload directory"})
		return
	}

	safeName := strings.ReplaceAll(filepath.Base(file.Filename), " ", "_")
	storedName := fmt.Sprintf("%d_%d_%s", userID, time.Now().UnixNano(), safeName)
	storedPath := filepath.Join(dir, storedName)
	if err := c.SaveUploadedFile(file, storedPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}

	if err := h.userRepo.UpdateAvatar(userID, storedPath); err != nil {
		_ = os.Remove(storedPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if previous.AvatarPath != nil {
		_ = os.Remove(*previous.AvatarPath)
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}
