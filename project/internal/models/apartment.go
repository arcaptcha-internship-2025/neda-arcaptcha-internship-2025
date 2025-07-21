package models

type Apartment struct {
	BaseModel
	ApartmentName string `json:"apartment_name" db:"apartment_name"`
	Address       string `json:"address" db:"address"`
	UnitsCount    int    `json:"units_count" db:"units_count"`
	ManagerID     int    `json:"manager_id" db:"manager_id"`
}
