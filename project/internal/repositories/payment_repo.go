package repositories

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

const (
	CREATE_PAYMENTS_TABLE = `CREATE TABLE IF NOT EXISTS payments(
		id SERIAL PRIMARY KEY,
		bill_id INTEGER REFERENCES bills(id) ON DELETE CASCADE,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		amount DECIMAL(12, 2) NOT NULL,
		paid_at TIMESTAMP WITH TIME ZONE,
		payment_status VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment models.Payment) (int, error)
	GetPaymentByID(id int) (*models.Payment, error)
	UpdatePayment(id int, payment models.Payment) error
	UpdatePaymentsStatus(ctx context.Context, payments []models.Payment)
	DeletePayment(id int) error
}

type paymentRepositoryImpl struct {
	db *sqlx.DB
}

func NewPaymentRepository(autoCreate bool, db *sqlx.DB) PaymentRepository {
	if autoCreate {
		if _, err := db.Exec(CREATE_PAYMENTS_TABLE); err != nil {
			log.Fatalf("failed to create payments table: %v", err)
		}
	}
	return &paymentRepositoryImpl{db: db}
}

func (r *paymentRepositoryImpl) CreatePayment(ctx context.Context, payment models.Payment) (int, error) {
	query := `INSERT INTO payments (bill_id, user_id, amount, paid_at, payment_status) 
			  VALUES (:bill_id, :user_id, :amount, :paid_at, :payment_status) 
			  RETURNING id`
	var id int
	if err := r.db.QueryRowxContext(ctx, query, payment).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *paymentRepositoryImpl) GetPaymentByID(id int) (*models.Payment, error) {
	var payment models.Payment
	query := `SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at 
			  FROM payments WHERE id = $1`
	err := r.db.Get(&payment, query, id)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepositoryImpl) UpdatePayment(id int, payment models.Payment) error {
	query := `UPDATE payments 
			  SET bill_id = :bill_id, user_id = :user_id, amount = :amount, 
			      paid_at = :paid_at, payment_status = :payment_status, 
			      updated_at = CURRENT_TIMESTAMP 
			  WHERE id = :id`
	payment.ID = id // Set the ID for the update query
	_, err := r.db.NamedExecContext(context.Background(), query, payment)
	if err != nil {
		log.Printf("error updating payment with ID %d: %v", id, err)
	}
	return err
}

func (r *paymentRepositoryImpl) UpdatePaymentsStatus(ctx context.Context, payments []models.Payment) {
	query := `UPDATE payments SET payment_status = :payment_status, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = :id`
	for _, payment := range payments {
		_, err := r.db.NamedExecContext(ctx, query, payment)
		if err != nil {
			log.Printf("error updating payment status for ID %d: %v", payment.ID, err)
			continue
			// log the error and continue processing other payments
			// we might want to handle the error differently based on your requirements
			// foeexample we could return the error or retry the operation
			// now we just log the error and continue with the next payment
			// this way if one payment fails, it won't stop the processing of others.
			// we can also choose to return the error if you want to stop processing on the first error.
			// return err
		}
	}
}

func (r *paymentRepositoryImpl) DeletePayment(id int) error {
	query := `DELETE FROM payments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
