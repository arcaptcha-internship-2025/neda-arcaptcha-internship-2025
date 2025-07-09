package payment

import "fmt"

//interface method:paybills-> slice of bills ->loop on slice->does nothing(mock)->update bill id in payment repo stutus to done

type Payment interface {
	PayBills(billIDs []int) error
}

type paymentImpl struct {
}

func NewPayment() Payment {
	return &paymentImpl{}
}

func (p *paymentImpl) PayBills(billIDs []int) error {
	for _, billID := range billIDs {
		// for demonstration, we will just print the bill ID
		fmt.Printf("Paying bill with ID: %d\n", billID)
	}
	return nil // so the payment was successful
}
