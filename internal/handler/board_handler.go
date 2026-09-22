package handler

import (
	"devsync/internal/model"
	"devsync/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BoardHandler struct {
	boardService service.BoardService
}

func NewBoardHandler(boardService service.BoardService) *BoardHandler {
	return &BoardHandler{boardService: boardService}
}

func (h *BoardHandler) Create(c *gin.Context) {
	workspaceID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.BoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.boardService.Create(currentUserID(c), workspaceID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *BoardHandler) ListByWorkspace(c *gin.Context) {
	workspaceID, ok := paramUint(c, "id")
	if !ok {
		return
	}
	boards, err := h.boardService.ListByWorkspace(currentUserID(c), workspaceID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, boards)
}

func (h *BoardHandler) GetDetail(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	detail, err := h.boardService.GetDetail(currentUserID(c), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *BoardHandler) Update(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.BoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.boardService.Update(currentUserID(c), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *BoardHandler) Delete(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	if err := h.boardService.Delete(currentUserID(c), id); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "board deleted"})
}

func (h *BoardHandler) AddMember(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	var req model.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.boardService.AddMember(currentUserID(c), id, req.UserID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

func (h *BoardHandler) RemoveMember(c *gin.Context) {
	id, ok := paramUint(c, "id")
	if !ok {
		return
	}
	userID, ok := paramUint(c, "userId")
	if !ok {
		return
	}
	if err := h.boardService.RemoveMember(currentUserID(c), id, userID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}
