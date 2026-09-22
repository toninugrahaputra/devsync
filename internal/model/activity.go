package model

import "time"

// Activity is a single auto-generated audit-log line ("Alice moved this card
// to Done") shown interleaved with comments in the card's "Comments and
// activity" feed.
type Activity struct {
	ID        uint      `json:"id" db:"id"`
	CardID    uint      `json:"card_id" db:"card_id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	UserName  string    `json:"user_name" db:"user_name"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
