package models

type InvitationLink struct {
	ID          int    `json:"id"`
	SenderID    int    `json:"sender_id"`
	ReceiverID  int    `json:"receiver_id"`
	ApartmentID int    `json:"apartment_id"`
	UnitID      int    `json:"unit_id"`
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"` // using string for date/time representation
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
