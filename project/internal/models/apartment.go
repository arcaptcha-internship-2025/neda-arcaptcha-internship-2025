package models

type Apartment struct {
	BaseModel
	Number     string `json:"number"`
	Address    string `json:"address"`
	UnitsCount int    `json:"units_count"`
	ManagerID  int    `json:"manager_id"`
}
