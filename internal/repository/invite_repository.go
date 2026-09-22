package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type InviteRepository interface {
	Create(i *model.Invite) error
	GetByToken(token string) (*model.Invite, error)
	MarkAccepted(id uint) error
}

type mysqlInviteRepository struct {
	db *sql.DB
}

func NewInviteRepository(db *sql.DB) InviteRepository {
	return &mysqlInviteRepository{db: db}
}

func (r *mysqlInviteRepository) Create(i *model.Invite) error {
	res, err := r.db.Exec(
		"INSERT INTO invites (workspace_id, email, role, token, invited_by) VALUES (?, ?, ?, ?, ?)",
		i.WorkspaceID, i.Email, i.Role, i.Token, i.InvitedBy,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	i.ID = uint(id)
	return nil
}

func (r *mysqlInviteRepository) GetByToken(token string) (*model.Invite, error) {
	var i model.Invite
	query := "SELECT id, workspace_id, email, role, token, invited_by, created_at, accepted_at FROM invites WHERE token = ?"
	err := r.db.QueryRow(query, token).Scan(
		&i.ID, &i.WorkspaceID, &i.Email, &i.Role, &i.Token, &i.InvitedBy, &i.CreatedAt, &i.AcceptedAt,
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *mysqlInviteRepository) MarkAccepted(id uint) error {
	_, err := r.db.Exec("UPDATE invites SET accepted_at = NOW() WHERE id = ?", id)
	return err
}
