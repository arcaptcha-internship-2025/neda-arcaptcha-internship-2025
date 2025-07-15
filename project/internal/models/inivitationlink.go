package models

import "time"

type InvitationLink struct {
	BaseModel
	SenderID         int              `json:"sender_id"`
	ReceiverUsername string           `json:"receiver_username"` // telegram username
	ChatID           *int64           `json:"chat_id,omitempty"` // telegram chat ID if available
	ApartmentID      int              `json:"apartment_id"`
	Token            string           `json:"token"`
	ExpiresAt        time.Time        `json:"expires_at"`
	Status           InvitationStatus `json:"status"`
}

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRejected InvitationStatus = "rejected"
)
