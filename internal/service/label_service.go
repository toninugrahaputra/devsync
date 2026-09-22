package service

import (
	"devsync/internal/model"
	"devsync/internal/repository"
)

type LabelService interface {
	Create(userID, boardID uint, req model.LabelRequest) (*model.Label, error)
	Update(userID, labelID uint, req model.LabelRequest) (*model.Label, error)
	Delete(userID, labelID uint) error
	ListByBoard(userID, boardID uint) ([]model.Label, error)
}

type labelService struct {
	boardRepo repository.BoardRepository
	labelRepo repository.LabelRepository
}

func NewLabelService(boardRepo repository.BoardRepository, labelRepo repository.LabelRepository) LabelService {
	return &labelService{boardRepo: boardRepo, labelRepo: labelRepo}
}

func (s *labelService) requireBoardMember(userID, boardID uint) error {
	ok, err := s.boardRepo.IsMember(boardID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *labelService) Create(userID, boardID uint, req model.LabelRequest) (*model.Label, error) {
	if err := s.requireBoardMember(userID, boardID); err != nil {
		return nil, err
	}
	l := &model.Label{BoardID: boardID, Name: req.Name, Color: req.Color}
	if err := s.labelRepo.Create(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *labelService) Update(userID, labelID uint, req model.LabelRequest) (*model.Label, error) {
	l, err := s.labelRepo.GetByID(labelID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, l.BoardID); err != nil {
		return nil, err
	}
	l.Name = req.Name
	l.Color = req.Color
	if err := s.labelRepo.Update(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *labelService) Delete(userID, labelID uint) error {
	l, err := s.labelRepo.GetByID(labelID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, l.BoardID); err != nil {
		return err
	}
	return s.labelRepo.Delete(labelID)
}

func (s *labelService) ListByBoard(userID, boardID uint) ([]model.Label, error) {
	if err := s.requireBoardMember(userID, boardID); err != nil {
		return nil, err
	}
	return s.labelRepo.ListByBoard(boardID)
}
