package models

type User_bill struct {
	ID            int     `json:"id"`
	BillID        int     `json:"bill_id"`
	UserID        int     `json:"user_id"`
	AmountDue     float64 `json:"amount_due"`
	PaymentStatus string  `json:"payment_status"`
	PaymentLink   string  `json:"payment_link"` // URL for payment
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}
