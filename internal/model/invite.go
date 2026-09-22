package model

import "time"

type Invite struct {
	ID          uint       `json:"id" db:"id"`
	WorkspaceID uint       `json:"workspace_id" db:"workspace_id"`
	Email       string     `json:"email" db:"email"`
	Role        string     `json:"role" db:"role"`
	Token       string     `json:"token" db:"token"`
	InvitedBy   uint       `json:"invited_by" db:"invited_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	AcceptedAt  *time.Time `json:"accepted_at" db:"accepted_at"`
}

type InviteRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// InviteResponse tells the caller what actually happened: if the invited
// email already has an account they're added immediately (no email needed).
// Otherwise a pending invite is created and InviteLink is always populated —
// EmailSent only reports whether SMTP is configured, since the link works
// for manual sharing (chat, WhatsApp, etc.) either way.
type InviteResponse struct {
	AlreadyMember bool   `json:"already_member"`
	EmailSent     bool   `json:"email_sent"`
	InviteLink    string `json:"invite_link,omitempty"`
}

type InvitePreview struct {
	WorkspaceName   string `json:"workspace_name"`
	Email           string `json:"email"`
	AlreadyAccepted bool   `json:"already_accepted"`
}
