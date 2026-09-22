package model

type Label struct {
	ID      uint   `json:"id" db:"id"`
	BoardID uint   `json:"board_id" db:"board_id"`
	Name    string `json:"name" db:"name"`
	Color   string `json:"color" db:"color"`
}

type LabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color" binding:"required"`
}
