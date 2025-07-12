package app

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/config"
)

func ConnectToDatabase(postgresConfig config.Postgres) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		postgresConfig.Host, postgresConfig.Port, postgresConfig.Username,
		postgresConfig.Password, postgresConfig.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %v", err)
	}
	fmt.Println("connected to Postgres")

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	return db, nil
}
