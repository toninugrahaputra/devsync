package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type ActivityRepository interface {
	Create(a *model.Activity) error
	ListByCard(cardID uint) ([]model.Activity, error)
}

type mysqlActivityRepository struct {
	db *sql.DB
}

func NewActivityRepository(db *sql.DB) ActivityRepository {
	return &mysqlActivityRepository{db: db}
}

func (r *mysqlActivityRepository) Create(a *model.Activity) error {
	res, err := r.db.Exec("INSERT INTO activities (card_id, user_id, message) VALUES (?, ?, ?)", a.CardID, a.UserID, a.Message)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	a.ID = uint(id)
	return nil
}

func (r *mysqlActivityRepository) ListByCard(cardID uint) ([]model.Activity, error) {
	query := `SELECT act.id, act.card_id, act.user_id, u.name, act.message, act.created_at
		FROM activities act JOIN users u ON u.id = act.user_id
		WHERE act.card_id = ? ORDER BY act.created_at DESC`
	rows, err := r.db.Query(query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := []model.Activity{}
	for rows.Next() {
		var a model.Activity
		if err := rows.Scan(&a.ID, &a.CardID, &a.UserID, &a.UserName, &a.Message, &a.CreatedAt); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}
