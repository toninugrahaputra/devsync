package model

import "time"

type Board struct {
	ID              uint       `json:"id" db:"id"`
	WorkspaceID     uint       `json:"workspace_id" db:"workspace_id"`
	Name            string     `json:"name" db:"name"`
	BackgroundColor string     `json:"background_color" db:"background_color"`
	CreatedBy       uint       `json:"created_by" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	ArchivedAt      *time.Time `json:"archived_at,omitempty" db:"archived_at"`
}

type BoardRequest struct {
	Name            string `json:"name" binding:"required"`
	BackgroundColor string `json:"background_color"`
}

type BoardMember struct {
	BoardID    uint    `json:"board_id" db:"board_id"`
	UserID     uint    `json:"user_id" db:"user_id"`
	Name       string  `json:"name" db:"name"`
	Email      string  `json:"email" db:"email"`
	AvatarPath *string `json:"avatar_path" db:"avatar_path"`
	Role       string  `json:"role" db:"role"`
}

type BoardDetail struct {
	Board
	Lists   []ListDetail  `json:"lists"`
	Labels  []Label       `json:"labels"`
	Members []BoardMember `json:"members"`
}
