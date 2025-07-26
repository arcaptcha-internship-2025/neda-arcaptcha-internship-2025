package models

type Bill struct {
	BaseModel
	ApartmentID     int      `json:"apartment_id" db:"apartment_id"`
	BillType        BillType `json:"bill_type" db:"bill_type"`
	TotalAmount     float64  `json:"total_amount" db:"total_amount"`
	DueDate         string   `json:"due_date" db:"due_date"`
	BillingDeadline string   `json:"billing_deadline" db:"billing_deadline"`
	Description     string   `json:"description" db:"description"`
	ImageURL        string   `json:"image_url" db:"image_url"`
}

type BillType string

const (
	WaterBill       BillType = "water"
	ElectricityBill BillType = "electricity"
	GasBill         BillType = "gas"
	MaintenanceBill BillType = "maintenance"
	OtherBill       BillType = "other"
)
