package models

type Bill struct {
	BaseModel
	ApartmentID     int      `json:"apartment_id"`
	BillType        BillType `json:"bill_type"`
	TotalAmount     float64  `json:"total_amount"`
	DueDate         string   `json:"due_date"`
	BillingDeadline string   `json:"billing_deadline"`
	Description     string   `json:"description"`
	ImageURL        string   `json:"image_url"`
}

type BillType string

const (
	WaterBill       BillType = "water"
	ElectricityBill BillType = "electricity"
	GasBill         BillType = "gas"
	MaintenanceBill BillType = "maintenance"
	OtherBill       BillType = "other"
)
