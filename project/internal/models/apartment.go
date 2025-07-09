package models

type Apartment struct {
	ID         int    `json:"id"`
	Number     string `json:"number"`
	Address    string `json:"address"`
	UnitsCount int    `json:"units_count"`
	ManagerID  int    `json:"manager_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
