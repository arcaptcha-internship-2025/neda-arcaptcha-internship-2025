package models

type Notification struct {
	ID               int    `json:"id"`
	SenderID         int    `json:"sender_id"`
	ReceiverID       int    `json:"receiver_id"`
	ApartmentID      int    `json:"apartment_id"`
	NotificationType string `json:"notification_type"` // e.g., 'bill', 'announcement'
	Title            string `json:"title"`
	Content          string `json:"content"`
	IsRead           bool   `json:"is_read"`
	CreatedAt        string `json:"created_at"` // using string for date/time representation
	UpdatedAt        string `json:"updated_at"`
}
