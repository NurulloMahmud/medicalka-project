package config

import (
	"os"
	"strconv"
)

type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Limiter struct {
	RPS      float64
	Burst    int
	Enabbled bool
}

type Config struct {
	Env         string
	ServerAddr  string
	DatabaseURL string
	JWTSecret   string
	Limiter     Limiter
	SMTP        SMTP
}

func Load() *Config {
	rps, _ := strconv.ParseFloat(getEnv("LIMITER_RPS", "2"), 64)
	burst, _ := strconv.Atoi(getEnv("LIMITER_BURST", "4"))
	enabled, _ := strconv.ParseBool(getEnv("LIMITER_ENABLED", "true"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	appLimiter := Limiter{
		RPS:      rps,
		Burst:    burst,
		Enabbled: enabled,
	}

	smtpConfig := SMTP{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     smtpPort,
		Username: getEnv("SMTP_USERNAME", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("SMTP_FROM", "nurullomahmud@gmail.com"),
	}

	return &Config{
		Env:         getEnv("ENV", "development"),
		ServerAddr:  getEnv("SERVER_ADDRESS", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/medicalka?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "9b36f2a2f8a1482690a671d16ca14932"),
		Limiter:     appLimiter,
		SMTP:        smtpConfig,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}
