package models

import "time"

type InvitationLink struct {
	BaseModel
	SenderID    int       `json:"sender_id"`
	ReceiverID  int       `json:"receiver_id"`
	ApartmentID int       `json:"apartment_id"`
	UnitID      int       `json:"unit_id"`
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
