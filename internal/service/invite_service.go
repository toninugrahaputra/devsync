package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"devsync/internal/model"
	"devsync/internal/repository"
	"devsync/pkg/mailer"
)

var ErrInviteEmailMismatch = errors.New("this invite was sent to a different email address")
var ErrInviteAlreadyAccepted = errors.New("this invite has already been accepted")

type InviteService interface {
	Create(userID, workspaceID uint, email string) (*model.InviteResponse, error)
	Preview(token string) (*model.InvitePreview, error)
	Accept(userID uint, token string) error
}

type inviteService struct {
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
	inviteRepo    repository.InviteRepository
	mailer        mailer.Mailer
	frontendURL   string
}

func NewInviteService(
	workspaceRepo repository.WorkspaceRepository,
	userRepo repository.UserRepository,
	inviteRepo repository.InviteRepository,
	m mailer.Mailer,
) InviteService {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	return &inviteService{workspaceRepo, userRepo, inviteRepo, m, frontendURL}
}

func (s *inviteService) Create(userID, workspaceID uint, email string) (*model.InviteResponse, error) {
	role, err := s.workspaceRepo.GetMemberRole(workspaceID, userID)
	if err != nil {
		return nil, ErrForbidden
	}
	if role != "owner" {
		return nil, ErrForbidden
	}

	email = strings.ToLower(strings.TrimSpace(email))

	// Already has an account: skip the invite dance entirely and add them now.
	if existing, err := s.userRepo.GetByEmail(email); err == nil && existing != nil {
		if err := s.workspaceRepo.AddMember(workspaceID, existing.ID, "member"); err != nil {
			return nil, err
		}
		return &model.InviteResponse{AlreadyMember: true}, nil
	}

	workspace, err := s.workspaceRepo.GetByID(workspaceID)
	if err != nil {
		return nil, err
	}

	token, err := randomToken()
	if err != nil {
		return nil, err
	}

	invite := &model.Invite{
		WorkspaceID: workspaceID,
		Email:       email,
		Role:        "member",
		Token:       token,
		InvitedBy:   userID,
	}
	if err := s.inviteRepo.Create(invite); err != nil {
		return nil, err
	}

	link := fmt.Sprintf("%s/invite/%s", s.frontendURL, token)
	subject := fmt.Sprintf("You've been invited to join %s on DevSync", workspace.Name)
	body := fmt.Sprintf(
		"You've been invited to join the \"%s\" workspace on DevSync.\n\nAccept the invite: %s\n\nIf you weren't expecting this, you can ignore this email.",
		workspace.Name, link,
	)
	_ = s.mailer.Send(email, subject, body)

	return &model.InviteResponse{
		AlreadyMember: false,
		EmailSent:     s.mailer.IsLive(),
		InviteLink:    link,
	}, nil
}

func (s *inviteService) Preview(token string) (*model.InvitePreview, error) {
	invite, err := s.inviteRepo.GetByToken(token)
	if err != nil {
		return nil, err
	}
	workspace, err := s.workspaceRepo.GetByID(invite.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return &model.InvitePreview{
		WorkspaceName:   workspace.Name,
		Email:           invite.Email,
		AlreadyAccepted: invite.AcceptedAt != nil,
	}, nil
}

func (s *inviteService) Accept(userID uint, token string) error {
	invite, err := s.inviteRepo.GetByToken(token)
	if err != nil {
		return err
	}
	if invite.AcceptedAt != nil {
		return ErrInviteAlreadyAccepted
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if !strings.EqualFold(user.Email, invite.Email) {
		return ErrInviteEmailMismatch
	}

	if err := s.workspaceRepo.AddMember(invite.WorkspaceID, userID, invite.Role); err != nil {
		return err
	}
	return s.inviteRepo.MarkAccepted(invite.ID)
}

func randomToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
