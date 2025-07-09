package models

type User_apartment struct {
	BaseModel
	UserID      int  `json:"user_id"`
	ApartmentID int  `json:"apartment_id"`
	IsManager   bool `json:"is_manager"`
}
