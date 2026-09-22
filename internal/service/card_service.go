package service

import (
	"fmt"
	"os"
	"time"

	"devsync/internal/model"
	"devsync/internal/repository"
)

type CardService interface {
	Create(userID, listID uint, req model.CreateCardRequest) (*model.Card, error)
	GetDetail(userID, cardID uint) (*model.CardDetail, error)
	Update(userID, cardID uint, req model.UpdateCardRequest) (*model.Card, error)
	Move(userID, cardID uint, req model.MoveCardRequest) (*model.Card, error)
	Archive(userID, cardID uint) error
	ListArchived(userID, boardID uint) ([]model.Card, error)
	Restore(userID, cardID uint) (*model.Card, error)
	PermanentDelete(userID, cardID uint) error
	AddMember(userID, cardID, targetUserID uint) error
	RemoveMember(userID, cardID, targetUserID uint) error
	AddLabel(userID, cardID, labelID uint) error
	RemoveLabel(userID, cardID, labelID uint) error
	AddAttachment(userID, cardID uint, fileName, filePath string, fileSize int64) (*model.Attachment, error)
	DeleteAttachment(userID, attachmentID uint) error
}

type cardService struct {
	boardRepo      repository.BoardRepository
	listRepo       repository.ListRepository
	cardRepo       repository.CardRepository
	labelRepo      repository.LabelRepository
	checklistRepo  repository.ChecklistRepository
	commentRepo    repository.CommentRepository
	attachmentRepo repository.AttachmentRepository
	activityRepo   repository.ActivityRepository
	userRepo       repository.UserRepository
}

func NewCardService(
	boardRepo repository.BoardRepository,
	listRepo repository.ListRepository,
	cardRepo repository.CardRepository,
	labelRepo repository.LabelRepository,
	checklistRepo repository.ChecklistRepository,
	commentRepo repository.CommentRepository,
	attachmentRepo repository.AttachmentRepository,
	activityRepo repository.ActivityRepository,
	userRepo repository.UserRepository,
) CardService {
	return &cardService{boardRepo, listRepo, cardRepo, labelRepo, checklistRepo, commentRepo, attachmentRepo, activityRepo, userRepo}
}

func (s *cardService) requireBoardMember(userID, boardID uint) error {
	ok, err := s.boardRepo.IsMember(boardID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

// log records a best-effort audit-trail line; a logging failure never fails
// the action it's describing.
func (s *cardService) log(cardID, userID uint, message string) {
	_ = s.activityRepo.Create(&model.Activity{CardID: cardID, UserID: userID, Message: message})
}

func (s *cardService) Create(userID, listID uint, req model.CreateCardRequest) (*model.Card, error) {
	list, err := s.listRepo.GetByID(listID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, list.BoardID); err != nil {
		return nil, err
	}

	existing, err := s.cardRepo.OtherPositions(listID, 0)
	if err != nil {
		return nil, err
	}
	position := computePosition(existing, len(existing))

	c := &model.Card{
		ListID:    listID,
		BoardID:   list.BoardID,
		Title:     req.Title,
		Position:  position,
		CreatedBy: userID,
	}
	if err := s.cardRepo.Create(c); err != nil {
		return nil, err
	}
	s.log(c.ID, userID, fmt.Sprintf("added this card to %s", list.Name))
	return c, nil
}

func (s *cardService) GetDetail(userID, cardID uint) (*model.CardDetail, error) {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}

	memberIDs, err := s.cardRepo.MemberIDs(cardID)
	if err != nil {
		return nil, err
	}
	members := []model.BoardMember{}
	if len(memberIDs) > 0 {
		boardMembers, err := s.boardRepo.ListMembers(c.BoardID)
		if err != nil {
			return nil, err
		}
		wanted := map[uint]bool{}
		for _, id := range memberIDs {
			wanted[id] = true
		}
		for _, m := range boardMembers {
			if wanted[m.UserID] {
				members = append(members, m)
			}
		}
	}

	labelIDs, err := s.cardRepo.LabelIDs(cardID)
	if err != nil {
		return nil, err
	}
	labels := []model.Label{}
	if len(labelIDs) > 0 {
		boardLabels, err := s.labelRepo.ListByBoard(c.BoardID)
		if err != nil {
			return nil, err
		}
		wanted := map[uint]bool{}
		for _, id := range labelIDs {
			wanted[id] = true
		}
		for _, l := range boardLabels {
			if wanted[l.ID] {
				labels = append(labels, l)
			}
		}
	}

	checklists, err := s.checklistRepo.ListByCard(cardID)
	if err != nil {
		return nil, err
	}
	itemsByChecklist, err := s.checklistRepo.ListItemsByCard(cardID)
	if err != nil {
		return nil, err
	}
	checklistDetails := make([]model.ChecklistDetail, 0, len(checklists))
	for _, cl := range checklists {
		items := itemsByChecklist[cl.ID]
		if items == nil {
			items = []model.ChecklistItem{}
		}
		checklistDetails = append(checklistDetails, model.ChecklistDetail{Checklist: cl, Items: items})
	}

	comments, err := s.commentRepo.ListByCard(cardID)
	if err != nil {
		return nil, err
	}

	attachments, err := s.attachmentRepo.ListByCard(cardID)
	if err != nil {
		return nil, err
	}

	activities, err := s.activityRepo.ListByCard(cardID)
	if err != nil {
		return nil, err
	}

	return &model.CardDetail{
		Card:        *c,
		Members:     members,
		Labels:      labels,
		Checklists:  checklistDetails,
		Comments:    comments,
		Attachments: attachments,
		Activities:  activities,
	}, nil
}

func (s *cardService) Update(userID, cardID uint, req model.UpdateCardRequest) (*model.Card, error) {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}

	if req.Title != nil && *req.Title != c.Title {
		c.Title = *req.Title
		s.log(cardID, userID, fmt.Sprintf("renamed this card to “%s”", c.Title))
	}
	if req.Description != nil {
		c.Description = *req.Description
	}
	if req.IsCompleted != nil && *req.IsCompleted != c.IsCompleted {
		c.IsCompleted = *req.IsCompleted
		if c.IsCompleted {
			s.log(cardID, userID, "marked this card complete")
		} else {
			s.log(cardID, userID, "marked this card incomplete")
		}
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			if c.DueDate != nil {
				s.log(cardID, userID, "removed the due date")
			}
			c.DueDate = nil
		} else {
			t, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				return nil, err
			}
			c.DueDate = &t
			s.log(cardID, userID, fmt.Sprintf("set this card to be due %s", t.Format("Jan 2, 2006")))
		}
	}
	if req.StartDate != nil {
		if *req.StartDate == "" {
			c.StartDate = nil
		} else {
			t, err := time.Parse("2006-01-02", *req.StartDate)
			if err != nil {
				return nil, err
			}
			c.StartDate = &t
		}
	}
	if req.CoverColor != nil {
		if *req.CoverColor == "" {
			c.CoverColor = nil
		} else {
			c.CoverColor = req.CoverColor
		}
	}

	if err := s.cardRepo.Update(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *cardService) Move(userID, cardID uint, req model.MoveCardRequest) (*model.Card, error) {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}

	targetList, err := s.listRepo.GetByID(req.ListID)
	if err != nil {
		return nil, err
	}
	if targetList.BoardID != c.BoardID {
		return nil, ErrForbidden
	}

	existing, err := s.cardRepo.OtherPositions(req.ListID, cardID)
	if err != nil {
		return nil, err
	}
	position := computePosition(existing, req.Index)

	if req.ListID != c.ListID {
		if sourceList, err := s.listRepo.GetByID(c.ListID); err == nil {
			s.log(cardID, userID, fmt.Sprintf("moved this card from %s to %s", sourceList.Name, targetList.Name))
		}
	}

	if err := s.cardRepo.Move(cardID, req.ListID, position); err != nil {
		return nil, err
	}
	c.ListID = req.ListID
	c.Position = position
	return c, nil
}

func (s *cardService) Archive(userID, cardID uint) error {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	if err := s.cardRepo.Archive(cardID); err != nil {
		return err
	}
	s.log(cardID, userID, "archived this card")
	return nil
}

func (s *cardService) ListArchived(userID, boardID uint) ([]model.Card, error) {
	if err := s.requireBoardMember(userID, boardID); err != nil {
		return nil, err
	}
	return s.cardRepo.ListArchivedByBoard(boardID)
}

func (s *cardService) Restore(userID, cardID uint) (*model.Card, error) {
	c, err := s.cardRepo.GetByIDAny(cardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}

	existing, err := s.cardRepo.OtherPositions(c.ListID, cardID)
	if err != nil {
		return nil, err
	}
	position := computePosition(existing, len(existing))
	if err := s.cardRepo.Move(cardID, c.ListID, position); err != nil {
		return nil, err
	}
	if err := s.cardRepo.Restore(cardID); err != nil {
		return nil, err
	}
	c.Position = position
	s.log(cardID, userID, "sent this card back to the board")
	return c, nil
}

func (s *cardService) PermanentDelete(userID, cardID uint) error {
	c, err := s.cardRepo.GetByIDAny(cardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	return s.cardRepo.PermanentDelete(cardID)
}

func (s *cardService) AddMember(userID, cardID, targetUserID uint) error {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	ok, err := s.boardRepo.IsMember(c.BoardID, targetUserID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	if err := s.cardRepo.AddMember(cardID, targetUserID); err != nil {
		return err
	}
	if target, err := s.userRepo.GetByID(targetUserID); err == nil {
		s.log(cardID, userID, fmt.Sprintf("added %s to this card", target.Name))
	}
	return nil
}

func (s *cardService) RemoveMember(userID, cardID, targetUserID uint) error {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	if err := s.cardRepo.RemoveMember(cardID, targetUserID); err != nil {
		return err
	}
	if target, err := s.userRepo.GetByID(targetUserID); err == nil {
		s.log(cardID, userID, fmt.Sprintf("removed %s from this card", target.Name))
	}
	return nil
}

func (s *cardService) AddLabel(userID, cardID, labelID uint) error {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	if err := s.cardRepo.AddLabel(cardID, labelID); err != nil {
		return err
	}
	if label, err := s.labelRepo.GetByID(labelID); err == nil {
		name := label.Name
		if name == "" {
			name = label.Color
		}
		s.log(cardID, userID, fmt.Sprintf("added the “%s” label", name))
	}
	return nil
}

func (s *cardService) RemoveLabel(userID, cardID, labelID uint) error {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	return s.cardRepo.RemoveLabel(cardID, labelID)
}

func (s *cardService) AddAttachment(userID, cardID uint, fileName, filePath string, fileSize int64) (*model.Attachment, error) {
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return nil, err
	}
	a := &model.Attachment{CardID: cardID, FileName: fileName, FilePath: filePath, FileSize: fileSize, UploadedBy: userID}
	if err := s.attachmentRepo.Create(a); err != nil {
		return nil, err
	}
	s.log(cardID, userID, fmt.Sprintf("attached %s", fileName))
	return a, nil
}

func (s *cardService) DeleteAttachment(userID, attachmentID uint) error {
	a, err := s.attachmentRepo.GetByID(attachmentID)
	if err != nil {
		return err
	}
	c, err := s.cardRepo.GetByID(a.CardID)
	if err != nil {
		return err
	}
	if err := s.requireBoardMember(userID, c.BoardID); err != nil {
		return err
	}
	if err := s.attachmentRepo.Delete(attachmentID); err != nil {
		return err
	}
	_ = os.Remove(a.FilePath)
	return nil
}
