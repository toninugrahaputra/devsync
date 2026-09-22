package model

type List struct {
	ID       uint    `json:"id" db:"id"`
	BoardID  uint    `json:"board_id" db:"board_id"`
	Name     string  `json:"name" db:"name"`
	Position float64 `json:"position" db:"position"`
}

type ListRequest struct {
	Name string `json:"name" binding:"required"`
}

type MoveListRequest struct {
	Index int `json:"index"`
}

type ListDetail struct {
	List
	Cards []CardSummary `json:"cards"`
}
