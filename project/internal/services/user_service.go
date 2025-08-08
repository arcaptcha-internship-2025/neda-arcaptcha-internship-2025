package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest, botAddress string) (*dto.SignUpResponse, error)
	AuthenticateUser(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	GetUserProfile(ctx context.Context, userID int) (*dto.ProfileResponse, error)
	UpdateUserProfile(ctx context.Context, userID int, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error)
	GetPublicUser(ctx context.Context, userID int) (*dto.PublicUserResponse, error)
	GetAllPublicUsers(ctx context.Context) ([]dto.PublicUserResponse, error)
	DeleteUser(ctx context.Context, userID int) error
}

type userServiceImpl struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
	}
}

func (s *userServiceImpl) CreateUser(ctx context.Context, req dto.CreateUserRequest, botAddress string) (*dto.SignUpResponse, error) {
	logger := logrus.WithFields(logrus.Fields{
		"username":      req.Username,
		"user_type":     req.UserType,
		"telegram_user": req.TelegramUser,
		"email":         req.Email,
	})

	logger.Info("Starting user creation")

	if req.UserType != models.Manager && req.UserType != models.Resident {
		logger.WithField("provided_type", req.UserType).Error("Invalid user type provided")
		return nil, fmt.Errorf("invalid user type")
	}

	if req.TelegramUser != "" && !isValidTelegramUsername(req.TelegramUser) {
		logger.WithField("telegram_username", req.TelegramUser).Error("Invalid Telegram username format")
		return nil, fmt.Errorf("invalid Telegram username format")
	}

	existingUser, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil && err != sql.ErrNoRows {
		logger.WithError(err).Error("Failed to check existing username")
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		logger.Warn("Attempt to create user with existing username")
		return nil, fmt.Errorf("username already exists")
	}

	if req.TelegramUser != "" {
		existingTelegramUser, err := s.userRepo.GetUserByTelegramUser(req.TelegramUser)
		if err != nil && err != sql.ErrNoRows {
			logger.WithError(err).Error("Failed to check existing Telegram username")
			return nil, fmt.Errorf("failed to check Telegram username: %w", err)
		}
		if existingTelegramUser != nil {
			logger.WithField("telegram_username", req.TelegramUser).Warn("Attempt to create user with existing Telegram username")
			return nil, fmt.Errorf("telegram username already in use")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.WithError(err).Error("Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		Username:     req.Username,
		Password:     string(hashedPassword),
		Email:        req.Email,
		Phone:        req.Phone,
		FullName:     req.FullName,
		UserType:     req.UserType,
		TelegramUser: req.TelegramUser,
	}

	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		logger.WithError(err).Error("Failed to create user in database")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.WithField("user_id", userID).Info("User created successfully")

	response := &dto.SignUpResponse{
		User: dto.UserInfo{
			ID:           userID,
			Username:     user.Username,
			Email:        user.Email,
			Phone:        user.Phone,
			FullName:     user.FullName,
			UserType:     user.UserType,
			TelegramUser: user.TelegramUser,
		},
		TelegramSetupRequired:     req.TelegramUser != "",
		TelegramSetupInstructions: "",
	}

	//add bot address hereeeeee
	if req.TelegramUser != "" {
		response.TelegramSetupInstructions = "Please start a chat with our bot in Telegram to complete setup : " + botAddress
		logger.WithField("bot_address", botAddress).Debug("Telegram setup instructions provided")
	}

	return response, nil
}

func (s *userServiceImpl) AuthenticateUser(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	logger := logrus.WithField("username", req.Username)
	logger.Info("Authentication attempt")

	existingUser, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Authentication failed - user not found")
			return nil, fmt.Errorf("invalid username or password")
		}
		logger.WithError(err).Error("Failed to retrieve user during authentication")
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		logger.WithField("user_id", existingUser.ID).Warn("Authentication failed - invalid password")
		return nil, fmt.Errorf("invalid username or password")
	}

	token, err := middleware.GenerateToken(strconv.Itoa(existingUser.ID), existingUser.UserType)
	if err != nil {
		logger.WithError(err).WithField("user_id", existingUser.ID).Error("Failed to generate authentication token")
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"user_id":   existingUser.ID,
		"user_type": existingUser.UserType,
	}).Info("Authentication successful")

	response := &dto.LoginResponse{
		Token:    "Bearer " + token,
		UserID:   strconv.Itoa(existingUser.ID),
		UserType: string(existingUser.UserType),
		Username: existingUser.Username,
		Email:    existingUser.Email,
		FullName: existingUser.FullName,
		Telegram: dto.TelegramInfo{
			Username:  existingUser.TelegramUser,
			Connected: existingUser.TelegramChatID != 0,
		},
	}

	return response, nil
}

func (s *userServiceImpl) GetUserProfile(ctx context.Context, userID int) (*dto.ProfileResponse, error) {
	logger := logrus.WithField("user_id", userID)
	logger.Debug("Retrieving user profile")

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		logger.WithError(err).Error("User profile not found")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	response := &dto.ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		FullName: user.FullName,
		UserType: user.UserType,
		Telegram: dto.TelegramInfo{
			Username:  user.TelegramUser,
			Connected: user.TelegramChatID != 0,
		},
	}

	logger.Debug("User profile retrieved successfully")
	return response, nil
}

func (s *userServiceImpl) UpdateUserProfile(ctx context.Context, userID int, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	logger := logrus.WithFields(logrus.Fields{
		"user_id":         userID,
		"update_username": req.Username != "",
		"update_email":    req.Email != "",
		"update_telegram": req.TelegramUser != "",
	})

	logger.Info("Starting user profile update")

	existingUser, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		logger.WithError(err).Error("User not found for profile update")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if req.TelegramUser != "" && req.TelegramUser != existingUser.TelegramUser {
		if !isValidTelegramUsername(req.TelegramUser) {
			logger.WithField("telegram_username", req.TelegramUser).Error("Invalid Telegram username format in update")
			return nil, fmt.Errorf("invalid Telegram username format")
		}

		existingTelegramUser, err := s.userRepo.GetUserByTelegramUser(req.TelegramUser)
		if err != nil && err != sql.ErrNoRows {
			logger.WithError(err).Error("Failed to check Telegram username during update")
			return nil, fmt.Errorf("failed to check Telegram username: %w", err)
		}
		if existingTelegramUser != nil && existingTelegramUser.ID != userID {
			logger.WithField("telegram_username", req.TelegramUser).Warn("Attempt to update to existing Telegram username")
			return nil, fmt.Errorf("telegram username already in use")
		}
	}

	originalTelegramUser := existingUser.TelegramUser

	if req.Username != "" {
		existingUser.Username = req.Username
	}
	if req.Email != "" {
		existingUser.Email = req.Email
	}
	if req.Phone != "" {
		existingUser.Phone = req.Phone
	}
	if req.FullName != "" {
		existingUser.FullName = req.FullName
	}
	if req.TelegramUser != "" {
		existingUser.TelegramUser = req.TelegramUser
		//reset chat id if Telegram username is changed
		if req.TelegramUser != originalTelegramUser {
			existingUser.TelegramChatID = 0
			logger.WithField("new_telegram_username", req.TelegramUser).Info("Telegram username changed, chat ID reset")
		}
	}

	if err := s.userRepo.UpdateUser(ctx, *existingUser); err != nil {
		logger.WithError(err).Error("Failed to update user profile in database")
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	logger.Info("User profile updated successfully")

	response := &dto.ProfileResponse{
		ID:       existingUser.ID,
		Username: existingUser.Username,
		Email:    existingUser.Email,
		Phone:    existingUser.Phone,
		FullName: existingUser.FullName,
		UserType: existingUser.UserType,
		Telegram: dto.TelegramInfo{
			Username:  existingUser.TelegramUser,
			Connected: existingUser.TelegramChatID != 0,
		},
	}

	return response, nil
}

func (s *userServiceImpl) GetPublicUser(ctx context.Context, userID int) (*dto.PublicUserResponse, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Public user not found")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	response := &dto.PublicUserResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
	}

	return response, nil
}

func (s *userServiceImpl) GetAllPublicUsers(ctx context.Context) ([]dto.PublicUserResponse, error) {
	logrus.Debug("Retrieving all public users")

	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to retrieve all users")
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}

	publicUsers := make([]dto.PublicUserResponse, len(users))
	for i, user := range users {
		publicUsers[i] = dto.PublicUserResponse{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
			UserType: user.UserType,
		}
	}

	logrus.WithField("users_count", len(users)).Debug("All public users retrieved successfully")
	return publicUsers, nil
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, userID int) error {
	logger := logrus.WithField("user_id", userID)
	logger.Info("Starting user deletion")

	if err := s.userRepo.DeleteUser(userID); err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Attempt to delete non-existent user")
			return fmt.Errorf("user not found")
		}
		logger.WithError(err).Error("Failed to delete user from database")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("User deleted successfully")
	return nil
}

func isValidTelegramUsername(username string) bool {
	if len(username) < 5 || len(username) > 32 {
		return false
	}

	//telegram usernames can only contain a-z, 0-9, and underscores
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}
