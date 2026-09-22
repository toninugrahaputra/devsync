package service

import (
	"devsync/internal/model"
	"devsync/internal/repository"
)

type CommentService interface {
	Create(userID, cardID uint, req model.CommentRequest) (*model.Comment, error)
	Delete(userID, commentID uint) error
	ListByCard(userID, cardID uint) ([]model.Comment, error)
}

type commentService struct {
	boardRepo   repository.BoardRepository
	cardRepo    repository.CardRepository
	commentRepo repository.CommentRepository
}

func NewCommentService(boardRepo repository.BoardRepository, cardRepo repository.CardRepository, commentRepo repository.CommentRepository) CommentService {
	return &commentService{boardRepo: boardRepo, cardRepo: cardRepo, commentRepo: commentRepo}
}

func (s *commentService) requireBoardMember(userID, boardID uint) error {
	ok, err := s.boardRepo.IsMember(boardID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *commentService) Create(userID, cardID uint, req model.CommentRequest) (*model.Comment, error) {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}
	comment := &model.Comment{CardID: cardID, UserID: userID, Body: req.Body}
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *commentService) Delete(userID, commentID uint) error {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return err
	}
	c, err := s.cardRepo.GetByID(comment.CardID)
	if err != nil {
		return err
	}
	if comment.UserID != userID {
		// Not your comment: only a board admin can remove it, not just any member.
		role, err := s.boardRepo.GetMemberRole(c.BoardID, userID)
		if err != nil || role != "admin" {
			return ErrForbidden
		}
	}
	return s.commentRepo.Delete(commentID)
}

func (s *commentService) ListByCard(userID, cardID uint) ([]model.Comment, error) {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}
	return s.commentRepo.ListByCard(cardID)
}
