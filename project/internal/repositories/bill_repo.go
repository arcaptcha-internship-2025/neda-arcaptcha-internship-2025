package repositories

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
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
	CreateBill(ctx context.Context, bill models.Bill) (int, error)
	GetBillByID(id int) (*models.Bill, error)
	GetBillsByApartmentID(apartmentID int) ([]models.Bill, error)
	UpdateBill(ctx context.Context, bill models.Bill) error
	DeleteBill(id int)
}
type billRepositoryImpl struct {
	db *sqlx.DB
}

func NewBillRepository(autoCreate bool, db *sqlx.DB) BillRepository {
	if autoCreate {
		if _, err := db.Exec(CREATE_BILLS_TABLE); err != nil {
			log.Fatalf("failed to create bills table: %v", err)
		}
	}
	return &billRepositoryImpl{db: db}
}

func (r *billRepositoryImpl) CreateBill(ctx context.Context, bill models.Bill) (int, error) {
	query := `INSERT INTO bills (apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url) 
			  VALUES (:apartment_id, :bill_type, :total_amount, :due_date, :billing_deadline, :description, :image_url) 
			  RETURNING id`
	var id int
	if err := r.db.QueryRowxContext(ctx, query, bill).Scan(&id); err != nil {
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

func (r *billRepositoryImpl) UpdateBill(ctx context.Context, bill models.Bill) error {
	query := `UPDATE bills 
			  SET apartment_id = :apartment_id, bill_type = :bill_type, total_amount = :total_amount, 
			      due_date = :due_date, billing_deadline = :billing_deadline, description = :description, 
			      image_url = :image_url, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, bill)
	if err != nil {
		return err
	}
	return nil
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
