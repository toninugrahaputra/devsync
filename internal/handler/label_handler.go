package handler

import (
	"devsync/internal/model"
	"devsync/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LabelHandler struct {
	labelService service.LabelService
}

func NewLabelHandler(labelService service.LabelService) *LabelHandler {
	return &LabelHandler{labelService: labelService}
}

func (h *LabelHandler) ListByBoard(c *gin.Context) {
	boardID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	labels, err := h.labelService.ListByBoard(currentUserID(c), boardID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, labels)
}

func (h *LabelHandler) Create(c *gin.Context) {
	boardID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.LabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	label, err := h.labelService.Create(currentUserID(c), boardID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, label)
}

func (h *LabelHandler) Update(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.LabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	label, err := h.labelService.Update(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, label)
}

func (h *LabelHandler) Delete(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.labelService.Delete(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "label deleted"})
}
