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
        billing_deadline DATE NOT NULL,
        description TEXT,
        Image_url VARCHAR(2000),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`
)

type BillRepository interface {
	CreateBill(apartmentID int, billType string, totalAmount float64, dueDate, billingDeadline string, description, imageURL string) (int, error)
	GetBillByID(id int) (*models.Bill, error)
	GetBillsByApartmentID(apartmentID int) ([]models.Bill, error)
	UpdateBill(id int, billType string, totalAmount float64, dueDate, billingDeadline string, description, imageURL string) error
	DeleteBill(id int)
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

func (r *billRepositoryImpl) CreateBill(apartmentID int, billType string, totalAmount float64, dueDate, billingDeadline string, description, imageURL string) (int, error) {
	var id int
	query := `INSERT INTO bills (apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := r.db.QueryRow(query, apartmentID, billType, totalAmount, dueDate, billingDeadline, description, imageURL).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *billRepositoryImpl) GetBillByID(id int) (*models.Bill, error) {
	var bill models.Bill
	query := `SELECT id, apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url, created_at, updated_at 
			  FROM bills WHERE id = $1`
	err := r.db.Get(&bill, query, id)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *billRepositoryImpl) GetBillsByApartmentID(apartmentID int) ([]models.Bill, error) {
	var bills []models.Bill
	query := `SELECT id, apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url, created_at, updated_at 
			  FROM bills WHERE apartment_id = $1`
	err := r.db.Select(&bills, query, apartmentID)
	if err != nil {
		return nil, err
	}
	return bills, nil
}
func (r *billRepositoryImpl) UpdateBill(id int, billType string, totalAmount float64, dueDate, billingDeadline string, description, imageURL string) error {
	query := `UPDATE bills SET bill_type = $1, total_amount = $2, due_date = $3, billing_deadline = $4, description = $5, image_url = $6, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = $7`
	_, err := r.db.Exec(query, billType, totalAmount, dueDate, billingDeadline, description, imageURL, id)
	return err
}
func (r *billRepositoryImpl) DeleteBill(id int) {
	query := `DELETE FROM bills WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		// handle error if needed, e.g., log it
		// for now, we just ignore the error
		return
	}
}
