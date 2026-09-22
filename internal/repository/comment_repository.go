package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type CommentRepository interface {
	Create(c *model.Comment) error
	Delete(id uint) error
	GetByID(id uint) (*model.Comment, error)
	ListByCard(cardID uint) ([]model.Comment, error)
}

type mysqlCommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &mysqlCommentRepository{db: db}
}

func (r *mysqlCommentRepository) Create(c *model.Comment) error {
	res, err := r.db.Exec("INSERT INTO comments (card_id, user_id, body) VALUES (?, ?, ?)", c.CardID, c.UserID, c.Body)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	c.ID = uint(id)
	return nil
}

func (r *mysqlCommentRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM comments WHERE id = ?", id)
	return err
}

func (r *mysqlCommentRepository) GetByID(id uint) (*model.Comment, error) {
	var c model.Comment
	query := "SELECT id, card_id, user_id, body, created_at FROM comments WHERE id = ?"
	err := r.db.QueryRow(query, id).Scan(&c.ID, &c.CardID, &c.UserID, &c.Body, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *mysqlCommentRepository) ListByCard(cardID uint) ([]model.Comment, error) {
	query := `SELECT cm.id, cm.card_id, cm.user_id, u.name, cm.body, cm.created_at
		FROM comments cm JOIN users u ON u.id = cm.user_id
		WHERE cm.card_id = ? ORDER BY cm.created_at ASC`
	rows, err := r.db.Query(query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []model.Comment{}
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.CardID, &c.UserID, &c.UserName, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}
