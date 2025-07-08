package models

type Apartment struct {
	ID        int    `json:"id"`
	Number    string `json:"number"`
	Address   string `json:"address"`
	UnitsNum  int    `json:"units_num"`
	ManagerID int    `json:"manager_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
