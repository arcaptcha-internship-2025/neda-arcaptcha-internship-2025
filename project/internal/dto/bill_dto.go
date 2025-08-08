package dto

import "github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"

type CreateBillRequest struct {
	BillType        models.BillType `json:"bill_type"`
	TotalAmount     float64         `json:"total_amount"`
	DueDate         string          `json:"due_date"`
	BillingDeadline string          `json:"billing_deadline"`
	Description     string          `json:"description"`
}

type PayBillsRequest struct {
	BillIDs []int `json:"bill_ids"`
}
