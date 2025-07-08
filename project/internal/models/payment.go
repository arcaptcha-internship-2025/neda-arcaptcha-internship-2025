package models

type Payment struct {
	ID         int    `json:"id"`
	BillID     int    `json:"bill_id"`
	UserID     int    `json:"user_id"`
	Amount     string `json:"amount"`  // using string to handle decimal values
	PaidAt     string `json:"paid_at"` // using string for date/time representation
	PaymentRef string `json:"payment_reference"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
