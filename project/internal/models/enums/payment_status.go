package enums

type PaymentStatus string

const (
	Pending PaymentStatus = "pending"
	Paid    PaymentStatus = "paid"
	Failed  PaymentStatus = "failed"
)
