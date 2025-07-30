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
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.SignUpResponse, error)
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

func (s *userServiceImpl) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.SignUpResponse, error) {
	if req.UserType != models.Manager && req.UserType != models.Resident {
		return nil, fmt.Errorf("invalid user type")
	}

	if req.TelegramUser != "" && !isValidTelegramUsername(req.TelegramUser) {
		return nil, fmt.Errorf("invalid Telegram username format")
	}

	existingUser, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	if req.TelegramUser != "" {
		existingTelegramUser, err := s.userRepo.GetUserByTelegramUser(req.TelegramUser)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check Telegram username: %w", err)
		}
		if existingTelegramUser != nil {
			return nil, fmt.Errorf("Telegram username already in use")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
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
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := middleware.GenerateToken(strconv.Itoa(userID), user.UserType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

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
		Token:                     token,
		TelegramSetupRequired:     req.TelegramUser != "",
		TelegramSetupInstructions: "",
	}

	//add bot address hereeeeee
	if req.TelegramUser != "" {
		response.TelegramSetupInstructions = "Please start a chat with our bot in Telegram to complete setup : "
	}

	return response, nil
}

func (s *userServiceImpl) AuthenticateUser(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	existingUser, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	token, err := middleware.GenerateToken(strconv.Itoa(existingUser.ID), existingUser.UserType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	response := &dto.LoginResponse{
		Token:    token,
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
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
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

	return response, nil
}

func (s *userServiceImpl) UpdateUserProfile(ctx context.Context, userID int, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	existingUser, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if req.TelegramUser != "" && req.TelegramUser != existingUser.TelegramUser {
		if !isValidTelegramUsername(req.TelegramUser) {
			return nil, fmt.Errorf("invalid Telegram username format")
		}

		existingTelegramUser, err := s.userRepo.GetUserByTelegramUser(req.TelegramUser)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check Telegram username: %w", err)
		}
		if existingTelegramUser != nil && existingTelegramUser.ID != userID {
			return nil, fmt.Errorf("Telegram username already in use")
		}
	}

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
		if req.TelegramUser != existingUser.TelegramUser {
			existingUser.TelegramChatID = 0
		}
	}

	if err := s.userRepo.UpdateUser(ctx, *existingUser); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

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
	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
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

	return publicUsers, nil
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, userID int) error {
	if err := s.userRepo.DeleteUser(userID); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
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
