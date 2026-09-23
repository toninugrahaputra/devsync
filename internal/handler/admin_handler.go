package handler

import (
	"devsync/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userRepo repository.UserRepository
}

func NewAdminHandler(userRepo repository.UserRepository) *AdminHandler {
	return &AdminHandler{userRepo: userRepo}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.ListForAdmin(strings.TrimSpace(c.Query("q")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}
