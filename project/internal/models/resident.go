package models

type Resident struct {
	ID                int    `json:"id"`
	UserID            int    `json:"user_id"`
	UnitID            int    `json:"unit_id"`
	MoveInDate        string `json:"move_in_date"`
	MoveOutDate       string `json:"move_out_date,omitempty"`
	IsPrimaryResident bool   `json:"is_primary_resident"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}
