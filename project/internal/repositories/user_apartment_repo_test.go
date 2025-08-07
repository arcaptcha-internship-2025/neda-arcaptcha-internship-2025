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

func TestUserApartmentRepository_CreateUserApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	userApartment := models.User_apartment{
		UserID:      1,
		ApartmentID: 2,
		IsManager:   true,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO user_apartments`).
			WithArgs(userApartment.UserID, userApartment.ApartmentID, userApartment.IsManager).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateUserApartment(context.Background(), userApartment)
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO user_apartments`).
			WithArgs(userApartment.UserID, userApartment.ApartmentID, userApartment.IsManager).
			WillReturnError(sql.ErrConnDone)

		err := repo.CreateUserApartment(context.Background(), userApartment)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserApartmentRepository_GetUserApartmentByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	userID := 1
	apartmentID := 2
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"user_id", "apartment_id", "is_manager", "created_at", "updated_at"}).
			AddRow(userID, apartmentID, true, now, now)

		mock.ExpectQuery(`SELECT user_id, apartment_id, is_manager, created_at, updated_at FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnRows(rows)

		result, err := repo.GetUserApartmentByID(userID, apartmentID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, apartmentID, result.ApartmentID)
		assert.True(t, result.IsManager)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT user_id, apartment_id, is_manager, created_at, updated_at FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetUserApartmentByID(userID, apartmentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserApartmentRepository_UpdateUserApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	userApartment := models.User_apartment{
		UserID:      1,
		ApartmentID: 2,
		IsManager:   false,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE user_apartments SET is_manager`).
			WithArgs(userApartment.IsManager, userApartment.UserID, userApartment.ApartmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateUserApartment(context.Background(), userApartment)
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(`UPDATE user_apartments SET is_manager`).
			WithArgs(userApartment.IsManager, userApartment.UserID, userApartment.ApartmentID).
			WillReturnError(sql.ErrConnDone)

		err := repo.UpdateUserApartment(context.Background(), userApartment)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserApartmentRepository_DeleteUserApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	userID := 1
	apartmentID := 2

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteUserApartment(userID, apartmentID)
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrConnDone)

		err := repo.DeleteUserApartment(userID, apartmentID)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserApartmentRepository_GetResidentsInApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	apartmentID := 2
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "email", "phone", "full_name", "user_type", "created_at", "updated_at"}).
			AddRow(1, "user1", "user1@example.com", "1234567890", "User One", "resident", now, now).
			AddRow(2, "user2", "user2@example.com", "0987654321", "User Two", "resident", now, now)

		mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.phone, u.full_name, u.user_type, u.created_at, u.updated_at FROM users u JOIN user_apartments ua`).
			WithArgs(apartmentID).
			WillReturnRows(rows)

		residents, err := repo.GetResidentsInApartment(apartmentID)
		assert.NoError(t, err)
		assert.Len(t, residents, 2)
		assert.Equal(t, "user1", residents[0].Username)
		assert.Equal(t, "user2", residents[1].Username)
	})

	t.Run("no residents", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "email", "phone", "full_name", "user_type", "created_at", "updated_at"})

		mock.ExpectQuery(`SELECT u.id, u.username, u.email, u.phone, u.full_name, u.user_type, u.created_at, u.updated_at FROM users u JOIN user_apartments ua`).
			WithArgs(apartmentID).
			WillReturnRows(rows)

		residents, err := repo.GetResidentsInApartment(apartmentID)
		assert.NoError(t, err)
		assert.Len(t, residents, 0)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserApartmentRepository_GetAllApartmentsForAResident(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	residentID := 1
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "apartment_name", "address", "units_count", "manager_id", "created_at", "updated_at"}).
			AddRow(1, "Apartment A", "123 Main St", 10, 1, now, now).
			AddRow(2, "Apartment B", "456 Oak Ave", 15, 2, now, now)

		mock.ExpectQuery(`SELECT a.id, a.apartment_name, a.address, a.units_count, a.manager_id, a.created_at, a.updated_at FROM apartments a JOIN user_apartments ua`).
			WithArgs(residentID).
			WillReturnRows(rows)

		apartments, err := repo.GetAllApartmentsForAResident(residentID)
		assert.NoError(t, err)
		assert.Len(t, apartments, 2)
		assert.Equal(t, "Apartment A", apartments[0].ApartmentName)
		assert.Equal(t, "Apartment B", apartments[1].ApartmentName)
	})

	t.Run("no apartments", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "apartment_name", "address", "units_count", "manager_id", "created_at", "updated_at"})

		mock.ExpectQuery(`SELECT a.id, a.apartment_name, a.address, a.units_count, a.manager_id, a.created_at, a.updated_at FROM apartments a JOIN user_apartments ua`).
			WithArgs(residentID).
			WillReturnRows(rows)

		apartments, err := repo.GetAllApartmentsForAResident(residentID)
		assert.NoError(t, err)
		assert.Len(t, apartments, 0)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserApartmentRepository_IsUserManagerOfApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	userID := 1
	apartmentID := 2

	t.Run("is manager", func(t *testing.T) {
		mock.ExpectQuery(`SELECT is_manager FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnRows(sqlmock.NewRows([]string{"is_manager"}).AddRow(true))

		isManager, err := repo.IsUserManagerOfApartment(context.Background(), userID, apartmentID)
		assert.NoError(t, err)
		assert.True(t, isManager)
	})

	t.Run("not manager but exists", func(t *testing.T) {
		mock.ExpectQuery(`SELECT is_manager FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnRows(sqlmock.NewRows([]string{"is_manager"}).AddRow(false))

		isManager, err := repo.IsUserManagerOfApartment(context.Background(), userID, apartmentID)
		assert.Error(t, err)
		assert.False(t, isManager)
		assert.Contains(t, err.Error(), "not manager")
	})

	t.Run("user not in apartment", func(t *testing.T) {
		mock.ExpectQuery(`SELECT is_manager FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrNoRows)

		isManager, err := repo.IsUserManagerOfApartment(context.Background(), userID, apartmentID)
		assert.Error(t, err)
		assert.False(t, isManager)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT is_manager FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrConnDone)

		isManager, err := repo.IsUserManagerOfApartment(context.Background(), userID, apartmentID)
		assert.Error(t, err)
		assert.False(t, isManager)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserApartmentRepository_IsUserInApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	userID := 1
	apartmentID := 2

	t.Run("user in apartment", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(userID, apartmentID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.IsUserInApartment(context.Background(), userID, apartmentID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("user not in apartment", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(userID, apartmentID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		exists, err := repo.IsUserInApartment(context.Background(), userID, apartmentID)
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "not in apartment")
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrConnDone)

		exists, err := repo.IsUserInApartment(context.Background(), userID, apartmentID)
		assert.Error(t, err)
		assert.False(t, exists)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
