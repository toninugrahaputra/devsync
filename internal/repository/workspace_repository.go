package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type WorkspaceRepository interface {
	Create(w *model.Workspace) error
	GetByID(id uint) (*model.Workspace, error)
	Update(w *model.Workspace) error
	Delete(id uint) error
	ListForUser(userID uint) ([]model.Workspace, error)
	AddMember(workspaceID, userID uint, role string) error
	RemoveMember(workspaceID, userID uint) error
	IsMember(workspaceID, userID uint) (bool, error)
	GetMemberRole(workspaceID, userID uint) (string, error)
	ListMembers(workspaceID uint) ([]model.WorkspaceMember, error)
}

type mysqlWorkspaceRepository struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) WorkspaceRepository {
	return &mysqlWorkspaceRepository{db: db}
}

func (r *mysqlWorkspaceRepository) Create(w *model.Workspace) error {
	res, err := r.db.Exec("INSERT INTO workspaces (name, created_by) VALUES (?, ?)", w.Name, w.CreatedBy)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	w.ID = uint(id)
	return nil
}

func (r *mysqlWorkspaceRepository) GetByID(id uint) (*model.Workspace, error) {
	var w model.Workspace
	err := r.db.QueryRow("SELECT id, name, created_by, created_at FROM workspaces WHERE id = ?", id).
		Scan(&w.ID, &w.Name, &w.CreatedBy, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *mysqlWorkspaceRepository) Update(w *model.Workspace) error {
	_, err := r.db.Exec("UPDATE workspaces SET name = ? WHERE id = ?", w.Name, w.ID)
	return err
}

func (r *mysqlWorkspaceRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM workspaces WHERE id = ?", id)
	return err
}

func (r *mysqlWorkspaceRepository) ListForUser(userID uint) ([]model.Workspace, error) {
	query := `SELECT w.id, w.name, w.created_by, w.created_at, wm.role
		FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id
		WHERE wm.user_id = ?
		ORDER BY w.created_at ASC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workspaces := []model.Workspace{}
	for rows.Next() {
		var w model.Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.CreatedBy, &w.CreatedAt, &w.Role); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, nil
}

func (r *mysqlWorkspaceRepository) AddMember(workspaceID, userID uint, role string) error {
	_, err := r.db.Exec("INSERT IGNORE INTO workspace_members (workspace_id, user_id, role) VALUES (?, ?, ?)", workspaceID, userID, role)
	return err
}

func (r *mysqlWorkspaceRepository) RemoveMember(workspaceID, userID uint) error {
	// Leaving a workspace must also revoke whatever board-level access they
	// picked up inside it — board_members isn't kept in sync automatically,
	// so without this a removed member keeps opening boards they were added
	// to. Do this first: if the second delete below fails, we fail toward
	// "access revoked, still listed as a workspace member" rather than the
	// other way around.
	if _, err := r.db.Exec(
		`DELETE bm FROM board_members bm
		 JOIN boards b ON b.id = bm.board_id
		 WHERE b.workspace_id = ? AND bm.user_id = ?`,
		workspaceID, userID,
	); err != nil {
		return err
	}
	_, err := r.db.Exec("DELETE FROM workspace_members WHERE workspace_id = ? AND user_id = ?", workspaceID, userID)
	return err
}

func (r *mysqlWorkspaceRepository) IsMember(workspaceID, userID uint) (bool, error) {
	var exists int
	err := r.db.QueryRow("SELECT 1 FROM workspace_members WHERE workspace_id = ? AND user_id = ?", workspaceID, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mysqlWorkspaceRepository) GetMemberRole(workspaceID, userID uint) (string, error) {
	var role string
	err := r.db.QueryRow("SELECT role FROM workspace_members WHERE workspace_id = ? AND user_id = ?", workspaceID, userID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (r *mysqlWorkspaceRepository) ListMembers(workspaceID uint) ([]model.WorkspaceMember, error) {
	query := `SELECT wm.workspace_id, wm.user_id, u.name, u.email, u.avatar_path, wm.role, wm.joined_at
		FROM workspace_members wm JOIN users u ON u.id = wm.user_id
		WHERE wm.workspace_id = ? ORDER BY wm.joined_at ASC`
	rows, err := r.db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []model.WorkspaceMember{}
	for rows.Next() {
		var m model.WorkspaceMember
		if err := rows.Scan(&m.WorkspaceID, &m.UserID, &m.Name, &m.Email, &m.AvatarPath, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}
