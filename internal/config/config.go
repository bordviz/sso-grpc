package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env                 string         `yaml:"env" env-required:"true"`
	TokenExpires        time.Duration  `yaml:"token_expires" env-required:"true"`
	RefreshTokenExpires time.Duration  `yaml:"refresh_token_expires" env-required:"true"`
	Database            DatabaseConfig `yaml:"database" env-required:"true"`
	GRPC                GRPCConfig     `yaml:"grpc" env-required:"true"`
	MigrationsPath      string         `yaml:"migrations_path" env-required:"true"`
}

type DatabaseConfig struct {
	Host     string        `yaml:"host" env-required:"true"`
	Port     int           `yaml:"port" env-required:"true"`
	User     string        `yaml:"user" env-required:"true"`
	Password string        `yaml:"password" env-required:"true"`
	Name     string        `yaml:"name" env-required:"true"`
	SSLMode  string        `yaml:"sslmode" env-required:"true"`
	Timeout  time.Duration `yaml:"timeout" env-required:"true"`
	Delay    time.Duration `yaml:"delay" env-required:"true"`
	Attempts int           `yaml:"attempts" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("failed to load environment file, error: ", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	cfg := MustLoadWithPath(configPath)
	return cfg
}

func MustLoadWithPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config file not found")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("failed to read config, error: ", err)
	}

	return &cfg
}
