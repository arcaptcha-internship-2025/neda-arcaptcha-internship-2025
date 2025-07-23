package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/app"
	myhttp "github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/spf13/cobra"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the Apartment service",
		Run:   start,
	}
)

func start(_ *cobra.Command, _ []string) {
	cfg, err := config.InitConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to initialize config: %v", err)
	}

	db, err := app.ConnectToDatabase(cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	minioClient, err := app.ConnectToMinio(cfg.Minio)
	if err != nil {
		log.Fatalf("failed to connect to MinIO: %v", err)
	}

	//telegram webhook
	if err := setupTelegramWebhook(cfg.TelegramConfig); err != nil {
		log.Fatalf("failed to setup Telegram webhook: %v", err)
	}

	redisClient := app.NewRedisClient(cfg.Redis)

	userRepo := repositories.NewUserRepository(cfg.Postgres.AutoCreate, db)
	apartmentRepo := repositories.NewApartmentRepository(cfg.Postgres.AutoCreate, db)
	userApartmentRepo := repositories.NewUserApartmentRepository(cfg.Postgres.AutoCreate, db)
	inviteLinkRepo := repositories.NewInvitationLinkRepository(redisClient)
	billRepo := repositories.NewBillRepository(cfg.Postgres.AutoCreate, db)

	notificationService := notification.NewNotification(
		cfg.TelegramConfig,
		cfg.Server.AppBaseURL,
		userRepo,
	)

	imageService := image.NewImage(cfg.Minio.Endpoint, cfg.Minio.AccessKey, cfg.Minio.SecretKey, cfg.Minio.Bucket)

	httpService := myhttp.NewApartmantService(
		cfg,
		db,
		minioClient,
		redisClient,
		userRepo,
		apartmentRepo,
		userApartmentRepo,
		inviteLinkRepo,
		notificationService,
		billRepo,
		imageService,
	)

	if err := httpService.Start("Apartment Service"); err != nil {
		log.Fatalf("failed to start apartment service: %v", err)
	}
	httpService.WaitForShutdown()
}

func setupTelegramWebhook(cfg config.TelegramConfig) error {
	if cfg.WebhookURL == "" {
		return errors.New("telegram webhook URL not configured")
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", cfg.BotToken)
	data := url.Values{}
	data.Set("url", cfg.WebhookURL)
	if cfg.MaxConnections > 0 {
		data.Set("max_connections", strconv.Itoa(cfg.MaxConnections))
	}

	resp, err := http.PostForm(endpoint, data)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
