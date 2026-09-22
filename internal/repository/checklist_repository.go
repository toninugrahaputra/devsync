package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type ChecklistRepository interface {
	Create(cl *model.Checklist) error
	Update(cl *model.Checklist) error
	Delete(id uint) error
	GetByID(id uint) (*model.Checklist, error)
	ListByCard(cardID uint) ([]model.Checklist, error)

	CreateItem(item *model.ChecklistItem) error
	UpdateItem(item *model.ChecklistItem) error
	DeleteItem(id uint) error
	GetItemByID(id uint) (*model.ChecklistItem, error)
	ListItemsByChecklist(checklistID uint) ([]model.ChecklistItem, error)
	ListItemsByCard(cardID uint) (map[uint][]model.ChecklistItem, error)
	ItemCountsByBoard(boardID uint) (map[uint]model.ChecklistCounts, error)
}

type mysqlChecklistRepository struct {
	db *sql.DB
}

func NewChecklistRepository(db *sql.DB) ChecklistRepository {
	return &mysqlChecklistRepository{db: db}
}

func (r *mysqlChecklistRepository) Create(cl *model.Checklist) error {
	res, err := r.db.Exec("INSERT INTO checklists (card_id, title, position) VALUES (?, ?, ?)", cl.CardID, cl.Title, cl.Position)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	cl.ID = uint(id)
	return nil
}

func (r *mysqlChecklistRepository) Update(cl *model.Checklist) error {
	_, err := r.db.Exec("UPDATE checklists SET title = ? WHERE id = ?", cl.Title, cl.ID)
	return err
}

func (r *mysqlChecklistRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM checklists WHERE id = ?", id)
	return err
}

func (r *mysqlChecklistRepository) GetByID(id uint) (*model.Checklist, error) {
	var cl model.Checklist
	err := r.db.QueryRow("SELECT id, card_id, title, position FROM checklists WHERE id = ?", id).Scan(&cl.ID, &cl.CardID, &cl.Title, &cl.Position)
	if err != nil {
		return nil, err
	}
	return &cl, nil
}

func (r *mysqlChecklistRepository) ListByCard(cardID uint) ([]model.Checklist, error) {
	rows, err := r.db.Query("SELECT id, card_id, title, position FROM checklists WHERE card_id = ? ORDER BY position ASC", cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checklists := []model.Checklist{}
	for rows.Next() {
		var cl model.Checklist
		if err := rows.Scan(&cl.ID, &cl.CardID, &cl.Title, &cl.Position); err != nil {
			return nil, err
		}
		checklists = append(checklists, cl)
	}
	return checklists, nil
}

func (r *mysqlChecklistRepository) CreateItem(item *model.ChecklistItem) error {
	res, err := r.db.Exec("INSERT INTO checklist_items (checklist_id, text, is_checked, position) VALUES (?, ?, ?, ?)", item.ChecklistID, item.Text, item.IsChecked, item.Position)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	item.ID = uint(id)
	return nil
}

func (r *mysqlChecklistRepository) UpdateItem(item *model.ChecklistItem) error {
	_, err := r.db.Exec("UPDATE checklist_items SET text = ?, is_checked = ? WHERE id = ?", item.Text, item.IsChecked, item.ID)
	return err
}

func (r *mysqlChecklistRepository) DeleteItem(id uint) error {
	_, err := r.db.Exec("DELETE FROM checklist_items WHERE id = ?", id)
	return err
}

func (r *mysqlChecklistRepository) GetItemByID(id uint) (*model.ChecklistItem, error) {
	var item model.ChecklistItem
	err := r.db.QueryRow("SELECT id, checklist_id, text, is_checked, position FROM checklist_items WHERE id = ?", id).
		Scan(&item.ID, &item.ChecklistID, &item.Text, &item.IsChecked, &item.Position)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *mysqlChecklistRepository) ListItemsByChecklist(checklistID uint) ([]model.ChecklistItem, error) {
	rows, err := r.db.Query("SELECT id, checklist_id, text, is_checked, position FROM checklist_items WHERE checklist_id = ? ORDER BY position ASC", checklistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.ChecklistItem{}
	for rows.Next() {
		var item model.ChecklistItem
		if err := rows.Scan(&item.ID, &item.ChecklistID, &item.Text, &item.IsChecked, &item.Position); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *mysqlChecklistRepository) ListItemsByCard(cardID uint) (map[uint][]model.ChecklistItem, error) {
	query := `SELECT ci.id, ci.checklist_id, ci.text, ci.is_checked, ci.position
		FROM checklist_items ci JOIN checklists cl ON cl.id = ci.checklist_id
		WHERE cl.card_id = ? ORDER BY ci.position ASC`
	rows, err := r.db.Query(query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[uint][]model.ChecklistItem{}
	for rows.Next() {
		var item model.ChecklistItem
		if err := rows.Scan(&item.ID, &item.ChecklistID, &item.Text, &item.IsChecked, &item.Position); err != nil {
			return nil, err
		}
		result[item.ChecklistID] = append(result[item.ChecklistID], item)
	}
	return result, nil
}

// ItemCountsByBoard returns, per card, how many checklist items it has and
// how many are checked — the "3/5" badge shown on the card face is derived
// from this without fetching each card's full checklist detail.
func (r *mysqlChecklistRepository) ItemCountsByBoard(boardID uint) (map[uint]model.ChecklistCounts, error) {
	query := `SELECT cl.card_id, COUNT(ci.id) AS total, COALESCE(SUM(ci.is_checked), 0) AS checked
		FROM checklist_items ci
		JOIN checklists cl ON cl.id = ci.checklist_id
		JOIN cards c ON c.id = cl.card_id
		WHERE c.board_id = ?
		GROUP BY cl.card_id`
	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[uint]model.ChecklistCounts{}
	for rows.Next() {
		var cardID uint
		var counts model.ChecklistCounts
		if err := rows.Scan(&cardID, &counts.Total, &counts.Checked); err != nil {
			return nil, err
		}
		result[cardID] = counts
	}
	return result, nil
}
