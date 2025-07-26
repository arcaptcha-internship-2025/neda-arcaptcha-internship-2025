package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	return sqlxDB, mock
}

func TestNewBillRepository(t *testing.T) {
	tests := []struct {
		name       string
		autoCreate bool
		setupMock  func(sqlmock.Sqlmock)
		wantPanic  bool
	}{
		{
			name:       "AutoCreate true - success",
			autoCreate: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS bills").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantPanic: false,
		},
		{
			name:       "AutoCreate false - no table creation",
			autoCreate: false,
			setupMock:  func(mock sqlmock.Sqlmock) {},
			wantPanic:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			if tt.wantPanic {
				assert.Panics(t, func() {
					NewBillRepository(tt.autoCreate, db)
				})
			} else {
				repo := NewBillRepository(tt.autoCreate, db)
				assert.NotNil(t, repo)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBillRepository_CreateBill_Basic(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	bill := models.Bill{
		ApartmentID:     1,
		BillType:        models.WaterBill,
		TotalAmount:     100.50,
		DueDate:         "2024-01-15",
		BillingDeadline: "2024-01-10",
		Description:     "Water bill for January",
		ImageURL:        "https://example.com/bill.jpg",
	}

	// for sqlx named queries, we can't easily predict the exact arguments
	// so we ll just expect any INSERT query and return an id
	mock.ExpectQuery("INSERT INTO bills").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	repo := &billRepositoryImpl{db: db}
	ctx := context.Background()

	id, err := repo.CreateBill(ctx, bill)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestBillRepository_GetBillByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		setupMock func(sqlmock.Sqlmock)
		wantBill  *models.Bill
		wantErr   bool
	}{
		{
			name: "Success",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "apartment_id", "bill_type", "total_amount", "due_date",
					"billing_deadline", "description", "image_url", "created_at", "updated_at",
				}).AddRow(
					1, 1, "water", 100.50, "2024-01-15",
					"2024-01-10", "Water bill", "https://example.com/bill.jpg",
					time.Now(), time.Now(),
				)
				mock.ExpectQuery(`SELECT id, apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url, created_at, updated_at FROM bills WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantBill: &models.Bill{
				BaseModel: models.BaseModel{
					ID: 1,
				},
				ApartmentID:     1,
				BillType:        models.WaterBill,
				TotalAmount:     100.50,
				DueDate:         "2024-01-15",
				BillingDeadline: "2024-01-10",
				Description:     "Water bill",
				ImageURL:        "https://example.com/bill.jpg",
			},
			wantErr: false,
		},
		{
			name: "Bill not found",
			id:   999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url, created_at, updated_at FROM bills WHERE id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantBill: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := &billRepositoryImpl{db: db}

			bill, err := repo.GetBillByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, bill)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bill)
				assert.Equal(t, tt.wantBill.ID, bill.ID)
				assert.Equal(t, tt.wantBill.ApartmentID, bill.ApartmentID)
				assert.Equal(t, tt.wantBill.BillType, bill.BillType)
				assert.Equal(t, tt.wantBill.TotalAmount, bill.TotalAmount)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBillRepository_GetBillsByApartmentID(t *testing.T) {
	tests := []struct {
		name        string
		apartmentID int
		setupMock   func(sqlmock.Sqlmock)
		wantBills   []models.Bill
		wantErr     bool
	}{
		{
			name:        "Success with multiple bills",
			apartmentID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "apartment_id", "bill_type", "total_amount", "due_date",
					"billing_deadline", "description", "image_url", "created_at", "updated_at",
				}).
					AddRow(1, 1, "water", 100.50, "2024-01-15", "2024-01-10", "Water bill", "url1", time.Now(), time.Now()).
					AddRow(2, 1, "electricity", 75.25, "2024-01-20", "2024-01-15", "Electricity bill", "url2", time.Now(), time.Now())

				mock.ExpectQuery(`SELECT id, apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url, created_at, updated_at FROM bills WHERE apartment_id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantBills: []models.Bill{
				{
					BaseModel:       models.BaseModel{ID: 1},
					ApartmentID:     1,
					BillType:        models.WaterBill,
					TotalAmount:     100.50,
					DueDate:         "2024-01-15",
					BillingDeadline: "2024-01-10",
					Description:     "Water bill",
					ImageURL:        "url1",
				},
				{
					BaseModel:       models.BaseModel{ID: 2},
					ApartmentID:     1,
					BillType:        models.ElectricityBill,
					TotalAmount:     75.25,
					DueDate:         "2024-01-20",
					BillingDeadline: "2024-01-15",
					Description:     "Electricity bill",
					ImageURL:        "url2",
				},
			},
			wantErr: false,
		},
		{
			name:        "No bills found",
			apartmentID: 999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url, created_at, updated_at FROM bills WHERE apartment_id = \$1`).
					WithArgs(999).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "apartment_id", "bill_type", "total_amount", "due_date",
						"billing_deadline", "description", "image_url", "created_at", "updated_at",
					}))
			},
			wantBills: []models.Bill{},
			wantErr:   false,
		},
		{
			name:        "Database error",
			apartmentID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, apartment_id, bill_type, total_amount, due_date, billing_deadline, description, image_url, created_at, updated_at FROM bills WHERE apartment_id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			wantBills: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := &billRepositoryImpl{db: db}

			bills, err := repo.GetBillsByApartmentID(tt.apartmentID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, bills)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.wantBills), len(bills))
				for i, expectedBill := range tt.wantBills {
					assert.Equal(t, expectedBill.ID, bills[i].ID)
					assert.Equal(t, expectedBill.ApartmentID, bills[i].ApartmentID)
					assert.Equal(t, expectedBill.BillType, bills[i].BillType)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBillRepository_UpdateBill(t *testing.T) {
	tests := []struct {
		name      string
		bill      models.Bill
		setupMock func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "Success",
			bill: models.Bill{
				BaseModel:       models.BaseModel{ID: 1},
				ApartmentID:     1,
				BillType:        models.WaterBill,
				TotalAmount:     150.75,
				DueDate:         "2024-01-20",
				BillingDeadline: "2024-01-15",
				Description:     "Updated water bill",
				ImageURL:        "https://example.com/updated-bill.jpg",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE bills`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "Database error",
			bill: models.Bill{
				BaseModel: models.BaseModel{ID: 1},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE bills`).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := &billRepositoryImpl{db: db}
			ctx := context.Background()

			err := repo.UpdateBill(ctx, tt.bill)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBillRepository_DeleteBill(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		setupMock func(sqlmock.Sqlmock)
	}{
		{
			name: "Success",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM bills WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Database error (ignored)",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM bills WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := &billRepositoryImpl{db: db}

			// DeleteBill doesn't return an error, so we just call it
			repo.DeleteBill(tt.id)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBillRepository_GetPaymentByBillAndUser(t *testing.T) {
	tests := []struct {
		name        string
		billID      int
		userID      int
		setupMock   func(sqlmock.Sqlmock)
		wantPayment *models.Payment
		wantErr     bool
	}{
		{
			name:   "Success",
			billID: 1,
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at",
				}).AddRow(
					1, 1, 1, 100.50, time.Now(), "completed", time.Now(), time.Now(),
				)
				mock.ExpectQuery(`SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE bill_id = \$1 AND user_id = \$2`).
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantPayment: &models.Payment{
				BaseModel: models.BaseModel{ID: 1},
				BillID:    1,
				UserID:    1,
				Amount:    "100.50",
			},
			wantErr: false,
		},
		{
			name:   "Payment not found",
			billID: 1,
			userID: 999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE bill_id = \$1 AND user_id = \$2`).
					WithArgs(1, 999).
					WillReturnError(sql.ErrNoRows)
			},
			wantPayment: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := &billRepositoryImpl{db: db}

			payment, err := repo.GetPaymentByBillAndUser(tt.billID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, tt.wantPayment.ID, payment.ID)
				assert.Equal(t, tt.wantPayment.BillID, payment.BillID)
				assert.Equal(t, tt.wantPayment.UserID, payment.UserID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
