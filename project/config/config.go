package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server         Server         `yaml:"server"`
	Postgres       Postgres       `yaml:"postgres"`
	Minio          Minio          `yaml:"minio"`
	Redis          Redis          `yaml:"redis"`
	TelegramConfig TelegramConfig `yaml:"telegram_config"`
}

type Server struct {
	Port       string `yaml:"port"`
	AppBaseURL string `yaml:"app_base_url" env:"APP_BASE_URL" env-default:"http://localhost:8080"`
}

type Postgres struct {
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Database   string `yaml:"database"`
	AutoCreate bool   `yaml:"autocreate"`
}

type Minio struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
}

type Redis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type TelegramConfig struct {
	BotToken   string        `yaml:"bot_token"`
	Timeout    time.Duration `yaml:"timeout"`
	BotAddress string        `yaml:"bot_address"`
}

func InitConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
