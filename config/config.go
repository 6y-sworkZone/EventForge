package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port       int
	DBPath     string
	JWTSecret  string
	SMTPHost   string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
	SMTPFrom   string
	UploadPath string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "18080"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		Port:       port,
		DBPath:     getEnv("DB_PATH", "eventforge.db"),
		JWTSecret:  getEnv("JWT_SECRET", "eventforge-secret-key-change-in-production"),
		SMTPHost:   getEnv("SMTP_HOST", ""),
		SMTPPort:   smtpPort,
		SMTPUser:   getEnv("SMTP_USER", ""),
		SMTPPass:   getEnv("SMTP_PASS", ""),
		SMTPFrom:   getEnv("SMTP_FROM", "noreply@eventforge.local"),
		UploadPath: getEnv("UPLOAD_PATH", "./uploads"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
