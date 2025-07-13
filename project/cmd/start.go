package cmd

import (
	"log"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http"

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
	httpService := http.NewApartmantService(cfg)
	if err := httpService.Start("Apartment Service"); err != nil {
		log.Fatalf("failed to start apartment service: %v", err)
	}
	httpService.WaitForShutdown()
}
