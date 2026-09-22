package model

import "time"

type Card struct {
	ID          uint       `json:"id" db:"id"`
	ListID      uint       `json:"list_id" db:"list_id"`
	BoardID     uint       `json:"board_id" db:"board_id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Position    float64    `json:"position" db:"position"`
	DueDate     *time.Time `json:"due_date" db:"due_date"`
	StartDate   *time.Time `json:"start_date" db:"start_date"`
	IsCompleted bool       `json:"is_completed" db:"is_completed"`
	CoverColor  *string    `json:"cover_color" db:"cover_color"`
	CreatedBy   uint       `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// CardSummary is what's embedded in a board's list columns (no members/labels/checklists detail, just ids).
type CardSummary struct {
	Card
	MemberIDs        []uint `json:"member_ids"`
	LabelIDs         []uint `json:"label_ids"`
	ChecklistTotal   int    `json:"checklist_total"`
	ChecklistChecked int    `json:"checklist_checked"`
}

// ChecklistCounts is the "checked/total" badge shown on a card face.
type ChecklistCounts struct {
	Total   int
	Checked int
}

// CardDetail is the full payload returned when opening a single card.
type CardDetail struct {
	Card
	Members     []BoardMember     `json:"members"`
	Labels      []Label           `json:"labels"`
	Checklists  []ChecklistDetail `json:"checklists"`
	Comments    []Comment         `json:"comments"`
	Attachments []Attachment      `json:"attachments"`
	Activities  []Activity        `json:"activities"`
}

type CreateCardRequest struct {
	Title string `json:"title" binding:"required"`
}

type UpdateCardRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	DueDate     *string `json:"due_date"`
	StartDate   *string `json:"start_date"`
	IsCompleted *bool   `json:"is_completed"`
	CoverColor  *string `json:"cover_color"`
}

type MoveCardRequest struct {
	ListID uint `json:"list_id" binding:"required"`
	Index  int  `json:"index"`
}
