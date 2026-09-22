package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type AttachmentRepository interface {
	Create(a *model.Attachment) error
	GetByID(id uint) (*model.Attachment, error)
	Delete(id uint) error
	ListByCard(cardID uint) ([]model.Attachment, error)
}

type mysqlAttachmentRepository struct {
	db *sql.DB
}

func NewAttachmentRepository(db *sql.DB) AttachmentRepository {
	return &mysqlAttachmentRepository{db: db}
}

func (r *mysqlAttachmentRepository) Create(a *model.Attachment) error {
	res, err := r.db.Exec(
		"INSERT INTO attachments (card_id, file_name, file_path, file_size, uploaded_by) VALUES (?, ?, ?, ?, ?)",
		a.CardID, a.FileName, a.FilePath, a.FileSize, a.UploadedBy,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	a.ID = uint(id)
	return nil
}

func (r *mysqlAttachmentRepository) GetByID(id uint) (*model.Attachment, error) {
	var a model.Attachment
	query := "SELECT id, card_id, file_name, file_path, file_size, uploaded_by, created_at FROM attachments WHERE id = ?"
	err := r.db.QueryRow(query, id).Scan(&a.ID, &a.CardID, &a.FileName, &a.FilePath, &a.FileSize, &a.UploadedBy, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *mysqlAttachmentRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM attachments WHERE id = ?", id)
	return err
}

func (r *mysqlAttachmentRepository) ListByCard(cardID uint) ([]model.Attachment, error) {
	query := `SELECT a.id, a.card_id, a.file_name, a.file_path, a.file_size, a.uploaded_by, u.name, a.created_at
		FROM attachments a JOIN users u ON u.id = a.uploaded_by
		WHERE a.card_id = ? ORDER BY a.created_at DESC`
	rows, err := r.db.Query(query, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := []model.Attachment{}
	for rows.Next() {
		var a model.Attachment
		if err := rows.Scan(&a.ID, &a.CardID, &a.FileName, &a.FilePath, &a.FileSize, &a.UploadedBy, &a.UploaderName, &a.CreatedAt); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, nil
}
