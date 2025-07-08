package models

type Bill struct {
	ID              int     `json:"id"`
	ApartmentID     int     `json:"apartment_id"`
	BillType        string  `json:"bill_type"`
	TotalAmount     float64 `json:"total_amount"`
	DueDate         string  `json:"due_date"`
	BillingDeadline string  `json:"billing_deadline"`
	Description     string  `json:"description"`
	ImageURL        string  `json:"image_url"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}
