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

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	return sqlxDB, mock
}

func TestNewUserApartmentRepository(t *testing.T) {
	t.Run("with autoCreate true", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mock.ExpectExec("CREATE TABLE IF NOT EXIST user_apartments").WillReturnResult(sqlmock.NewResult(0, 0))

		repo := NewUserApartmentRepository(true, db)
		assert.NotNil(t, repo)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with autoCreate false", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewUserApartmentRepository(false, db)
		assert.NotNil(t, repo)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCreateUserApartment(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := &userApartmentRepositoryImpl{db: db}
	ctx := context.Background()

	userApartment := models.User_apartment{
		UserID:      1,
		ApartmentID: 2,
		IsManager:   true,
	}

	t.Run("successful creation", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO user_apartments").
			WithArgs(userApartment.UserID, userApartment.ApartmentID, userApartment.IsManager).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateUserApartment(ctx, userApartment)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO user_apartments").
			WithArgs(userApartment.UserID, userApartment.ApartmentID, userApartment.IsManager).
			WillReturnError(sql.ErrConnDone)

		err := repo.CreateUserApartment(ctx, userApartment)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetUserApartmentByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := &userApartmentRepositoryImpl{db: db}
	userID := 1
	apartmentID := 2
	now := time.Now()

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"user_id", "apartment_id", "is_manager", "created_at", "updated_at"}).
			AddRow(userID, apartmentID, true, now, now)

		mock.ExpectQuery("SELECT user_id, apartment_id, is_manager, created_at, updated_at FROM user_apartments").
			WithArgs(userID, apartmentID).
			WillReturnRows(rows)

		result, err := repo.GetUserApartmentByID(userID, apartmentID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, apartmentID, result.ApartmentID)
		assert.True(t, result.IsManager)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT user_id, apartment_id, is_manager, created_at, updated_at FROM user_apartments").
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetUserApartmentByID(userID, apartmentID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT user_id, apartment_id, is_manager, created_at, updated_at FROM user_apartments").
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrConnDone)

		result, err := repo.GetUserApartmentByID(userID, apartmentID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateUserApartment(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := &userApartmentRepositoryImpl{db: db}
	ctx := context.Background()

	userApartment := models.User_apartment{
		UserID:      1,
		ApartmentID: 2,
		IsManager:   false,
	}

	t.Run("successful update", func(t *testing.T) {
		mock.ExpectExec("UPDATE user_apartments").
			WithArgs(userApartment.IsManager, userApartment.UserID, userApartment.ApartmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateUserApartment(ctx, userApartment)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no rows affected", func(t *testing.T) {
		mock.ExpectExec("UPDATE user_apartments").
			WithArgs(userApartment.IsManager, userApartment.UserID, userApartment.ApartmentID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateUserApartment(ctx, userApartment)
		assert.NoError(t, err) //doesnt check affected rows
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("UPDATE user_apartments").
			WithArgs(userApartment.IsManager, userApartment.UserID, userApartment.ApartmentID).
			WillReturnError(sql.ErrConnDone)

		err := repo.UpdateUserApartment(ctx, userApartment)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteUserApartment(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := &userApartmentRepositoryImpl{db: db}
	userID := 1
	apartmentID := 2

	t.Run("successful deletion", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM user_apartments").
			WithArgs(userID, apartmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteUserApartment(userID, apartmentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no rows affected", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM user_apartments").
			WithArgs(userID, apartmentID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteUserApartment(userID, apartmentID)
		assert.NoError(t, err) //doesnt check affected rows
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM user_apartments").
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrConnDone)

		err := repo.DeleteUserApartment(userID, apartmentID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetResidentsInApartment(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := &userApartmentRepositoryImpl{db: db}
	apartmentID := 1
	now := time.Now()

	t.Run("successful retrieval with residents", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "phone", "full_name", "user_type", "created_at", "updated_at",
		}).
			AddRow(1, "user1", "user1@example.com", "123456789", "User One", "resident", now, now).
			AddRow(2, "user2", "user2@example.com", "987654321", "User Two", "resident", now, now)

		mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.phone, u.full_name, u.user_type, u.created_at, u.updated_at FROM users u JOIN user_apartments ua ON u.id = ua.user_id WHERE ua.apartment_id = \$1`).
			WithArgs(apartmentID).
			WillReturnRows(rows)

		residents, err := repo.GetResidentsInApartment(apartmentID)
		assert.NoError(t, err)
		assert.Len(t, residents, 2)
		if len(residents) >= 2 {
			assert.Equal(t, "user1", residents[0].Username)
			assert.Equal(t, "user1@example.com", residents[0].Email)
			assert.Equal(t, "user2", residents[1].Username)
			assert.Equal(t, "user2@example.com", residents[1].Email)
		} else {
			t.Errorf("Expected 2 residents, got %d", len(residents))
		}
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("no residents found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "phone", "full_name", "user_type", "created_at", "updated_at",
		})

		mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.phone, u.full_name, u.user_type, u.created_at, u.updated_at FROM users u JOIN user_apartments ua ON u.id = ua.user_id WHERE ua.apartment_id = \$1`).
			WithArgs(apartmentID).
			WillReturnRows(rows)

		residents, err := repo.GetResidentsInApartment(apartmentID)
		assert.NoError(t, err)
		assert.Empty(t, residents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.phone, u.full_name, u.user_type, u.created_at, u.updated_at FROM users u JOIN user_apartments ua ON u.id = ua.user_id WHERE ua.apartment_id = \$1`).
			WithArgs(apartmentID).
			WillReturnError(sql.ErrConnDone)

		residents, err := repo.GetResidentsInApartment(apartmentID)
		assert.Error(t, err)
		assert.Nil(t, residents)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetAllApartmentsForAResident(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := &userApartmentRepositoryImpl{db: db}
	residentID := 1
	now := time.Now()

	t.Run("successful retrieval with apartments", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "apartment_name", "address", "units_count", "manager_id", "created_at", "updated_at",
		}).
			AddRow(1, "Darya Apartments", "123 Shariati St", 50, 1, now, now).
			AddRow(2, "Garden View", "456 Moz Ave", 30, 2, now, now)

		mock.ExpectQuery("SELECT a.id, a.apartment_name, a.address, a.units_count, a.manager_id, a.created_at, a.updated_at").
			WithArgs(residentID).
			WillReturnRows(rows)

		apartments, err := repo.GetAllApartmentsForAResident(residentID)
		assert.NoError(t, err)
		assert.Len(t, apartments, 2)
		assert.Equal(t, "Darya Apartments", apartments[0].ApartmentName)
		assert.Equal(t, "123 Shariati St", apartments[0].Address)
		assert.Equal(t, 50, apartments[0].UnitsCount)
		assert.Equal(t, "Garden View", apartments[1].ApartmentName)
		assert.Equal(t, "456 Moz Ave", apartments[1].Address)
		assert.Equal(t, 30, apartments[1].UnitsCount)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no apartments found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "apartment_name", "address", "units_count", "manager_id", "created_at", "updated_at",
		})

		mock.ExpectQuery("SELECT a.id, a.apartment_name, a.address, a.units_count, a.manager_id, a.created_at, a.updated_at").
			WithArgs(residentID).
			WillReturnRows(rows)

		apartments, err := repo.GetAllApartmentsForAResident(residentID)
		assert.NoError(t, err)
		assert.Empty(t, apartments)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT a.id, a.apartment_name, a.address, a.units_count, a.manager_id, a.created_at, a.updated_at").
			WithArgs(residentID).
			WillReturnError(sql.ErrConnDone)

		apartments, err := repo.GetAllApartmentsForAResident(residentID)
		assert.Error(t, err)
		assert.Nil(t, apartments)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
