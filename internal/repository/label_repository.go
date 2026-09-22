package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type LabelRepository interface {
	Create(l *model.Label) error
	Update(l *model.Label) error
	Delete(id uint) error
	GetByID(id uint) (*model.Label, error)
	ListByBoard(boardID uint) ([]model.Label, error)
}

type mysqlLabelRepository struct {
	db *sql.DB
}

func NewLabelRepository(db *sql.DB) LabelRepository {
	return &mysqlLabelRepository{db: db}
}

func (r *mysqlLabelRepository) Create(l *model.Label) error {
	res, err := r.db.Exec("INSERT INTO labels (board_id, name, color) VALUES (?, ?, ?)", l.BoardID, l.Name, l.Color)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	l.ID = uint(id)
	return nil
}

func (r *mysqlLabelRepository) Update(l *model.Label) error {
	_, err := r.db.Exec("UPDATE labels SET name = ?, color = ? WHERE id = ?", l.Name, l.Color, l.ID)
	return err
}

func (r *mysqlLabelRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM labels WHERE id = ?", id)
	return err
}

func (r *mysqlLabelRepository) GetByID(id uint) (*model.Label, error) {
	var l model.Label
	err := r.db.QueryRow("SELECT id, board_id, name, color FROM labels WHERE id = ?", id).Scan(&l.ID, &l.BoardID, &l.Name, &l.Color)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *mysqlLabelRepository) ListByBoard(boardID uint) ([]model.Label, error) {
	rows, err := r.db.Query("SELECT id, board_id, name, color FROM labels WHERE board_id = ? ORDER BY id ASC", boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	labels := []model.Label{}
	for rows.Next() {
		var l model.Label
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, nil
}
