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

func setupPaymentTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	return sqlxDB, mock
}

func TestNewPaymentRepository(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	t.Run("with autoCreate true", func(t *testing.T) {
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS payments").WillReturnResult(sqlmock.NewResult(0, 0))

		repo := NewPaymentRepository(true, db)
		assert.NotNil(t, repo)

		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("with autoCreate false", func(t *testing.T) {
		repo := NewPaymentRepository(false, db)
		assert.NotNil(t, repo)

		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_CreatePayment(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	ctx := context.Background()

	payment := models.Payment{
		BillID:        1,
		UserID:        1,
		Amount:        "100.50",
		PaidAt:        time.Now(),
		PaymentStatus: models.Pending,
	}

	t.Run("successful creation", func(t *testing.T) {
		expectedID := 1
		mock.ExpectQuery("INSERT INTO payments").
			WithArgs(payment.BillID, payment.UserID, payment.Amount, payment.PaidAt, payment.PaymentStatus).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

		id, err := repo.CreatePayment(ctx, payment)

		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO payments").
			WithArgs(payment.BillID, payment.UserID, payment.Amount, payment.PaidAt, payment.PaymentStatus).
			WillReturnError(sql.ErrConnDone)

		id, err := repo.CreatePayment(ctx, payment)

		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Equal(t, sql.ErrConnDone, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_GetPaymentByID(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	paymentID := 1

	t.Run("successful retrieval", func(t *testing.T) {
		expectedPayment := models.Payment{
			BaseModel: models.BaseModel{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			BillID:        1,
			UserID:        1,
			Amount:        "100.50",
			PaidAt:        time.Now(),
			PaymentStatus: models.Paid,
		}

		rows := sqlmock.NewRows([]string{"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at"}).
			AddRow(expectedPayment.ID, expectedPayment.BillID, expectedPayment.UserID, expectedPayment.Amount,
				expectedPayment.PaidAt, expectedPayment.PaymentStatus, expectedPayment.CreatedAt, expectedPayment.UpdatedAt)

		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE id = \\$1").
			WithArgs(paymentID).
			WillReturnRows(rows)

		payment, err := repo.GetPaymentByID(paymentID)

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, expectedPayment.ID, payment.ID)
		assert.Equal(t, expectedPayment.BillID, payment.BillID)
		assert.Equal(t, expectedPayment.UserID, payment.UserID)
		assert.Equal(t, expectedPayment.Amount, payment.Amount)
		assert.Equal(t, expectedPayment.PaymentStatus, payment.PaymentStatus)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("payment not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE id = \\$1").
			WithArgs(paymentID).
			WillReturnError(sql.ErrNoRows)

		payment, err := repo.GetPaymentByID(paymentID)

		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Equal(t, sql.ErrNoRows, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_GetPaymentByBillAndUser(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	billID, userID := 1, 1

	t.Run("successful retrieval", func(t *testing.T) {
		expectedPayment := models.Payment{
			BaseModel: models.BaseModel{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			BillID:        billID,
			UserID:        userID,
			Amount:        "100.50",
			PaidAt:        time.Now(),
			PaymentStatus: models.Paid,
		}

		rows := sqlmock.NewRows([]string{"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at"}).
			AddRow(expectedPayment.ID, expectedPayment.BillID, expectedPayment.UserID, expectedPayment.Amount,
				expectedPayment.PaidAt, expectedPayment.PaymentStatus, expectedPayment.CreatedAt, expectedPayment.UpdatedAt)

		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE bill_id = \\$1 AND user_id = \\$2").
			WithArgs(billID, userID).
			WillReturnRows(rows)

		payment, err := repo.GetPaymentByBillAndUser(billID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, expectedPayment.BillID, payment.BillID)
		assert.Equal(t, expectedPayment.UserID, payment.UserID)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_GetPaymentsByUser(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	userID := 1

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at"}).
			AddRow(1, 1, userID, "100.50", time.Now(), models.Paid, time.Now(), time.Now()).
			AddRow(2, 2, userID, "200.00", time.Now(), models.Pending, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE user_id = \\$1").
			WithArgs(userID).
			WillReturnRows(rows)

		payments, err := repo.GetPaymentsByUser(userID)

		assert.NoError(t, err)
		assert.Len(t, payments, 2)
		assert.Equal(t, userID, payments[0].UserID)
		assert.Equal(t, userID, payments[1].UserID)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("no payments found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at"})

		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE user_id = \\$1").
			WithArgs(userID).
			WillReturnRows(rows)

		payments, err := repo.GetPaymentsByUser(userID)

		assert.NoError(t, err)
		assert.Len(t, payments, 0)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_UpdatePaymentStatus(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	ctx := context.Background()

	payment := models.Payment{
		BaseModel:     models.BaseModel{ID: 1},
		PaymentStatus: models.Paid,
		PaidAt:        time.Now(),
	}

	t.Run("successful update", func(t *testing.T) {
		mock.ExpectExec("UPDATE payments SET").
			WithArgs(payment.PaymentStatus, payment.PaidAt, payment.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdatePaymentStatus(ctx, payment)

		assert.NoError(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("UPDATE payments SET").
			WithArgs(payment.PaymentStatus, payment.PaidAt, payment.ID).
			WillReturnError(sql.ErrConnDone)

		err := repo.UpdatePaymentStatus(ctx, payment)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_UpdatePaymentsStatus(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	ctx := context.Background()

	payments := []models.Payment{
		{
			BaseModel:     models.BaseModel{ID: 1},
			PaymentStatus: models.Paid,
			PaidAt:        time.Now(),
		},
		{
			BaseModel:     models.BaseModel{ID: 2},
			PaymentStatus: models.Paid,
			PaidAt:        time.Now(),
		},
	}

	t.Run("successful batch update", func(t *testing.T) {
		mock.ExpectBegin()

		for _, payment := range payments {
			mock.ExpectExec("UPDATE payments SET").
				WithArgs(payment.PaymentStatus, payment.PaidAt, payment.ID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		mock.ExpectCommit()

		err := repo.UpdatePaymentsStatus(ctx, payments)

		assert.NoError(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("transaction rollback on error", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec("UPDATE payments SET").
			WithArgs(payments[0].PaymentStatus, payments[0].PaidAt, payments[0].ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("UPDATE payments SET").
			WithArgs(payments[1].PaymentStatus, payments[1].PaidAt, payments[1].ID).
			WillReturnError(sql.ErrConnDone)

		mock.ExpectRollback()

		err := repo.UpdatePaymentsStatus(ctx, payments)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_DeletePayment(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	paymentID := 1

	t.Run("successful deletion", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM payments WHERE id = \\$1").
			WithArgs(paymentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeletePayment(paymentID)

		assert.NoError(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM payments WHERE id = \\$1").
			WithArgs(paymentID).
			WillReturnError(sql.ErrConnDone)

		err := repo.DeletePayment(paymentID)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_GetPendingPaymentsByUser(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	userID := 1

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at",
		}).AddRow(1, 1, userID, "50.00", time.Now(), models.Pending, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE user_id = \\$1 and payment_status = 'pending'").
			WithArgs(userID).
			WillReturnRows(rows)

		payments, err := repo.GetPendingPaymentsByUser(userID)

		assert.NoError(t, err)
		assert.Len(t, payments, 1)
		assert.Equal(t, userID, payments[0].UserID)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("no pending payments", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at",
		})

		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE user_id = \\$1 and payment_status = 'pending'").
			WithArgs(userID).
			WillReturnRows(rows)

		payments, err := repo.GetPendingPaymentsByUser(userID)

		assert.NoError(t, err)
		assert.Len(t, payments, 0)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestPaymentRepository_GetPaymentsByBill(t *testing.T) {
	db, mock := setupPaymentTestDB(t)
	defer db.Close()

	repo := &paymentRepositoryImpl{db: db}
	billID := 1

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at",
		}).AddRow(1, billID, 1, "75.00", time.Now(), models.Paid, time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE bill_id = \\$1").
			WithArgs(billID).
			WillReturnRows(rows)

		payments, err := repo.GetPaymentsByBill(billID)

		assert.NoError(t, err)
		assert.Len(t, payments, 1)
		assert.Equal(t, billID, payments[0].BillID)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("no payments for bill", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "bill_id", "user_id", "amount", "paid_at", "payment_status", "created_at", "updated_at",
		})

		mock.ExpectQuery("SELECT id, bill_id, user_id, amount, paid_at, payment_status, created_at, updated_at FROM payments WHERE bill_id = \\$1").
			WithArgs(billID).
			WillReturnRows(rows)

		payments, err := repo.GetPaymentsByBill(billID)

		assert.NoError(t, err)
		assert.Len(t, payments, 0)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
