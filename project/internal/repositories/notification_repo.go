package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_NOTIFICATION_TABLE = `CREATE TABLE IF NOT EXIST notifications(
		id SERIAL PRIMARY KEY,
		sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		apartment_id INTEGER REFERENCES apartments(id) ON DELETE CASCADE,
		notification_type VARCHAR(50) NOT NULL, 
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		is_sent BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`
)

type NotificationRepository interface {
	CreateNotification(senderID, apartmentID int, notificationType, title, content string) (int, error)
	GetNotificationByID(id int) (*models.Notification, error)
	UpdateNotification(id int, title, content string) error
	DeleteNotification(id int) error
}

type notificationRepositoryImpl struct {
	db *sqlx.DB
}

func NewNotificationRepository(autoCreate bool, db *sqlx.DB) (NotificationRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_NOTIFICATION_TABLE); err != nil {
			return nil, err
		}
	}
	return &notificationRepositoryImpl{db: db}, nil
}

func (r *notificationRepositoryImpl) CreateNotification(senderID, apartmentID int, notificationType, title, content string) (int, error) {
	var id int
	query := `INSERT INTO notifications (sender_id, apartment_id, notification_type, title, content) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, senderID, apartmentID, notificationType, title, content).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *notificationRepositoryImpl) GetNotificationByID(id int) (*models.Notification, error) {
	var notification models.Notification
	query := `SELECT id, sender_id, apartment_id, notification_type, title, content, is_sent, created_at, updated_at 
			  FROM notifications WHERE id = $1`
	err := r.db.Get(&notification, query, id)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepositoryImpl) UpdateNotification(id int, title, content string) error {
	query := `UPDATE notifications SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = $3`
	_, err := r.db.Exec(query, title, content, id)
	return err
}

func (r *notificationRepositoryImpl) DeleteNotification(id int) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
