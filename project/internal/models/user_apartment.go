package models

type User_apartment struct {
	BaseModel
	UserID      int  `json:"user_id" db:"user_id"`
	ApartmentID int  `json:"apartment_id" db:"apartment_id"`
	IsManager   bool `json:"is_manager" db:"is_manager"`
}
