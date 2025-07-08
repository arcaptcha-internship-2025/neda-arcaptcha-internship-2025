package models

type Notification struct {
	ID               int    `json:"id"`
	SenderID         int    `json:"sender_id"`
	ApartmentID      int    `json:"apartment_id"`
	NotificationType string `json:"notification_type"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	Is_sent          bool   `json:"is_sent"`
	CreatedAt        string `json:"created_at"` // using string for date/time representation
	UpdatedAt        string `json:"updated_at"`
}
