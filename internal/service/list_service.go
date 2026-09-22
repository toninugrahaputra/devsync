package service

import (
	"devsync/internal/model"
	"devsync/internal/repository"
)

type ListService interface {
	Create(userID, boardID uint, req model.ListRequest) (*model.List, error)
	Rename(userID, listID uint, req model.ListRequest) (*model.List, error)
	Move(userID, listID uint, req model.MoveListRequest) (*model.List, error)
	Archive(userID, listID uint) error
}

type listService struct {
	boardRepo repository.BoardRepository
	listRepo  repository.ListRepository
}

func NewListService(boardRepo repository.BoardRepository, listRepo repository.ListRepository) ListService {
	return &listService{boardRepo: boardRepo, listRepo: listRepo}
}

func (s *listService) requireBoardMember(userID, boardID uint) error {
	ok, err := s.boardRepo.IsMember(boardID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *listService) Create(userID, boardID uint, req model.ListRequest) (*model.List, error) {
	if err := s.requireBoardMember(userID, boardID); err != nil {
		return nil, err
	}
	existing, err := s.listRepo.OtherPositions(boardID, 0)
	if err != nil {
		return nil, err
	}
	position := computePosition(existing, len(existing))
	l := &model.List{BoardID: boardID, Name: req.Name, Position: position}
	if err := s.listRepo.Create(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *listService) Rename(userID, listID uint, req model.ListRequest) (*model.List, error) {
	l, err := s.listRepo.GetByID(listID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, l.BoardID); err != nil {
		return nil, err
	}
	l.Name = req.Name
	if err := s.listRepo.Update(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *listService) Move(userID, listID uint, req model.MoveListRequest) (*model.List, error) {
	l, err := s.listRepo.GetByID(listID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, l.BoardID); err != nil {
		return nil, err
	}

	existing, err := s.listRepo.OtherPositions(l.BoardID, l.ID)
	if err != nil {
		return nil, err
	}
	l.Position = computePosition(existing, req.Index)
	if err := s.listRepo.Update(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *listService) Archive(userID, listID uint) error {
	l, err := s.listRepo.GetByID(listID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, l.BoardID); err != nil {
		return err
	}
	return s.listRepo.Archive(listID)
}
