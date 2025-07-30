package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

type ApartmentService interface {
	CreateApartment(ctx context.Context, userID int, apartmentName, address string, unitsCount int) (int, error)
	GetApartmentByID(ctx context.Context, id int) (*models.Apartment, error)
	GetResidentsInApartment(ctx context.Context, apartmentID int) ([]models.User, error)
	GetAllApartmentsForResident(ctx context.Context, residentID int) ([]models.Apartment, error)
	UpdateApartment(ctx context.Context, id int, apartmentName, address string, unitsCount, managerID int) error
	DeleteApartment(ctx context.Context, id int) error
	InviteUserToApartment(ctx context.Context, managerID, apartmentID int, telegramUsername string) (map[string]interface{}, error)
	JoinApartment(ctx context.Context, userID int, token string) (map[string]interface{}, error)
	LeaveApartment(ctx context.Context, userID, apartmentID int) error
}

type apartmentServiceImpl struct {
	apartmentRepo       repositories.ApartmentRepository
	userRepo            repositories.UserRepository
	userApartmentRepo   repositories.UserApartmentRepository
	inviteLinkRepo      repositories.InviteLinkFlagRepo
	notificationService notification.Notification
	appBaseURL          string
}

func NewApartmentService(
	apartmentRepo repositories.ApartmentRepository,
	userRepo repositories.UserRepository,
	userApartmentRepo repositories.UserApartmentRepository,
	inviteLinkRepo repositories.InviteLinkFlagRepo,
	notificationService notification.Notification,
	appBaseURL string,
) ApartmentService {
	return &apartmentServiceImpl{
		apartmentRepo:       apartmentRepo,
		userRepo:            userRepo,
		userApartmentRepo:   userApartmentRepo,
		inviteLinkRepo:      inviteLinkRepo,
		notificationService: notificationService,
		appBaseURL:          appBaseURL,
	}
}

func (s *apartmentServiceImpl) CreateApartment(ctx context.Context, userID int, apartmentName, address string, unitsCount int) (int, error) {
	apartment := models.Apartment{
		BaseModel: models.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ApartmentName: apartmentName,
		Address:       address,
		UnitsCount:    unitsCount,
		ManagerID:     userID,
	}

	id, err := s.apartmentRepo.CreateApartment(ctx, apartment)
	if err != nil {
		return 0, fmt.Errorf("failed to create apartment: %w", err)
	}

	userApartment := models.User_apartment{
		UserID:      userID,
		ApartmentID: id,
		IsManager:   true,
	}

	if err := s.userApartmentRepo.CreateUserApartment(ctx, userApartment); err != nil {
		return 0, fmt.Errorf("failed to assign manager to apartment: %w", err)
	}

	return id, nil
}

func (s *apartmentServiceImpl) GetApartmentByID(ctx context.Context, id int) (*models.Apartment, error) {
	apartment, err := s.apartmentRepo.GetApartmentByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}
	return apartment, nil
}

func (s *apartmentServiceImpl) GetResidentsInApartment(ctx context.Context, apartmentID int) ([]models.User, error) {
	residents, err := s.userApartmentRepo.GetResidentsInApartment(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get residents: %w", err)
	}
	return residents, nil
}

func (s *apartmentServiceImpl) GetAllApartmentsForResident(ctx context.Context, residentID int) ([]models.Apartment, error) {
	apartments, err := s.userApartmentRepo.GetAllApartmentsForAResident(residentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments: %w", err)
	}
	return apartments, nil
}

func (s *apartmentServiceImpl) UpdateApartment(ctx context.Context, id int, apartmentName, address string, unitsCount, managerID int) error {
	apartment := models.Apartment{
		BaseModel: models.BaseModel{
			ID:        id,
			UpdatedAt: time.Now(),
		},
		ApartmentName: apartmentName,
		Address:       address,
		UnitsCount:    unitsCount,
		ManagerID:     managerID,
	}

	if err := s.apartmentRepo.UpdateApartment(ctx, apartment); err != nil {
		return fmt.Errorf("failed to update apartment: %w", err)
	}
	return nil
}

func (s *apartmentServiceImpl) DeleteApartment(ctx context.Context, id int) error {
	if err := s.apartmentRepo.DeleteApartment(id); err != nil {
		return fmt.Errorf("failed to delete apartment: %w", err)
	}
	return nil
}

func (s *apartmentServiceImpl) InviteUserToApartment(ctx context.Context, managerID, apartmentID int, telegramUsername string) (map[string]interface{}, error) {
	isManager, err := s.userApartmentRepo.IsUserManagerOfApartment(ctx, managerID, apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify apartment manager status: %w", err)
	}
	if !isManager {
		return nil, fmt.Errorf("only apartment managers can send invitations")
	}

	receiver, err := s.userRepo.GetUserByTelegramUser(telegramUsername)
	if err != nil {
		return nil, fmt.Errorf("user with this Telegram username not found: %w", err)
	}

	isResident, err := s.userApartmentRepo.IsUserInApartment(ctx, receiver.ID, apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check resident status: %w", err)
	}
	if isResident {
		return nil, fmt.Errorf("user is already a resident of this apartment")
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	apartment, err := s.apartmentRepo.GetApartmentByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment details: %w", err)
	}

	manager, err := s.userRepo.GetUserByID(managerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manager details: %w", err)
	}

	inviteURL := fmt.Sprintf("%s/apartment/join?token=%s", s.appBaseURL, token)

	invitation := models.InvitationLink{
		SenderID:         managerID,
		SenderUsername:   manager.Username,
		ReceiverUsername: telegramUsername,
		ApartmentID:      apartmentID,
		ApartmentName:    apartment.ApartmentName,
		Token:            token,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Status:           models.InvitationStatusPending,
		InviteURL:        inviteURL,
	}

	if err := s.inviteLinkRepo.CreateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to store invitation: %w", err)
	}

	if err := s.notificationService.SendInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("invitation created but failed to send notification: %w", err)
	}

	return map[string]interface{}{
		"status":     "invitation sent",
		"token":      token,
		"invite_url": inviteURL,
		"expires_at": invitation.ExpiresAt,
	}, nil
}

func (s *apartmentServiceImpl) JoinApartment(ctx context.Context, userID int, token string) (map[string]interface{}, error) {
	inv, err := s.inviteLinkRepo.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	if inv.Status != models.InvitationStatusPending && inv.Status != models.InvitationStatusNotified {
		return nil, fmt.Errorf("invitation is no longer valid")
	}

	isResident, err := s.userApartmentRepo.IsUserInApartment(ctx, userID, inv.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check resident status: %w", err)
	}
	if isResident {
		return nil, fmt.Errorf("you are already a resident of this apartment")
	}

	userApartment := models.User_apartment{
		UserID:      userID,
		ApartmentID: inv.ApartmentID,
		IsManager:   false,
	}

	if err := s.userApartmentRepo.CreateUserApartment(ctx, userApartment); err != nil {
		return nil, fmt.Errorf("failed to join apartment: %w", err)
	}

	return map[string]interface{}{
		"status":       "joined apartment",
		"apartment_id": inv.ApartmentID,
		"apartment":    inv.ApartmentName,
	}, nil
}

func (s *apartmentServiceImpl) LeaveApartment(ctx context.Context, userID, apartmentID int) error {
	if err := s.userApartmentRepo.DeleteUserApartment(userID, apartmentID); err != nil {
		return fmt.Errorf("failed to leave apartment: %w", err)
	}
	return nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
