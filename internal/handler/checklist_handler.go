package handler

import (
	"devsync/internal/model"
	"devsync/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChecklistHandler struct {
	checklistService service.ChecklistService
}

func NewChecklistHandler(checklistService service.ChecklistService) *ChecklistHandler {
	return &ChecklistHandler{checklistService: checklistService}
}

func (h *ChecklistHandler) Create(c *gin.Context) {
	cardID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.ChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl, err := h.checklistService.Create(currentUserID(c), cardID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, cl)
}

func (h *ChecklistHandler) Update(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.ChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl, err := h.checklistService.Update(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, cl)
}

func (h *ChecklistHandler) Delete(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.checklistService.Delete(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "checklist deleted"})
}

func (h *ChecklistHandler) CreateItem(c *gin.Context) {
	checklistID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.ChecklistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.checklistService.CreateItem(currentUserID(c), checklistID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ChecklistHandler) UpdateItem(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.UpdateChecklistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.checklistService.UpdateItem(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ChecklistHandler) DeleteItem(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.checklistService.DeleteItem(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "item deleted"})
}
