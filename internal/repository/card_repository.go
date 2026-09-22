package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type CardRepository interface {
	Create(c *model.Card) error
	GetByID(id uint) (*model.Card, error)
	GetByIDAny(id uint) (*model.Card, error)
	Update(c *model.Card) error
	Move(cardID, listID uint, position float64) error
	Archive(id uint) error
	Restore(id uint) error
	PermanentDelete(id uint) error
	ListByBoard(boardID uint) ([]model.Card, error)
	ListArchivedByBoard(boardID uint) ([]model.Card, error)
	OtherPositions(listID uint, excludeID uint) ([]float64, error)
	AddMember(cardID, userID uint) error
	RemoveMember(cardID, userID uint) error
	MemberIDs(cardID uint) ([]uint, error)
	MemberIDsByBoard(boardID uint) (map[uint][]uint, error)
	AddLabel(cardID, labelID uint) error
	RemoveLabel(cardID, labelID uint) error
	LabelIDs(cardID uint) ([]uint, error)
	LabelIDsByBoard(boardID uint) (map[uint][]uint, error)
}

type mysqlCardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) CardRepository {
	return &mysqlCardRepository{db: db}
}

const cardColumns = "id, list_id, board_id, title, COALESCE(description, ''), position, due_date, start_date, is_completed, cover_color, created_by, created_at, updated_at"

func scanCard(row interface {
	Scan(dest ...interface{}) error
}, c *model.Card) error {
	return row.Scan(
		&c.ID, &c.ListID, &c.BoardID, &c.Title, &c.Description, &c.Position,
		&c.DueDate, &c.StartDate, &c.IsCompleted, &c.CoverColor, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
}

func (r *mysqlCardRepository) Create(c *model.Card) error {
	res, err := r.db.Exec(
		"INSERT INTO cards (list_id, board_id, title, description, position, created_by) VALUES (?, ?, ?, ?, ?, ?)",
		c.ListID, c.BoardID, c.Title, c.Description, c.Position, c.CreatedBy,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	c.ID = uint(id)
	return nil
}

func (r *mysqlCardRepository) GetByID(id uint) (*model.Card, error) {
	var c model.Card
	query := "SELECT " + cardColumns + " FROM cards WHERE id = ? AND archived_at IS NULL"
	if err := scanCard(r.db.QueryRow(query, id), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetByIDAny fetches a card regardless of archived state — used by the
// restore/permanent-delete paths, which need to reach archived cards that
// GetByID deliberately hides from the rest of the app.
func (r *mysqlCardRepository) GetByIDAny(id uint) (*model.Card, error) {
	var c model.Card
	query := "SELECT " + cardColumns + " FROM cards WHERE id = ?"
	if err := scanCard(r.db.QueryRow(query, id), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *mysqlCardRepository) Update(c *model.Card) error {
	_, err := r.db.Exec(
		"UPDATE cards SET title = ?, description = ?, due_date = ?, start_date = ?, is_completed = ?, cover_color = ? WHERE id = ?",
		c.Title, c.Description, c.DueDate, c.StartDate, c.IsCompleted, c.CoverColor, c.ID,
	)
	return err
}

func (r *mysqlCardRepository) Move(cardID, listID uint, position float64) error {
	_, err := r.db.Exec("UPDATE cards SET list_id = ?, position = ? WHERE id = ?", listID, position, cardID)
	return err
}

func (r *mysqlCardRepository) Archive(id uint) error {
	_, err := r.db.Exec("UPDATE cards SET archived_at = NOW() WHERE id = ?", id)
	return err
}

func (r *mysqlCardRepository) Restore(id uint) error {
	_, err := r.db.Exec("UPDATE cards SET archived_at = NULL WHERE id = ?", id)
	return err
}

func (r *mysqlCardRepository) PermanentDelete(id uint) error {
	_, err := r.db.Exec("DELETE FROM cards WHERE id = ?", id)
	return err
}

func (r *mysqlCardRepository) ListByBoard(boardID uint) ([]model.Card, error) {
	query := "SELECT " + cardColumns + " FROM cards WHERE board_id = ? AND archived_at IS NULL ORDER BY position ASC"
	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := []model.Card{}
	for rows.Next() {
		var c model.Card
		if err := scanCard(rows, &c); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *mysqlCardRepository) ListArchivedByBoard(boardID uint) ([]model.Card, error) {
	query := "SELECT " + cardColumns + " FROM cards WHERE board_id = ? AND archived_at IS NOT NULL ORDER BY archived_at DESC"
	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := []model.Card{}
	for rows.Next() {
		var c model.Card
		if err := scanCard(rows, &c); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *mysqlCardRepository) OtherPositions(listID uint, excludeID uint) ([]float64, error) {
	query := "SELECT position FROM cards WHERE list_id = ? AND archived_at IS NULL AND id != ? ORDER BY position ASC"
	rows, err := r.db.Query(query, listID, excludeID)
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

func (r *mysqlCardRepository) AddMember(cardID, userID uint) error {
	_, err := r.db.Exec("INSERT IGNORE INTO card_members (card_id, user_id) VALUES (?, ?)", cardID, userID)
	return err
}

func (r *mysqlCardRepository) RemoveMember(cardID, userID uint) error {
	_, err := r.db.Exec("DELETE FROM card_members WHERE card_id = ? AND user_id = ?", cardID, userID)
	return err
}

func (r *mysqlCardRepository) MemberIDs(cardID uint) ([]uint, error) {
	rows, err := r.db.Query("SELECT user_id FROM card_members WHERE card_id = ?", cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []uint{}
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *mysqlCardRepository) MemberIDsByBoard(boardID uint) (map[uint][]uint, error) {
	query := "SELECT cm.card_id, cm.user_id FROM card_members cm JOIN cards c ON c.id = cm.card_id WHERE c.board_id = ?"
	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[uint][]uint{}
	for rows.Next() {
		var cardID, userID uint
		if err := rows.Scan(&cardID, &userID); err != nil {
			return nil, err
		}
		result[cardID] = append(result[cardID], userID)
	}
	return result, nil
}

func (r *mysqlCardRepository) AddLabel(cardID, labelID uint) error {
	_, err := r.db.Exec("INSERT IGNORE INTO card_labels (card_id, label_id) VALUES (?, ?)", cardID, labelID)
	return err
}

func (r *mysqlCardRepository) RemoveLabel(cardID, labelID uint) error {
	_, err := r.db.Exec("DELETE FROM card_labels WHERE card_id = ? AND label_id = ?", cardID, labelID)
	return err
}

func (r *mysqlCardRepository) LabelIDs(cardID uint) ([]uint, error) {
	rows, err := r.db.Query("SELECT label_id FROM card_labels WHERE card_id = ?", cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []uint{}
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *mysqlCardRepository) LabelIDsByBoard(boardID uint) (map[uint][]uint, error) {
	query := "SELECT cl.card_id, cl.label_id FROM card_labels cl JOIN cards c ON c.id = cl.card_id WHERE c.board_id = ?"
	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[uint][]uint{}
	for rows.Next() {
		var cardID, labelID uint
		if err := rows.Scan(&cardID, &labelID); err != nil {
			return nil, err
		}
		result[cardID] = append(result[cardID], labelID)
	}
	return result, nil
}
