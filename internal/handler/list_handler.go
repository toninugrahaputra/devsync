package handler

import (
	"devsync/internal/model"
	"devsync/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ListHandler struct {
	listService service.ListService
}

func NewListHandler(listService service.ListService) *ListHandler {
	return &ListHandler{listService: listService}
}

func (h *ListHandler) Create(c *gin.Context) {
	boardID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l, err := h.listService.Create(currentUserID(c), boardID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, l)
}

func (h *ListHandler) Rename(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.ListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l, err := h.listService.Rename(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *ListHandler) Move(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.MoveListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l, err := h.listService.Move(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *ListHandler) Archive(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.listService.Archive(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "list archived"})
}
