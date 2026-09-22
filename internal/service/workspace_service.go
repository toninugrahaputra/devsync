package service

import (
	"devsync/internal/model"
	"devsync/internal/repository"
)

type WorkspaceService interface {
	Create(userID uint, req model.WorkspaceRequest) (*model.Workspace, error)
	ListMine(userID uint) ([]model.Workspace, error)
	Get(userID, workspaceID uint) (*model.Workspace, error)
	Update(userID, workspaceID uint, req model.WorkspaceRequest) (*model.Workspace, error)
	Delete(userID, workspaceID uint) error
	AddMember(userID, workspaceID, targetUserID uint) error
	RemoveMember(userID, workspaceID, targetUserID uint) error
	ListMembers(userID, workspaceID uint) ([]model.WorkspaceMember, error)
}

type workspaceService struct {
	workspaceRepo repository.WorkspaceRepository
}

func NewWorkspaceService(workspaceRepo repository.WorkspaceRepository) WorkspaceService {
	return &workspaceService{workspaceRepo: workspaceRepo}
}

func (s *workspaceService) requireMember(userID, workspaceID uint) error {
	ok, err := s.workspaceRepo.IsMember(workspaceID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

// requireOwner gates renaming/deleting the workspace and its membership list —
// a regular member could otherwise remove the owner or delete the workspace.
func (s *workspaceService) requireOwner(userID, workspaceID uint) error {
	role, err := s.workspaceRepo.GetMemberRole(workspaceID, userID)
	if err != nil {
		return ErrForbidden
	}
	if role != "owner" {
		return ErrForbidden
	}
	return nil
}

func (s *workspaceService) Create(userID uint, req model.WorkspaceRequest) (*model.Workspace, error) {
	w := &model.Workspace{Name: req.Name, CreatedBy: userID}
	if err := s.workspaceRepo.Create(w); err != nil {
		return nil, err
	}
	if err := s.workspaceRepo.AddMember(w.ID, userID, "owner"); err != nil {
		return nil, err
	}
	w.Role = "owner"
	return w, nil
}

func (s *workspaceService) ListMine(userID uint) ([]model.Workspace, error) {
	return s.workspaceRepo.ListForUser(userID)
}

func (s *workspaceService) Get(userID, workspaceID uint) (*model.Workspace, error) {
	if err := s.requireMember(userID, workspaceID); err != nil {
		return nil, err
	}
	w, err := s.workspaceRepo.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}
	if role, err := s.workspaceRepo.GetMemberRole(workspaceID, userID); err == nil {
		w.Role = role
	}
	return w, nil
}

func (s *workspaceService) Update(userID, workspaceID uint, req model.WorkspaceRequest) (*model.Workspace, error) {
	if err := s.requireOwner(userID, workspaceID); err != nil {
		return nil, err
	}
	w, err := s.workspaceRepo.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}
	w.Name = req.Name
	if err := s.workspaceRepo.Update(w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *workspaceService) Delete(userID, workspaceID uint) error {
	if err := s.requireOwner(userID, workspaceID); err != nil {
		return err
	}
	return s.workspaceRepo.Delete(workspaceID)
}

func (s *workspaceService) AddMember(userID, workspaceID, targetUserID uint) error {
	if err := s.requireOwner(userID, workspaceID); err != nil {
		return err
	}
	return s.workspaceRepo.AddMember(workspaceID, targetUserID, "member")
}

func (s *workspaceService) RemoveMember(userID, workspaceID, targetUserID uint) error {
	if err := s.requireOwner(userID, workspaceID); err != nil {
		return err
	}
	return s.workspaceRepo.RemoveMember(workspaceID, targetUserID)
}

func (s *workspaceService) ListMembers(userID, workspaceID uint) ([]model.WorkspaceMember, error) {
	if err := s.requireMember(userID, workspaceID); err != nil {
		return nil, err
	}
	return s.workspaceRepo.ListMembers(workspaceID)
}
