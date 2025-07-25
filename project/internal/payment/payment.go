package payment

import (
	"log"
)

type Payment interface {
	PayBills(billIDs []int) error
}

type paymentImpl struct {
}

func NewPayment() Payment {
	return &paymentImpl{}
}

func (p *paymentImpl) PayBills(billIDs []int) error {
	//mock payment processing
	for _, billID := range billIDs {
		log.Printf("Processing mock payment for bill ID: %d", billID)
	}

	return nil
}
