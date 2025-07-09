package models

type Payment struct {
	ID            int           `json:"id"`
	BillID        int           `json:"bill_id"`
	UserID        int           `json:"user_id"`
	Amount        string        `json:"amount"`  // using string to handle decimal values
	PaidAt        string        `json:"paid_at"` // using string for date/time representation
	PaymentStatus PaymentStatus `json:"payment_status"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
}

type PaymentStatus string

const (
	Pending PaymentStatus = "pending"
	Paid    PaymentStatus = "paid"
	Failed  PaymentStatus = "failed"
)
