package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type BoardRepository interface {
	Create(b *model.Board) error
	GetByID(id uint) (*model.Board, error)
	Update(b *model.Board) error
	Delete(id uint) error
	ListByWorkspace(workspaceID uint) ([]model.Board, error)
	AddMember(boardID, userID uint, role string) error
	RemoveMember(boardID, userID uint) error
	IsMember(boardID, userID uint) (bool, error)
	GetMemberRole(boardID, userID uint) (string, error)
	ListMembers(boardID uint) ([]model.BoardMember, error)
}

type mysqlBoardRepository struct {
	db *sql.DB
}

func NewBoardRepository(db *sql.DB) BoardRepository {
	return &mysqlBoardRepository{db: db}
}

func (r *mysqlBoardRepository) Create(b *model.Board) error {
	res, err := r.db.Exec(
		"INSERT INTO boards (workspace_id, name, background_color, created_by) VALUES (?, ?, ?, ?)",
		b.WorkspaceID, b.Name, b.BackgroundColor, b.CreatedBy,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	b.ID = uint(id)
	return nil
}

func (r *mysqlBoardRepository) GetByID(id uint) (*model.Board, error) {
	var b model.Board
	query := "SELECT id, workspace_id, name, background_color, created_by, created_at, archived_at FROM boards WHERE id = ?"
	err := r.db.QueryRow(query, id).Scan(&b.ID, &b.WorkspaceID, &b.Name, &b.BackgroundColor, &b.CreatedBy, &b.CreatedAt, &b.ArchivedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *mysqlBoardRepository) Update(b *model.Board) error {
	_, err := r.db.Exec("UPDATE boards SET name = ?, background_color = ? WHERE id = ?", b.Name, b.BackgroundColor, b.ID)
	return err
}

func (r *mysqlBoardRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM boards WHERE id = ?", id)
	return err
}

func (r *mysqlBoardRepository) ListByWorkspace(workspaceID uint) ([]model.Board, error) {
	query := "SELECT id, workspace_id, name, background_color, created_by, created_at, archived_at FROM boards WHERE workspace_id = ? AND archived_at IS NULL ORDER BY created_at ASC"
	rows, err := r.db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	boards := []model.Board{}
	for rows.Next() {
		var b model.Board
		if err := rows.Scan(&b.ID, &b.WorkspaceID, &b.Name, &b.BackgroundColor, &b.CreatedBy, &b.CreatedAt, &b.ArchivedAt); err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, nil
}

func (r *mysqlBoardRepository) AddMember(boardID, userID uint, role string) error {
	_, err := r.db.Exec("INSERT IGNORE INTO board_members (board_id, user_id, role) VALUES (?, ?, ?)", boardID, userID, role)
	return err
}

func (r *mysqlBoardRepository) RemoveMember(boardID, userID uint) error {
	_, err := r.db.Exec("DELETE FROM board_members WHERE board_id = ? AND user_id = ?", boardID, userID)
	return err
}

func (r *mysqlBoardRepository) IsMember(boardID, userID uint) (bool, error) {
	var exists int
	err := r.db.QueryRow("SELECT 1 FROM board_members WHERE board_id = ? AND user_id = ?", boardID, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mysqlBoardRepository) GetMemberRole(boardID, userID uint) (string, error) {
	var role string
	err := r.db.QueryRow("SELECT role FROM board_members WHERE board_id = ? AND user_id = ?", boardID, userID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (r *mysqlBoardRepository) ListMembers(boardID uint) ([]model.BoardMember, error) {
	query := `SELECT bm.board_id, bm.user_id, u.name, u.email, u.avatar_path, bm.role
		FROM board_members bm JOIN users u ON u.id = bm.user_id
		WHERE bm.board_id = ? ORDER BY bm.joined_at ASC`
	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []model.BoardMember{}
	for rows.Next() {
		var m model.BoardMember
		if err := rows.Scan(&m.BoardID, &m.UserID, &m.Name, &m.Email, &m.AvatarPath, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}
