package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_PAYMENTS_TABLE = `CREATE TABLE IF NOT EXISTS payments(
		id SERIAL PRIMARY KEY,
		bill_id INTEGER REFERENCES bills(id) ON DELETE CASCADE,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		amount DECIMAL(12, 2) NOT NULL,
		is_succesful BOOLEAN DEFAULT FALSE,
		paid_at TIMESTAMP WITH TIME ZONE,
		payment_reference VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
)

// change
type PaymentRepository interface {
	CreatePayment(billID, userID int, amount float64, isPaid bool, paidAt *string, paymentReference string) (int, error)
	GetPaymentByID(id int) (*models.Payment, error)
}

type paymentRepositoryImpl struct {
	db *sqlx.DB
}

func NewPaymentRepository(autoCreate bool, db *sqlx.DB) (PaymentRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_PAYMENTS_TABLE); err != nil {
			return nil, err
		}
	}
	return &paymentRepositoryImpl{db: db}, nil
}

func (r *paymentRepositoryImpl) CreatePayment(billID, userID int, amount float64, isPaid bool, paidAt *string, paymentReference string) (int, error) {
	var id int
	query := `INSERT INTO payments (bill_id, user_id, amount, is_paid, paid_at, payment_reference) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := r.db.QueryRow(query, billID, userID, amount, isPaid, paidAt, paymentReference).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *paymentRepositoryImpl) GetPaymentByID(id int) (*models.Payment, error) {
	var payment models.Payment
	query := `SELECT id, bill_id, user_id, amount, is_paid, paid_at, payment_reference, created_at, updated_at 
			  FROM payments WHERE id = $1`
	err := r.db.Get(&payment, query, id)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
