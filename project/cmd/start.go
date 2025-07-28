package cmd

import (
	"log"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/app"
	myhttp "github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/payment"
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

	redisClient := app.NewRedisClient(cfg.Redis)

	userRepo := repositories.NewUserRepository(cfg.Postgres.AutoCreate, db)
	apartmentRepo := repositories.NewApartmentRepository(cfg.Postgres.AutoCreate, db)
	userApartmentRepo := repositories.NewUserApartmentRepository(cfg.Postgres.AutoCreate, db)
	inviteLinkRepo := repositories.NewInvitationLinkRepository(redisClient)
	billRepo := repositories.NewBillRepository(cfg.Postgres.AutoCreate, db)
	paymentRepo := repositories.NewPaymentRepository(cfg.Postgres.AutoCreate, db)

	notificationService := notification.NewNotification(
		cfg.TelegramConfig,
		cfg.Server.AppBaseURL,
		userRepo,
	)

	imageService := image.NewImage(cfg.Minio.Endpoint, cfg.Minio.AccessKey, cfg.Minio.SecretKey, cfg.Minio.Bucket)
	paymentService := payment.NewPayment()
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
		paymentRepo,
		paymentService,
	)

	if err := httpService.Start("Apartment Service"); err != nil {
		log.Fatalf("failed to start apartment service: %v", err)
	}
	httpService.WaitForShutdown()
}
