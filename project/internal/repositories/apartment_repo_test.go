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
)

func TestApartmentRepository_CreateApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewApartmentRepository(false, sqlxDB)

	apartment := models.Apartment{
		ApartmentName: "Erfan Apartments",
		Address:       "123 Enghelab St",
		UnitsCount:    10,
		ManagerID:     1,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO apartments`).
			WithArgs(apartment.ApartmentName, apartment.Address, apartment.UnitsCount, apartment.ManagerID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		id, err := repo.CreateApartment(context.Background(), apartment)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO apartments`).
			WithArgs(apartment.ApartmentName, apartment.Address, apartment.UnitsCount, apartment.ManagerID).
			WillReturnError(sql.ErrConnDone)

		_, err := repo.CreateApartment(context.Background(), apartment)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApartmentRepository_GetApartmentByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewApartmentRepository(false, sqlxDB)

	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "apartment_name", "address", "units_count", "manager_id", "created_at", "updated_at"}).
			AddRow(1, "Erfan Apartments", "123 Enghelab St", 10, 1, now, now)

		mock.ExpectQuery(`SELECT id, apartment_name, address, units_count, manager_id, created_at, updated_at FROM apartments WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(rows)

		apartment, err := repo.GetApartmentByID(1)
		assert.NoError(t, err)
		assert.Equal(t, "Erfan Apartments", apartment.ApartmentName)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, apartment_name, address, units_count, manager_id, created_at, updated_at FROM apartments WHERE id = \$1`).
			WithArgs(2).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetApartmentByID(2)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApartmentRepository_UpdateApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewApartmentRepository(false, sqlxDB)

	apartment := models.Apartment{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		ApartmentName: "Updated Name",
		Address:       "456 New St",
		UnitsCount:    15,
		ManagerID:     2,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE apartments SET`).
			WithArgs(apartment.ApartmentName, apartment.Address, apartment.UnitsCount, apartment.ManagerID, apartment.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateApartment(context.Background(), apartment)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(`UPDATE apartments SET`).
			WithArgs(apartment.ApartmentName, apartment.Address, apartment.UnitsCount, apartment.ManagerID, apartment.ID).
			WillReturnError(sql.ErrConnDone)

		err := repo.UpdateApartment(context.Background(), apartment)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApartmentRepository_DeleteApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewApartmentRepository(false, sqlxDB)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM apartments WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteApartment(1)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM apartments WHERE id = \$1`).
			WithArgs(2).
			WillReturnError(sql.ErrConnDone)

		err := repo.DeleteApartment(2)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
