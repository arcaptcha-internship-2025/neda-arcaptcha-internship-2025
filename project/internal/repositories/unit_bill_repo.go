package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/nda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_UNIT_BILLS_TABLE = `CREATE TABLE IF NOT EXISTS unit_bills(
        id SERIAL PRIMARY KEY,
        bill_id INTEGER NOT NULL REFERENCES bills(id),
        unit_id INTEGER NOT NULL REFERENCES units(id),
        amount_due DECIMAL(10,2) NOT NULL,
        payment_status VARCHAR(20) DEFAULT 'pending',
        payment_link VARCHAR(255),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );`
)

type UnitBillRepository interface {
	CreateUnitBill(billID, unitID int, amountDue float64, paymentStatus, paymentLink string) (int, error)
	GetUnitBillByID(id int) (*models.Unit_bills, error)
}

type unitBillRepositoryImpl struct {
	db *sqlx.DB
}

func NewUnitBillRepository(autoCreate bool, db *sqlx.DB) (UnitBillRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_UNIT_BILLS_TABLE); err != nil {
			return nil, err
		}
	}
	return &unitBillRepositoryImpl{db: db}, nil
}
func (r *unitBillRepositoryImpl) CreateUnitBill(billID, unitID int, amountDue float64, paymentStatus, paymentLink string) (int, error) {
	var id int
	query := `INSERT INTO unit_bills (bill_id, unit_id, amount_due, payment_status, payment_link) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, billID, unitID, amountDue, paymentStatus, paymentLink).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *unitBillRepositoryImpl) GetUnitBillByID(id int) (*models.Unit_bills, error) {
	var unitBill models.Unit_bills
	query := `SELECT id, bill_id, unit_id, amount_due, payment_status, payment_link, created_at, updated_at 
              FROM unit_bills WHERE id = $1`
	err := r.db.Get(&unitBill, query, id)
	if err != nil {
		return nil, err
	}
	return &unitBill, nil
}
