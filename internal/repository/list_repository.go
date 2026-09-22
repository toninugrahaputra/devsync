package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type ListRepository interface {
	Create(l *model.List) error
	GetByID(id uint) (*model.List, error)
	Update(l *model.List) error
	Archive(id uint) error
	ListByBoard(boardID uint) ([]model.List, error)
	OtherPositions(boardID uint, excludeID uint) ([]float64, error)
}

type mysqlListRepository struct {
	db *sql.DB
}

func NewListRepository(db *sql.DB) ListRepository {
	return &mysqlListRepository{db: db}
}

func (r *mysqlListRepository) Create(l *model.List) error {
	res, err := r.db.Exec("INSERT INTO lists (board_id, name, position) VALUES (?, ?, ?)", l.BoardID, l.Name, l.Position)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	l.ID = uint(id)
	return nil
}

func (r *mysqlListRepository) GetByID(id uint) (*model.List, error) {
	var l model.List
	query := "SELECT id, board_id, name, position FROM lists WHERE id = ? AND archived_at IS NULL"
	err := r.db.QueryRow(query, id).Scan(&l.ID, &l.BoardID, &l.Name, &l.Position)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *mysqlListRepository) Update(l *model.List) error {
	_, err := r.db.Exec("UPDATE lists SET name = ?, position = ? WHERE id = ?", l.Name, l.Position, l.ID)
	return err
}

func (r *mysqlListRepository) Archive(id uint) error {
	_, err := r.db.Exec("UPDATE lists SET archived_at = NOW() WHERE id = ?", id)
	return err
}

func (r *mysqlListRepository) ListByBoard(boardID uint) ([]model.List, error) {
	query := "SELECT id, board_id, name, position FROM lists WHERE board_id = ? AND archived_at IS NULL ORDER BY position ASC"
	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lists := []model.List{}
	for rows.Next() {
		var l model.List
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Position); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, nil
}

func (r *mysqlListRepository) OtherPositions(boardID uint, excludeID uint) ([]float64, error) {
	query := "SELECT position FROM lists WHERE board_id = ? AND archived_at IS NULL AND id != ? ORDER BY position ASC"
	rows, err := r.db.Query(query, boardID, excludeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []float64
	for rows.Next() {
		var p float64
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, nil
}
