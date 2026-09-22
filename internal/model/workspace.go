package model

import "time"

type Workspace struct {
	ID        uint      `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedBy uint      `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// Role is the requesting user's role in this workspace. Only populated when listing "my workspaces".
	Role string `json:"role,omitempty" db:"-"`
}

type WorkspaceRequest struct {
	Name string `json:"name" binding:"required"`
}

type WorkspaceMember struct {
	WorkspaceID uint      `json:"workspace_id" db:"workspace_id"`
	UserID      uint      `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Email       string    `json:"email" db:"email"`
	AvatarPath  *string   `json:"avatar_path" db:"avatar_path"`
	Role        string    `json:"role" db:"role"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

type AddMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}
