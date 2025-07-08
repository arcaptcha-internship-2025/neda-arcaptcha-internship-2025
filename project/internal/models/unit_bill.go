package models

type Unit_bills struct {
	ID            int     `json:"id"`
	UnitID        int     `json:"unit_id"`
	BillID        int     `json:"bill_id"`
	AmountDue     float64 `json:"amount_due"`
	PaymentStatus string  `json:"payment_status"` // e.g., 'pending', 'paid'
	PaymentLink   string  `json:"payment_link"`   // URL for payment
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}
