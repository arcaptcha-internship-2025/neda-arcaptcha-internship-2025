package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_INVITATION_LINKS_TABLE = `CREATE TABLE IF NOT EXISTS invitation_links(
		id SERIAL PRIMARY KEY,
		sender_id INTEGER NOT NULL REFERENCES users(id)
		reciever_id INTEGER NOT NULL REFERENCES users(id),
        apartment_id INTEGER NOT NULL REFERENCES apartments(id),
        unit_id INTEGER NOT NULL REFERENCES units(id),
        token VARCHAR(255) UNIQUE NOT NULL,
        expires_at TIMESTAMP NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`
)

type InvitationLinkRepository interface {
	CreateInvitationLink(apartmentID, unitID, createdBy int, token string, expiresAt string) (int, error)
	GetInvitationLinkByID(id int) (*models.InvitationLink, error)
	GetInvitationLinkByToken(token string) (*models.InvitationLink, error)
	MarkLinkAsUsed(id int, usedBy int, usedAt string) error
}

type invitationLinkRepositoryImpl struct {
	db *sqlx.DB
}

func NewInvitationLinkRepository(autoCreate bool, db *sqlx.DB) (InvitationLinkRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_INVITATION_LINKS_TABLE); err != nil {
			return nil, err
		}
	}
	return &invitationLinkRepositoryImpl{db: db}, nil
}

func (r *invitationLinkRepositoryImpl) CreateInvitationLink(apartmentID, unitID, createdBy int, token string, expiresAt string) (int, error) {
	var id int
	query := `INSERT INTO invitation_links (apartment_id, unit_id, token, created_by, expires_at) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, apartmentID, unitID, token, createdBy, expiresAt).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *invitationLinkRepositoryImpl) GetInvitationLinkByID(id int) (*models.InvitationLink, error) {
	var invitationLink models.InvitationLink
	query := `SELECT id, apartment_id, unit_id, token, created_by, expires_at, is_used, used_by, used_at, created_at, updated_at 
              FROM invitation_links WHERE id = $1`
	err := r.db.Get(&invitationLink, query, id)
	if err != nil {
		return nil, err
	}
	return &invitationLink, nil
}

func (r *invitationLinkRepositoryImpl) GetInvitationLinkByToken(token string) (*models.InvitationLink, error) {
	var invitationLink models.InvitationLink
	query := `SELECT id, apartment_id, unit_id, token, created_by, expires_at, is_used, used_by, used_at, created_at, updated_at 
              FROM invitation_links WHERE token = $1`
	err := r.db.Get(&invitationLink, query, token)
	if err != nil {
		return nil, err
	}
	return &invitationLink, nil
}

func (r *invitationLinkRepositoryImpl) MarkLinkAsUsed(id int, usedBy int, usedAt string) error {
	query := `UPDATE invitation_links SET is_used = TRUE, used_by = $1, used_at = $2, updated_at = CURRENT_TIMESTAMP 
              WHERE id = $3`
	_, err := r.db.Exec(query, usedBy, usedAt, id)
	return err
}
