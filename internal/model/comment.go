package model

import "time"

type Comment struct {
	ID        uint      `json:"id" db:"id"`
	CardID    uint      `json:"card_id" db:"card_id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	UserName  string    `json:"user_name" db:"user_name"`
	Body      string    `json:"body" db:"body"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CommentRequest struct {
	Body string `json:"body" binding:"required"`
}
