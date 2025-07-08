package enums

type BillType string

const (
	WaterBill       BillType = "water"
	ElectricityBill BillType = "electricity"
	GasBill         BillType = "gas"
	MaintenanceBill BillType = "maintenance"
	OtherBill       BillType = "other"
)
