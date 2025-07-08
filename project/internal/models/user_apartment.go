package models

type User_apartment struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	ApartmentID int    `json:"apartment_id"`
	IsManager   bool   `json:"is_manager"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
