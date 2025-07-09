package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   Server   `yaml:"server"`
	Postgres Postgres `yaml:"postgres"`
}

type Server struct {
	Port string `yaml:"port"`
}

type Postgres struct {
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Database   string `yaml:"database"`
	AutoCreate bool   `yaml:"autocreate"`
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
