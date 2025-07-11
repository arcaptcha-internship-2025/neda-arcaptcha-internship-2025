package models

type Apartment struct {
	BaseModel
	ApartmentName string `json:"apartment_name"`
	Number        string `json:"number"`
	Address       string `json:"address"`
	UnitsCount    int    `json:"units_count"`
	ManagerID     int    `json:"manager_id"`
}
