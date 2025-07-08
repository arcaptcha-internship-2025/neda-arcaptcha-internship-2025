package models

type InvitationLink struct {
	ID          int    `json:"id"`
	ApartmentID int    `json:"apartment_id"`
	UnitID      int    `json:"unit_id"`
	Token       string `json:"token"`
	CreatedBy   int    `json:"created_by"`
	ExpiresAt   string `json:"expires_at"` // using string for date/time representation
	IsUsed      bool   `json:"is_used"`
	UsedBy      int    `json:"used_by"`
	UsedAt      string `json:"used_at"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
