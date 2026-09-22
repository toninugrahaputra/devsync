package model

type Checklist struct {
	ID       uint    `json:"id" db:"id"`
	CardID   uint    `json:"card_id" db:"card_id"`
	Title    string  `json:"title" db:"title"`
	Position float64 `json:"position" db:"position"`
}

type ChecklistItem struct {
	ID          uint    `json:"id" db:"id"`
	ChecklistID uint    `json:"checklist_id" db:"checklist_id"`
	Text        string  `json:"text" db:"text"`
	IsChecked   bool    `json:"is_checked" db:"is_checked"`
	Position    float64 `json:"position" db:"position"`
}

type ChecklistDetail struct {
	Checklist
	Items []ChecklistItem `json:"items"`
}

type ChecklistRequest struct {
	Title string `json:"title" binding:"required"`
}

type ChecklistItemRequest struct {
	Text string `json:"text" binding:"required"`
}

type UpdateChecklistItemRequest struct {
	Text      *string `json:"text"`
	IsChecked *bool   `json:"is_checked"`
}
