package handler

import (
	"devsync/internal/model"
	"devsync/internal/service"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxAttachmentSize = 15 << 20 // 15MB

type CardHandler struct {
	cardService service.CardService
}

func NewCardHandler(cardService service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

func (h *CardHandler) Create(c *gin.Context) {
	listID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	card, err := h.cardService.Create(currentUserID(c), listID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, card)
}

func (h *CardHandler) GetDetail(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	detail, err := h.cardService.GetDetail(currentUserID(c), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *CardHandler) Update(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	card, err := h.cardService.Update(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, card)
}

func (h *CardHandler) Move(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.MoveCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	card, err := h.cardService.Move(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, card)
}

func (h *CardHandler) Archive(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.cardService.Archive(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "card archived"})
}

func (h *CardHandler) ListArchived(c *gin.Context) {
	boardID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	cards, err := h.cardService.ListArchived(currentUserID(c), boardID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, cards)
}

func (h *CardHandler) Restore(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	card, err := h.cardService.Restore(currentUserID(c), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, card)
}

func (h *CardHandler) PermanentDelete(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.cardService.PermanentDelete(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "card deleted"})
}

func (h *CardHandler) AddMember(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	userID, ok := paramUint(c, "userId")
	if !ok {
		return
	}
	if err := h.cardService.AddMember(currentUserID(c), id, userID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

func (h *CardHandler) RemoveMember(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	userID, ok := paramUint(c, "userId")
	if !ok {
		return
	}
	if err := h.cardService.RemoveMember(currentUserID(c), id, userID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

func (h *CardHandler) AddLabel(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	labelID, ok := paramUint(c, "labelId")
	if !ok {
		return
	}
	if err := h.cardService.AddLabel(currentUserID(c), id, labelID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "label added"})
}

func (h *CardHandler) RemoveLabel(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	labelID, ok := paramUint(c, "labelId")
	if !ok {
		return
	}
	if err := h.cardService.RemoveLabel(currentUserID(c), id, labelID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "label removed"})
}

func (h *CardHandler) AddAttachment(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if file.Size > maxAttachmentSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is too large (max 15MB)"})
		return
	}

	dir := "./uploads/attachments"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not prepare upload directory"})
		return
	}

	safeName := strings.ReplaceAll(filepath.Base(file.Filename), " ", "_")
	storedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	storedPath := filepath.Join(dir, storedName)
	if err := c.SaveUploadedFile(file, storedPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}

	attachment, err := h.cardService.AddAttachment(currentUserID(c), id, file.Filename, storedPath, file.Size)
	if err != nil {
		_ = os.Remove(storedPath)
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, attachment)
}

func (h *CardHandler) DeleteAttachment(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.cardService.DeleteAttachment(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "attachment deleted"})
}
