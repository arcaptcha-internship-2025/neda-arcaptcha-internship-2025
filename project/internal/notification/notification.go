package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

type Notification interface {
	SendNotification(ctx context.Context, userID int, message string) error
	SendInvitation(ctx context.Context, inviteURL string, apartmentID int, receiverUsername string) error
	SendBillNotification(ctx context.Context, userID int, bill models.Bill, amount float64) error
	ListenForUpdates(ctx context.Context)
}

type notificationImpl struct {
	userRepo repositories.UserRepository
	bot      *tgbotapi.BotAPI
}

func NewNotification(cfg config.TelegramConfig, userRepo repositories.UserRepository) Notification {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	return &notificationImpl{
		userRepo: userRepo,
		bot:      bot,
	}
}

func (n *notificationImpl) sendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	_, err := n.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message via tgbot: %w", err)
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

	return n.sendMessage(user.TelegramChatID, message)
}

func (n *notificationImpl) SendInvitation(ctx context.Context, inviteURL string, apartmentID int, receiverUsername string) error {
	receiver, err := n.userRepo.GetUserByTelegramUser(receiverUsername)
	if err != nil {
		return fmt.Errorf("failed to get receiver user: %w", err)
	}

	if receiver.TelegramChatID == 0 {
		return fmt.Errorf("receiver hasn't started the bot yet")
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	message := fmt.Sprintf(
		"🏠 *New Apartment Invitation*\n\n"+
			"You've been invited to join apartment *%d*!\n\n"+
			"🔗 Accept Invitation: %s\n\n"+
			"⏰ Expires: %s",
		apartmentID,
		inviteURL,
		expiresAt.Format("2006-01-02 15:04:05"),
	)

	return n.sendMessage(receiver.TelegramChatID, message)
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
		"*New Bill Notification*\n\n"+
			"Type: %s\n"+
			"Your Share: %.2f\n"+
			"Due Date: %s\n"+
			"Description: %s\n",
		bill.BillType, amount, bill.DueDate, bill.Description)

	return n.sendMessage(user.TelegramChatID, message)
}

func (n *notificationImpl) ListenForUpdates(ctx context.Context) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := n.bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			go n.handleMessage(ctx, n.bot, update)
		}
	}
}

func (n *notificationImpl) handleMessage(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message.IsCommand() && update.Message.Command() == "start" {
		chatID := update.Message.Chat.ID
		username := update.SentFrom().UserName
		n.userRepo.UpdateTelegramChatID(ctx, username, chatID)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Welcome, %s! Bot is active.", username))
		bot.Send(msg)
		return
	}
}
