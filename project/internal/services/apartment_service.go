package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

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
	inviteLinkRepo      repositories.InviteLinkRepo
	notificationService notification.Notification
}

func NewApartmentService(
	apartmentRepo repositories.ApartmentRepository,
	userRepo repositories.UserRepository,
	userApartmentRepo repositories.UserApartmentRepository,
	inviteLinkRepo repositories.InviteLinkRepo,
	notificationService notification.Notification,
) ApartmentService {
	return &apartmentServiceImpl{
		apartmentRepo:       apartmentRepo,
		userRepo:            userRepo,
		userApartmentRepo:   userApartmentRepo,
		inviteLinkRepo:      inviteLinkRepo,
		notificationService: notificationService,
	}
}

func (s *apartmentServiceImpl) CreateApartment(ctx context.Context, userID int, apartmentName, address string, unitsCount int) (int, error) {
	logrus.Infof("Creating apartment '%s' for user %d", apartmentName, userID)

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
		logrus.WithError(err).Error("Failed to create apartment")
		return 0, fmt.Errorf("failed to create apartment: %w", err)
	}

	logrus.Infof("Apartment created with ID %d", id)

	userApartment := models.User_apartment{
		UserID:      userID,
		ApartmentID: id,
		IsManager:   true,
	}

	if err := s.userApartmentRepo.CreateUserApartment(ctx, userApartment); err != nil {
		logrus.WithError(err).Error("Failed to assign manager to apartment")
		return 0, fmt.Errorf("failed to assign manager to apartment: %w", err)
	}

	logrus.Infof("User %d assigned as manager for apartment %d", userID, id)
	return id, nil
}

func (s *apartmentServiceImpl) GetApartmentByID(ctx context.Context, id int) (*models.Apartment, error) {
	logrus.Infof("Fetching apartment by ID %d", id)
	apartment, err := s.apartmentRepo.GetApartmentByID(id)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to fetch apartment %d", id)
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}
	return apartment, nil
}

func (s *apartmentServiceImpl) GetResidentsInApartment(ctx context.Context, apartmentID int) ([]models.User, error) {
	logrus.Infof("Fetching residents for apartment %d", apartmentID)
	residents, err := s.userApartmentRepo.GetResidentsInApartment(apartmentID)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to get residents for apartment %d", apartmentID)
		return nil, fmt.Errorf("failed to get residents: %w", err)
	}
	return residents, nil
}

func (s *apartmentServiceImpl) GetAllApartmentsForResident(ctx context.Context, residentID int) ([]models.Apartment, error) {
	logrus.Infof("Fetching all apartments for resident %d", residentID)
	apartments, err := s.userApartmentRepo.GetAllApartmentsForAResident(residentID)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to get apartments for resident %d", residentID)
		return nil, fmt.Errorf("failed to get apartments: %w", err)
	}
	return apartments, nil
}

func (s *apartmentServiceImpl) UpdateApartment(ctx context.Context, id int, apartmentName, address string, unitsCount, managerID int) error {
	logrus.Infof("Updating apartment %d by manager %d", id, managerID)

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
		logrus.WithError(err).Errorf("Failed to update apartment %d", id)
		return fmt.Errorf("failed to update apartment: %w", err)
	}
	return nil
}

func (s *apartmentServiceImpl) DeleteApartment(ctx context.Context, id int) error {
	logrus.Infof("Deleting apartment %d", id)

	if err := s.apartmentRepo.DeleteApartment(id); err != nil {
		logrus.WithError(err).Errorf("Failed to delete apartment %d", id)
		return fmt.Errorf("failed to delete apartment: %w", err)
	}
	return nil
}

func (s *apartmentServiceImpl) InviteUserToApartment(ctx context.Context, managerID, apartmentID int, telegramUsername string) (map[string]interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"managerID":        managerID,
		"apartmentID":      apartmentID,
		"telegramUsername": telegramUsername,
	}).Info("Inviting user to apartment")

	isManager, err := s.userApartmentRepo.IsUserManagerOfApartment(ctx, managerID, apartmentID)
	if err != nil {
		logrus.WithError(err).Error("Failed to verify manager status")
		return nil, fmt.Errorf("failed to verify apartment manager status: %w", err)
	}
	if !isManager {
		logrus.Warn("Non-manager attempted to invite user")
		return nil, fmt.Errorf("only apartment managers can send invitations")
	}

	receiver, err := s.userRepo.GetUserByTelegramUser(telegramUsername)
	if err != nil {
		logrus.WithError(err).Error("Receiver Telegram user not found")
		return nil, fmt.Errorf("user with this Telegram username not found: %w", err)
	}

	isResident, err := s.userApartmentRepo.IsUserInApartment(ctx, receiver.ID, apartmentID)
	if err != nil {
		if err.Error() != "not in apartment" {
			logrus.WithError(err).Error("Failed to check if user is resident")
			return nil, fmt.Errorf("failed to check resident status: %w", err)
		}
	}
	if isResident {
		logrus.Warn("User is already a resident")
		return nil, fmt.Errorf("user is already a resident of this apartment")
	}

	generatedCode, err := s.inviteLinkRepo.CreateInvitation(ctx, receiver.ID, apartmentID, managerID)
	if err != nil {
		logrus.WithError(err).Error("Failed to create invitation")
		return nil, errors.New("failed to created invitation")
	}

	fmt.Println("Generated invitation code:", generatedCode)

	inviteURL := fmt.Sprintf("http://localhost:8080/api/v1/resident/apartment/invite/%s", generatedCode)

	fmt.Println("Invite URL: ", inviteURL)

	err = s.notificationService.SendInvitation(ctx, inviteURL, apartmentID, telegramUsername)
	if err != nil {
		logrus.WithError(err).Error("Failed to send Telegram invitation")
		return nil, errors.New("invitation created but failed to send notification")
	}

	logrus.Infof("Invitation sent successfully to %s", telegramUsername)

	return map[string]interface{}{
		"status":     "invitation sent",
		"expires_at": time.Now().Add(24 * time.Hour),
	}, nil
}

func (s *apartmentServiceImpl) JoinApartment(ctx context.Context, userID int, invitationCode string) (map[string]interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"userID":         userID,
		"invitationCode": invitationCode,
	}).Info("User attempting to join apartment")

	apartmentID, err := s.inviteLinkRepo.ValidateAndConsumeInvitation(ctx, invitationCode)
	if err != nil {
		logrus.WithError(err).Error("Invitation validation failed")
		return nil, err
	}

	isResident, err := s.userApartmentRepo.IsUserInApartment(ctx, userID, apartmentID)
	if err != nil {
		if err.Error() != "not in apartment" {
			logrus.WithError(err).Error("Failed to check if user is resident")
			return nil, fmt.Errorf("failed to check resident status: %w", err)
		}
	}
	if isResident {
		logrus.Warn("User is already a resident of this apartment")
		return nil, fmt.Errorf("you are already a resident of this apartment")
	}

	userApartment := models.User_apartment{
		UserID:      userID,
		ApartmentID: apartmentID,
		IsManager:   false,
	}

	if err := s.userApartmentRepo.CreateUserApartment(ctx, userApartment); err != nil {
		logrus.WithError(err).Error("Failed to join apartment")
		return nil, fmt.Errorf("failed to join apartment: %w", err)
	}

	logrus.Infof("User %d joined apartment %d", userID, apartmentID)
	return map[string]interface{}{
		"status": "joined apartment",
	}, nil
}

func (s *apartmentServiceImpl) LeaveApartment(ctx context.Context, userID, apartmentID int) error {
	logrus.Infof("User %d is leaving apartment %d", userID, apartmentID)

	if err := s.userApartmentRepo.DeleteUserApartment(userID, apartmentID); err != nil {
		logrus.WithError(err).Error("Failed to leave apartment")
		return fmt.Errorf("failed to leave apartment: %w", err)
	}
	return nil
}
