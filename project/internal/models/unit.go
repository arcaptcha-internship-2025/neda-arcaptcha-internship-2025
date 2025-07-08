package models

type Unit struct {
	ID          int    `json:"id"`
	Number      string `json:"number"`
	ApartmentID int    `json:"apartment_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
