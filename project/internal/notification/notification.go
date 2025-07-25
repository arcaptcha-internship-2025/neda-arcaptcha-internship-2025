package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

// for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Notification interface {
	SendNotification(ctx context.Context, userID int, message string) error
	SendInvitation(ctx context.Context, inv models.InvitationLink) error
	HandleStartCommand(ctx context.Context, telegramUser string, chatID int64) error
	HandleWebhookUpdate(ctx context.Context, update Update) error
	SendBillNotification(ctx context.Context, userID int, bill models.Bill, amount float64) error
}

type notificationImpl struct {
	httpClient HTTPClient
	botToken   string
	baseURL    string
	appBaseURL string
	userRepo   repositories.UserRepository
}

// Update represents a telegram bot update from getUpdates or webhook
type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

// telegram message
type Message struct {
	MessageID int    `json:"message_id"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

// telegram user
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// telegram chat
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func NewNotification(cfg config.TelegramConfig, appBaseURL string, userRepo repositories.UserRepository) Notification {
	return &notificationImpl{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		botToken:   cfg.BotToken,
		baseURL:    fmt.Sprintf("https://api.telegram.org/bot%s/", cfg.BotToken),
		appBaseURL: appBaseURL,
		userRepo:   userRepo,
	}
}

func (n *notificationImpl) HandleWebhookUpdate(ctx context.Context, update Update) error {
	// Handle /start command
	if strings.HasPrefix(update.Message.Text, "/start") {
		return n.HandleStartCommand(
			ctx,
			update.Message.From.Username,
			update.Message.Chat.ID,
		)
	}

	// todo: add handling for other commands
	return nil
}

func (n *notificationImpl) HandleStartCommand(ctx context.Context, telegramUser string, chatID int64) error {
	user, err := n.userRepo.GetUserByTelegramUser(telegramUser)
	if err != nil {
		return fmt.Errorf("user not found with Telegram username: %s", telegramUser)
	}

	if err := n.userRepo.UpdateTelegramChatID(ctx, user.ID, chatID); err != nil {
		return fmt.Errorf("failed to update Telegram chat ID: %w", err)
	}

	welcomeMsg := fmt.Sprintf(
		"Hello %s! You've successfully connected your account.\n\n"+
			"You'll now receive notifications from our service here.",
		user.FullName,
	)
	return n.sendMessage(ctx, chatID, welcomeMsg)
}

func (n *notificationImpl) sendMessage(ctx context.Context, chatID int64, message string) error {
	endpoint := n.baseURL + "sendMessage"
	data := url.Values{}
	data.Set("chat_id", strconv.FormatInt(chatID, 10))
	data.Set("text", message)
	data.Set("parse_mode", "Markdown") //md formatting

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

func (n *notificationImpl) SendNotification(ctx context.Context, userID int, message string) error {
	user, err := n.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.TelegramChatID == 0 {
		return fmt.Errorf("user hasn't started the bot yet")
	}

	return n.sendMessage(ctx, user.TelegramChatID, message)
}

func (n *notificationImpl) SendInvitation(ctx context.Context, inv models.InvitationLink) error {
	receiver, err := n.userRepo.GetUserByTelegramUser(inv.ReceiverUsername)
	if err != nil {
		return fmt.Errorf("failed to get receiver user: %w", err)
	}

	if receiver.TelegramChatID == 0 {
		return fmt.Errorf("receiver hasn't started the bot yet")
	}

	inviteURL := fmt.Sprintf("%s/join?token=%s", n.appBaseURL, inv.Token)
	message := fmt.Sprintf(
		"🏠 *New Apartment Invitation*\n\n"+
			"You've been invited to join apartment *%d*!\n\n"+
			"🔗 [Accept Invitation](%s)\n\n"+
			"⏰ Expires: %s",
		inv.ApartmentID,
		inviteURL,
		inv.ExpiresAt.Format("2006-01-02 15:04:05"),
	)

	return n.sendMessage(ctx, receiver.TelegramChatID, message)
}

func (n *notificationImpl) SendBillNotification(ctx context.Context, userID int, bill models.Bill, amount float64) error {
	user, err := n.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.TelegramChatID == 0 {
		return fmt.Errorf("user hasn't started the bot yet")
	}

	message := fmt.Sprintf(
		" *New Bill Notification*\n\n"+
			"Type: %s\n"+
			"Your Share: %.2f\n"+
			"Due Date: %s\n"+
			"Description: %s\n\n",
		bill.BillType, amount, bill.DueDate, bill.Description)

	if bill.ImageURL != "" {
		message += "Bill image is available in your dashboard"
	}

	return n.sendMessage(ctx, user.TelegramChatID, message)
}
