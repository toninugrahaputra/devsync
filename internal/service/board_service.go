package service

import (
	"devsync/internal/model"
	"devsync/internal/repository"
	"math/rand"
)

// boardColorPalette mirrors Trello's classic board background colors — all
// tuned to stay readable with white board-header text, so a freshly created
// board never looks bare.
var boardColorPalette = []string{
	"#0079BF", // blue
	"#D29034", // orange
	"#519839", // green
	"#B04632", // red
	"#89609E", // purple
	"#CD5A91", // pink
	"#4BBF6B", // lime
	"#00AECC", // sky
	"#B9871A", // gold (brand-adjacent)
}

type BoardService interface {
	Create(userID, workspaceID uint, req model.BoardRequest) (*model.Board, error)
	ListByWorkspace(userID, workspaceID uint) ([]model.Board, error)
	GetDetail(userID, boardID uint) (*model.BoardDetail, error)
	Update(userID, boardID uint, req model.BoardRequest) (*model.Board, error)
	Delete(userID, boardID uint) error
	AddMember(userID, boardID, targetUserID uint) error
	RemoveMember(userID, boardID, targetUserID uint) error
}

type boardService struct {
	workspaceRepo repository.WorkspaceRepository
	boardRepo     repository.BoardRepository
	listRepo      repository.ListRepository
	cardRepo      repository.CardRepository
	labelRepo     repository.LabelRepository
	checklistRepo repository.ChecklistRepository
}

func NewBoardService(
	workspaceRepo repository.WorkspaceRepository,
	boardRepo repository.BoardRepository,
	listRepo repository.ListRepository,
	cardRepo repository.CardRepository,
	labelRepo repository.LabelRepository,
	checklistRepo repository.ChecklistRepository,
) BoardService {
	return &boardService{workspaceRepo, boardRepo, listRepo, cardRepo, labelRepo, checklistRepo}
}

func (s *boardService) requireBoardMember(userID, boardID uint) error {
	ok, err := s.boardRepo.IsMember(boardID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

// requireBoardAdmin gates the actions that change the board itself or who's
// on it (rename, recolor, delete, membership) — a regular member could
// otherwise delete the board or remove the admin who added them.
func (s *boardService) requireBoardAdmin(userID, boardID uint) error {
	role, err := s.boardRepo.GetMemberRole(boardID, userID)
	if err != nil {
		return ErrForbidden
	}
	if role != "admin" {
		return ErrForbidden
	}
	return nil
}

func (s *boardService) Create(userID, workspaceID uint, req model.BoardRequest) (*model.Board, error) {
	ok, err := s.workspaceRepo.IsMember(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}

	bg := req.BackgroundColor
	if bg == "" {
		bg = boardColorPalette[rand.Intn(len(boardColorPalette))]
	}
	b := &model.Board{WorkspaceID: workspaceID, Name: req.Name, BackgroundColor: bg, CreatedBy: userID}
	if err := s.boardRepo.Create(b); err != nil {
		return nil, err
	}
	if err := s.boardRepo.AddMember(b.ID, userID, "admin"); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *boardService) ListByWorkspace(userID, workspaceID uint) ([]model.Board, error) {
	ok, err := s.workspaceRepo.IsMember(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return s.boardRepo.ListByWorkspace(workspaceID)
}

func (s *boardService) GetDetail(userID, boardID uint) (*model.BoardDetail, error) {
	if err := s.requireBoardMember(userID, boardID); err != nil {
		return nil, err
	}

	board, err := s.boardRepo.GetByID(boardID)
	if err != nil {
		return nil, err
	}

	lists, err := s.listRepo.ListByBoard(boardID)
	if err != nil {
		return nil, err
	}

	cards, err := s.cardRepo.ListByBoard(boardID)
	if err != nil {
		return nil, err
	}

	memberIDs, err := s.cardRepo.MemberIDsByBoard(boardID)
	if err != nil {
		return nil, err
	}

	labelIDs, err := s.cardRepo.LabelIDsByBoard(boardID)
	if err != nil {
		return nil, err
	}

	checklistCounts, err := s.checklistRepo.ItemCountsByBoard(boardID)
	if err != nil {
		return nil, err
	}

	cardsByList := map[uint][]model.CardSummary{}
	for _, c := range cards {
		counts := checklistCounts[c.ID]
		summary := model.CardSummary{
			Card:             c,
			MemberIDs:        memberIDs[c.ID],
			LabelIDs:         labelIDs[c.ID],
			ChecklistTotal:   counts.Total,
			ChecklistChecked: counts.Checked,
		}
		if summary.MemberIDs == nil {
			summary.MemberIDs = []uint{}
		}
		if summary.LabelIDs == nil {
			summary.LabelIDs = []uint{}
		}
		cardsByList[c.ListID] = append(cardsByList[c.ListID], summary)
	}

	listDetails := make([]model.ListDetail, 0, len(lists))
	for _, l := range lists {
		cardsForList := cardsByList[l.ID]
		if cardsForList == nil {
			cardsForList = []model.CardSummary{}
		}
		listDetails = append(listDetails, model.ListDetail{List: l, Cards: cardsForList})
	}

	labels, err := s.labelRepo.ListByBoard(boardID)
	if err != nil {
		return nil, err
	}

	members, err := s.boardRepo.ListMembers(boardID)
	if err != nil {
		return nil, err
	}

	return &model.BoardDetail{
		Board:   *board,
		Lists:   listDetails,
		Labels:  labels,
		Members: members,
	}, nil
}

func (s *boardService) Update(userID, boardID uint, req model.BoardRequest) (*model.Board, error) {
	if err := s.requireBoardAdmin(userID, boardID); err != nil {
		return nil, err
	}
	board, err := s.boardRepo.GetByID(boardID)
	if err != nil {
		return nil, err
	}
	board.Name = req.Name
	if req.BackgroundColor != "" {
		board.BackgroundColor = req.BackgroundColor
	}
	if err := s.boardRepo.Update(board); err != nil {
		return nil, err
	}
	return board, nil
}

func (s *boardService) Delete(userID, boardID uint) error {
	if err := s.requireBoardAdmin(userID, boardID); err != nil {
		return err
	}
	return s.boardRepo.Delete(boardID)
}

func (s *boardService) AddMember(userID, boardID, targetUserID uint) error {
	if err := s.requireBoardAdmin(userID, boardID); err != nil {
		return err
	}
	board, err := s.boardRepo.GetByID(boardID)
	if err != nil {
		return err
	}
	// board membership is meant to be a subset of workspace membership
	inWorkspace, err := s.workspaceRepo.IsMember(board.WorkspaceID, targetUserID)
	if err != nil {
		return err
	}
	if !inWorkspace {
		return ErrForbidden
	}
	return s.boardRepo.AddMember(boardID, targetUserID, "member")
}

func (s *boardService) RemoveMember(userID, boardID, targetUserID uint) error {
	if err := s.requireBoardAdmin(userID, boardID); err != nil {
		return err
	}
	return s.boardRepo.RemoveMember(boardID, targetUserID)
}
