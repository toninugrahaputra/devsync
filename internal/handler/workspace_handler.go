package handler

import (
	"devsync/internal/model"
	"devsync/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WorkspaceHandler struct {
	workspaceService service.WorkspaceService
}

func NewWorkspaceHandler(workspaceService service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceService: workspaceService}
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	var req model.WorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.workspaceService.Create(currentUserID(c), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *WorkspaceHandler) ListMine(c *gin.Context) {
	workspaces, err := h.workspaceService.ListMine(currentUserID(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, workspaces)
}

func (h *WorkspaceHandler) Get(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	w, err := h.workspaceService.Get(currentUserID(c), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *WorkspaceHandler) Update(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.WorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.workspaceService.Update(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *WorkspaceHandler) Delete(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.workspaceService.Delete(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "workspace deleted"})
}

func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	members, err := h.workspaceService.ListMembers(currentUserID(c), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *WorkspaceHandler) AddMember(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.workspaceService.AddMember(currentUserID(c), id, req.UserID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	userID, ok := paramUint(c, "userId")
	if !ok {
		return
	}
	if err := h.workspaceService.RemoveMember(currentUserID(c), id, userID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}
