package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type AppConfig struct {
	Host  string
	Port  string
	Debug bool
	Env   string
}

type DatabaseConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	Name           string
	SSLMode        string
	MaxConnections int
	MigrationsPath string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type AuthConfig struct {
	JWTSecret     string
	JWTExpiration int
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Warn("could not read .env file, using environment variables")
		}
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetDefault("APP_HOST", "0.0.0.0")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_DEBUG", false)
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_CONNECTIONS", 25)
	v.SetDefault("DB_MIGRATIONS_PATH", "file://migrations")
	v.SetDefault("AUTH_JWT_EXPIRATION", 24)

	cfg := &Config{
		App: AppConfig{
			Host:  v.GetString("APP_HOST"),
			Port:  v.GetString("APP_PORT"),
			Debug: v.GetBool("APP_DEBUG"),
			Env:   v.GetString("APP_ENV"),
		},
		Database: DatabaseConfig{
			Host:           v.GetString("DB_HOST"),
			Port:           v.GetString("DB_PORT"),
			User:           v.GetString("DB_USER"),
			Password:       v.GetString("DB_PASSWORD"),
			Name:           v.GetString("DB_NAME"),
			SSLMode:        v.GetString("DB_SSLMODE"),
			MaxConnections: v.GetInt("DB_MAX_CONNECTIONS"),
			MigrationsPath: v.GetString("DB_MIGRATIONS_PATH"),
		},
		Auth: AuthConfig{
			JWTSecret:     v.GetString("AUTH_JWT_SECRET"),
			JWTExpiration: v.GetInt("AUTH_JWT_EXPIRATION"),
		},
	}

	if cfg.Database.Host == "" || cfg.Auth.JWTSecret == "" {
		return nil, fmt.Errorf("DB_HOST and AUTH_JWT_SECRET are required")
	}

	return cfg, nil
}
