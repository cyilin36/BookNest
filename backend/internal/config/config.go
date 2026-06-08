package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                    string
	HTTPAddr                  string
	PublicBaseURL             string
	DatabaseDSN               string
	DataDir                   string
	BooksDir                  string
	CoversDir                 string
	TempDir                   string
	FrontendDistDir           string
	JWTSecret                 string
	AccessTokenTTL            time.Duration
	RefreshTokenTTL           time.Duration
	MaxUploadSizeMB           int
	RequestBodyLimitMB        int
	TXTChunkSize              int
	AllowRegistration         bool
	LibraryReviewRequired     bool
	DefaultUserStorageQuotaMB int
	LogLevel                  string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:        getenv("APP_ENV", "development"),
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		PublicBaseURL: getenv("PUBLIC_BASE_URL", ""),
		DatabaseDSN:   getenv("DATABASE_DSN", "postgres://booknest:password@localhost:5432/booknest?sslmode=disable"),
		DataDir:       getenv("DATA_DIR", "/data"),
		JWTSecret:     getenv("JWT_SECRET", "change-me"),
		LogLevel:      getenv("LOG_LEVEL", "info"),
	}
	cfg.BooksDir = getenv("BOOKS_DIR", filepath.Join(cfg.DataDir, "books"))
	cfg.CoversDir = getenv("COVERS_DIR", filepath.Join(cfg.DataDir, "covers"))
	cfg.TempDir = getenv("TEMP_DIR", filepath.Join(cfg.DataDir, "temp"))
	cfg.FrontendDistDir = getenv("FRONTEND_DIST_DIR", "./frontend/dist")
	cfg.AccessTokenTTL = mustDuration(getenv("ACCESS_TOKEN_TTL", "2h"), 2*time.Hour)
	cfg.RefreshTokenTTL = mustDuration(getenv("REFRESH_TOKEN_TTL", "720h"), 30*24*time.Hour)
	cfg.MaxUploadSizeMB = mustInt(getenv("MAX_UPLOAD_SIZE_MB", "100"), 100)
	cfg.RequestBodyLimitMB = mustInt(getenv("REQUEST_BODY_LIMIT_MB", "110"), 110)
	cfg.TXTChunkSize = mustInt(getenv("TXT_CHUNK_SIZE", "65536"), 65536)
	cfg.AllowRegistration = mustBool(getenv("ALLOW_REGISTRATION", "true"), true)
	cfg.LibraryReviewRequired = mustBool(getenv("LIBRARY_REVIEW_REQUIRED", "false"), false)
	cfg.DefaultUserStorageQuotaMB = mustInt(getenv("DEFAULT_USER_STORAGE_QUOTA_MB", "10240"), 10240)
	if cfg.RequestBodyLimitMB < cfg.MaxUploadSizeMB {
		return nil, fmt.Errorf("REQUEST_BODY_LIMIT_MB must be >= MAX_UPLOAD_SIZE_MB")
	}
	if cfg.AppEnv == "production" && cfg.JWTSecret == "change-me" {
		return nil, fmt.Errorf("JWT_SECRET must be changed in production")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func mustInt(v string, fallback int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
		return n
	}
	return fallback
}

func mustBool(v string, fallback bool) bool {
	if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
		return b
	}
	return fallback
}

func mustDuration(v string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
		return d
	}
	return fallback
}
