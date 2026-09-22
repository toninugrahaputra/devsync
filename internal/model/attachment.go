package model

import "time"

type Attachment struct {
	ID           uint      `json:"id" db:"id"`
	CardID       uint      `json:"card_id" db:"card_id"`
	FileName     string    `json:"file_name" db:"file_name"`
	FilePath     string    `json:"file_path" db:"file_path"`
	FileSize     int64     `json:"file_size" db:"file_size"`
	UploadedBy   uint      `json:"uploaded_by" db:"uploaded_by"`
	UploaderName string    `json:"uploader_name" db:"uploader_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
