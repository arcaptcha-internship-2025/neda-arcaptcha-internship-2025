package models

import "time"

type Payment struct {
	BaseModel
	BillID        int           `json:"bill_id"`
	UserID        int           `json:"user_id"`
	Amount        string        `json:"amount"` // using string to handle decimal values
	PaidAt        time.Time     `json:"paid_at"`
	PaymentStatus PaymentStatus `json:"payment_status"`
}

type PaymentStatus string

const (
	Pending PaymentStatus = "pending"
	Paid    PaymentStatus = "paid"
	Failed  PaymentStatus = "failed"
)
