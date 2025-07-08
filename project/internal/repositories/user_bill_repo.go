package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_USER_BILLS_TABLE = `CREATE TABLE IF NOT EXISTS user_bills(
        id SERIAL PRIMARY KEY,
        bill_id INTEGER NOT NULL REFERENCES bills(id),
        user_id INTEGER NOT NULL REFERENCES users(id),
        amount_due DECIMAL(10,2) NOT NULL,
        payment_status VARCHAR(20) NOT NULL,
        payment_link VARCHAR(255),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );`
)

type UserBillRepository interface {
	CreateUserBill(billID, userID int, amountDue float64, paymentStatus, paymentLink string) (int, error)
	GetUserBillByID(id int) (*models.User_bill, error)
	UpdateUserBill(id int, amountDue float64, paymentStatus, paymentLink string) error
	DeleteUserBill(id int) error
}

type userBillRepositoryImpl struct {
	db *sqlx.DB
}

func NewUserBillRepository(autoCreate bool, db *sqlx.DB) (UserBillRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_USER_BILLS_TABLE); err != nil {
			return nil, err
		}
	}
	return &userBillRepositoryImpl{db: db}, nil
}

func (r *userBillRepositoryImpl) CreateUserBill(billID, userID int, amountDue float64, paymentStatus, paymentLink string) (int, error) {
	var id int
	query := `INSERT INTO user_bills (bill_id, user_id, amount_due, payment_status, payment_link) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, billID, userID, amountDue, paymentStatus, paymentLink).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *userBillRepositoryImpl) GetUserBillByID(id int) (*models.User_bill, error) {
	var userBill models.User_bill
	query := `SELECT id, bill_id, user_id, amount_due, payment_status, payment_link, created_at, updated_at 
			  FROM user_bills WHERE id = $1`
	err := r.db.Get(&userBill, query, id)
	if err != nil {
		return nil, err
	}
	return &userBill, nil
}

func (r *userBillRepositoryImpl) UpdateUserBill(id int, amountDue float64, paymentStatus, paymentLink string) error {
	query := `UPDATE user_bills SET amount_due = $1, payment_status = $2, payment_link = $3, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = $4`
	_, err := r.db.Exec(query, amountDue, paymentStatus, paymentLink, id)
	return err
}

func (r *userBillRepositoryImpl) DeleteUserBill(id int) error {
	query := `DELETE FROM user_bills WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
