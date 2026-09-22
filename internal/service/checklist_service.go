package service

import (
	"fmt"

	"devsync/internal/model"
	"devsync/internal/repository"
)

type ChecklistService interface {
	Create(userID, cardID uint, req model.ChecklistRequest) (*model.Checklist, error)
	Update(userID, checklistID uint, req model.ChecklistRequest) (*model.Checklist, error)
	Delete(userID, checklistID uint) error

	CreateItem(userID, checklistID uint, req model.ChecklistItemRequest) (*model.ChecklistItem, error)
	UpdateItem(userID, itemID uint, req model.UpdateChecklistItemRequest) (*model.ChecklistItem, error)
	DeleteItem(userID, itemID uint) error
}

type checklistService struct {
	boardRepo     repository.BoardRepository
	cardRepo      repository.CardRepository
	checklistRepo repository.ChecklistRepository
	activityRepo  repository.ActivityRepository
}

func NewChecklistService(
	boardRepo repository.BoardRepository,
	cardRepo repository.CardRepository,
	checklistRepo repository.ChecklistRepository,
	activityRepo repository.ActivityRepository,
) ChecklistService {
	return &checklistService{boardRepo: boardRepo, cardRepo: cardRepo, checklistRepo: checklistRepo, activityRepo: activityRepo}
}

func (s *checklistService) requireBoardMember(userID, boardID uint) error {
	ok, err := s.boardRepo.IsMember(boardID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *checklistService) log(cardID, userID uint, message string) {
	_ = s.activityRepo.Create(&model.Activity{CardID: cardID, UserID: userID, Message: message})
}

func (s *checklistService) Create(userID, cardID uint, req model.ChecklistRequest) (*model.Checklist, error) {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}

	existing, err := s.checklistRepo.ListByCard(cardID)
	if err != nil {
		return nil, err
	}
	positions := make([]float64, len(existing))
	for i, e := range existing {
		positions[i] = e.Position
	}
	position := computePosition(positions, len(positions))

	cl := &model.Checklist{CardID: cardID, Title: req.Title, Position: position}
	if err := s.checklistRepo.Create(cl); err != nil {
		return nil, err
	}
	return cl, nil
}

func (s *checklistService) Update(userID, checklistID uint, req model.ChecklistRequest) (*model.Checklist, error) {
	cl, err := s.checklistRepo.GetByID(checklistID)
	if err != nil {
		return nil, err
	}
	c, err := s.cardRepo.GetByID(cl.CardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}
	cl.Title = req.Title
	if err := s.checklistRepo.Update(cl); err != nil {
		return nil, err
	}
	return cl, nil
}

func (s *checklistService) Delete(userID, checklistID uint) error {
	cl, err := s.checklistRepo.GetByID(checklistID)
	if err != nil {
		return err
	}
	c, err := s.cardRepo.GetByID(cl.CardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	return s.checklistRepo.Delete(checklistID)
}

func (s *checklistService) CreateItem(userID, checklistID uint, req model.ChecklistItemRequest) (*model.ChecklistItem, error) {
	cl, err := s.checklistRepo.GetByID(checklistID)
	if err != nil {
		return nil, err
	}
	c, err := s.cardRepo.GetByID(cl.CardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}

	existing, err := s.checklistRepo.ListItemsByChecklist(checklistID)
	if err != nil {
		return nil, err
	}
	positions := make([]float64, len(existing))
	for i, e := range existing {
		positions[i] = e.Position
	}
	position := computePosition(positions, len(positions))

	item := &model.ChecklistItem{ChecklistID: checklistID, Text: req.Text, Position: position}
	if err := s.checklistRepo.CreateItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *checklistService) UpdateItem(userID, itemID uint, req model.UpdateChecklistItemRequest) (*model.ChecklistItem, error) {
	item, err := s.checklistRepo.GetItemByID(itemID)
	if err != nil {
		return nil, err
	}
	cl, err := s.checklistRepo.GetByID(item.ChecklistID)
	if err != nil {
		return nil, err
	}
	c, err := s.cardRepo.GetByID(cl.CardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}

	if req.Text != nil {
		item.Text = *req.Text
	}
	if req.IsChecked != nil && *req.IsChecked != item.IsChecked {
		item.IsChecked = *req.IsChecked
		if item.IsChecked {
			s.log(c.ID, userID, fmt.Sprintf("completed %s", item.Text))
		} else {
			s.log(c.ID, userID, fmt.Sprintf("marked %s incomplete", item.Text))
		}
	}
	if err := s.checklistRepo.UpdateItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *checklistService) DeleteItem(userID, itemID uint) error {
	item, err := s.checklistRepo.GetItemByID(itemID)
	if err != nil {
		return err
	}
	cl, err := s.checklistRepo.GetByID(item.ChecklistID)
	if err != nil {
		return err
	}
	c, err := s.cardRepo.GetByID(cl.CardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	return s.checklistRepo.DeleteItem(itemID)
}
