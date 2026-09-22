package handler

import (
	"devsync/internal/model"
	"devsync/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InviteHandler struct {
	inviteService service.InviteService
}

func NewInviteHandler(inviteService service.InviteService) *InviteHandler {
	return &InviteHandler{inviteService: inviteService}
}

func (h *InviteHandler) Create(c *gin.Context) {
	workspaceID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.inviteService.Create(currentUserID(c), workspaceID, req.Email)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *InviteHandler) Preview(c *gin.Context) {
	token := c.Param("token")
	preview, err := h.inviteService.Preview(token)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *InviteHandler) Accept(c *gin.Context) {
	token := c.Param("token")
	if err := h.inviteService.Accept(currentUserID(c), token); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "invite accepted"})
}
