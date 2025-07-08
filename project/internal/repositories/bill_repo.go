package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_BILLS_TABLE = `CREATE TABLE IF NOT EXISTS bills(
		id SERIAL PRIMARY KEY,
        apartment_id INTEGER NOT NULL REFERENCES apartments(id),
        bill_type VARCHAR(50) NOT NULL,
        total_amount DECIMAL(10,2) NOT NULL,
        due_date DATE NOT NULL,
        billing_period_start DATE NOT NULL,
        billing_period_end DATE NOT NULL,
        description TEXT,
        Image_url VARCHAR(255),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`
)

type BillRepository interface {
	CreateBill(apartmentID int, billType string, totalAmount float64, dueDate string, billingPeriodStart string, billingPeriodEnd string, description string, imageUrl string) (int, error)
	GetBillByID(id int) (*models.Bill, error)
}

type billRepositoryImpl struct {
	db *sqlx.DB
}

func NewBillRepository(autoCreate bool, db *sqlx.DB) (BillRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_BILLS_TABLE); err != nil {
			return nil, err
		}
	}
	return &billRepositoryImpl{db: db}, nil
}

func (r *billRepositoryImpl) CreateBill(apartmentID int, billType string, totalAmount float64, dueDate string, billingPeriodStart string, billingPeriodEnd string, description string, imageUrl string) (int, error) {
	var id int
	query := `INSERT INTO bills (apartment_id, bill_type, total_amount, due_date, billing_period_start, billing_period_end, description, image_url) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := r.db.QueryRow(query, apartmentID, billType, totalAmount, dueDate, billingPeriodStart, billingPeriodEnd, description, imageUrl).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *billRepositoryImpl) GetBillByID(id int) (*models.Bill, error) {
	var bill models.Bill
	query := `SELECT id, apartment_id, bill_type, total_amount, due_date, billing_period_start, billing_period_end, description, image_url, created_at, updated_at 
              FROM bills WHERE id = $1`
	err := r.db.Get(&bill, query, id)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}
