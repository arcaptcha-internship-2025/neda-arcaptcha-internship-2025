package models

import "time"

type InvitationLink struct {
	BaseModel
	SenderID           int              `json:"sender_id"`
	SenderUsername     string           `json:"sender_username"`   //for notifications
	ReceiverUsername   string           `json:"receiver_username"` // telegram username
	ReceiverChatID     int64            `json:"receiver_chat_id"`  //for direct messaging
	ApartmentID        int              `json:"apartment_id"`
	ApartmentName      string           `json:"apartment_name"` //for notifications
	Token              string           `json:"token"`
	ExpiresAt          time.Time        `json:"expires_at"`
	Status             InvitationStatus `json:"status"`
	InviteURL          string           `json:"invite_url"`           // full invitation URL
	NotificationSentAt *time.Time       `json:"notification_sent_at"` //tracking if notification was sent
}

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRejected InvitationStatus = "rejected"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusNotified InvitationStatus = "notified"
)
