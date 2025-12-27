package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// 必須の環境変数を設定
	t.Setenv("DB_PASSWORD", "testpassword")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Database defaults
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "localhost")
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want %d", cfg.Database.Port, 5432)
	}
	if cfg.Database.User != "postgres" {
		t.Errorf("Database.User = %q, want %q", cfg.Database.User, "postgres")
	}
	if cfg.Database.Password != "testpassword" {
		t.Errorf("Database.Password = %q, want %q", cfg.Database.Password, "testpassword")
	}
	if cfg.Database.Name != "field_manager_db" {
		t.Errorf("Database.Name = %q, want %q", cfg.Database.Name, "field_manager_db")
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("Database.SSLMode = %q, want %q", cfg.Database.SSLMode, "disable")
	}

	// Cache defaults
	if cfg.Cache.Host != "localhost" {
		t.Errorf("Cache.Host = %q, want %q", cfg.Cache.Host, "localhost")
	}
	if cfg.Cache.Port != 6379 {
		t.Errorf("Cache.Port = %d, want %d", cfg.Cache.Port, 6379)
	}
	if cfg.Cache.Database != 0 {
		t.Errorf("Cache.Database = %d, want %d", cfg.Cache.Database, 0)
	}
	if cfg.Cache.MaxRetries != 3 {
		t.Errorf("Cache.MaxRetries = %d, want %d", cfg.Cache.MaxRetries, 3)
	}
	if cfg.Cache.ConnectTimeout != 5*time.Second {
		t.Errorf("Cache.ConnectTimeout = %v, want %v", cfg.Cache.ConnectTimeout, 5*time.Second)
	}
	if cfg.Cache.ReadTimeout != 5*time.Second {
		t.Errorf("Cache.ReadTimeout = %v, want %v", cfg.Cache.ReadTimeout, 5*time.Second)
	}
	if cfg.Cache.WriteTimeout != 5*time.Second {
		t.Errorf("Cache.WriteTimeout = %v, want %v", cfg.Cache.WriteTimeout, 5*time.Second)
	}
	if cfg.Cache.PoolSize != 10 {
		t.Errorf("Cache.PoolSize = %d, want %d", cfg.Cache.PoolSize, 10)
	}
	if cfg.Cache.MinIdleConns != 5 {
		t.Errorf("Cache.MinIdleConns = %d, want %d", cfg.Cache.MinIdleConns, 5)
	}
	if cfg.Cache.TLSEnabled != false {
		t.Errorf("Cache.TLSEnabled = %v, want %v", cfg.Cache.TLSEnabled, false)
	}

	// Storage defaults
	if cfg.Storage.S3Enabled != false {
		t.Errorf("Storage.S3Enabled = %v, want %v", cfg.Storage.S3Enabled, false)
	}
	if cfg.Storage.Endpoint != "http://localhost:9000" {
		t.Errorf("Storage.Endpoint = %q, want %q", cfg.Storage.Endpoint, "http://localhost:9000")
	}
	if cfg.Storage.Region != "ap-northeast-1" {
		t.Errorf("Storage.Region = %q, want %q", cfg.Storage.Region, "ap-northeast-1")
	}
	if cfg.Storage.Bucket != "pts-soa-bucket" {
		t.Errorf("Storage.Bucket = %q, want %q", cfg.Storage.Bucket, "pts-soa-bucket")
	}
	if cfg.Storage.UsePathStyle != true {
		t.Errorf("Storage.UsePathStyle = %v, want %v", cfg.Storage.UsePathStyle, true)
	}
	if cfg.Storage.PresignedURLExpiry != 900*time.Second {
		t.Errorf("Storage.PresignedURLExpiry = %v, want %v", cfg.Storage.PresignedURLExpiry, 900*time.Second)
	}
}

func TestLoadWithCustomValues(t *testing.T) {
	// カスタム値を設定
	t.Setenv("DB_HOST", "custom-db-host")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "custom-user")
	t.Setenv("DB_PASSWORD", "custom-password")
	t.Setenv("DB_NAME", "custom-db")
	t.Setenv("DB_SSL_MODE", "require")

	t.Setenv("CACHE_HOST", "custom-cache-host")
	t.Setenv("CACHE_PORT", "6380")
	t.Setenv("CACHE_PASSWORD", "cache-password")
	t.Setenv("CACHE_DATABASE", "1")
	t.Setenv("CACHE_MAX_RETRIES", "5")
	t.Setenv("CACHE_CONNECT_TIMEOUT", "10s")
	t.Setenv("CACHE_READ_TIMEOUT", "10s")
	t.Setenv("CACHE_WRITE_TIMEOUT", "10s")
	t.Setenv("CACHE_POOL_SIZE", "20")
	t.Setenv("CACHE_MIN_IDLE_CONNS", "10")
	t.Setenv("CACHE_TLS_ENABLED", "true")

	t.Setenv("STORAGE_S3_ENABLED", "true")
	t.Setenv("STORAGE_ENDPOINT", "https://s3.amazonaws.com")
	t.Setenv("STORAGE_PUBLIC_ENDPOINT", "https://cdn.example.com")
	t.Setenv("STORAGE_REGION", "us-east-1")
	t.Setenv("STORAGE_BUCKET", "my-bucket")
	t.Setenv("STORAGE_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	t.Setenv("STORAGE_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	t.Setenv("STORAGE_USE_PATH_STYLE", "false")
	t.Setenv("STORAGE_PRESIGNED_URL_EXPIRY", "1800s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Database
	if cfg.Database.Host != "custom-db-host" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "custom-db-host")
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("Database.Port = %d, want %d", cfg.Database.Port, 5433)
	}
	if cfg.Database.User != "custom-user" {
		t.Errorf("Database.User = %q, want %q", cfg.Database.User, "custom-user")
	}
	if cfg.Database.Password != "custom-password" {
		t.Errorf("Database.Password = %q, want %q", cfg.Database.Password, "custom-password")
	}
	if cfg.Database.Name != "custom-db" {
		t.Errorf("Database.Name = %q, want %q", cfg.Database.Name, "custom-db")
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("Database.SSLMode = %q, want %q", cfg.Database.SSLMode, "require")
	}

	// Cache
	if cfg.Cache.Host != "custom-cache-host" {
		t.Errorf("Cache.Host = %q, want %q", cfg.Cache.Host, "custom-cache-host")
	}
	if cfg.Cache.Port != 6380 {
		t.Errorf("Cache.Port = %d, want %d", cfg.Cache.Port, 6380)
	}
	if cfg.Cache.Password != "cache-password" {
		t.Errorf("Cache.Password = %q, want %q", cfg.Cache.Password, "cache-password")
	}
	if cfg.Cache.Database != 1 {
		t.Errorf("Cache.Database = %d, want %d", cfg.Cache.Database, 1)
	}
	if cfg.Cache.MaxRetries != 5 {
		t.Errorf("Cache.MaxRetries = %d, want %d", cfg.Cache.MaxRetries, 5)
	}
	if cfg.Cache.ConnectTimeout != 10*time.Second {
		t.Errorf("Cache.ConnectTimeout = %v, want %v", cfg.Cache.ConnectTimeout, 10*time.Second)
	}
	if cfg.Cache.ReadTimeout != 10*time.Second {
		t.Errorf("Cache.ReadTimeout = %v, want %v", cfg.Cache.ReadTimeout, 10*time.Second)
	}
	if cfg.Cache.WriteTimeout != 10*time.Second {
		t.Errorf("Cache.WriteTimeout = %v, want %v", cfg.Cache.WriteTimeout, 10*time.Second)
	}
	if cfg.Cache.PoolSize != 20 {
		t.Errorf("Cache.PoolSize = %d, want %d", cfg.Cache.PoolSize, 20)
	}
	if cfg.Cache.MinIdleConns != 10 {
		t.Errorf("Cache.MinIdleConns = %d, want %d", cfg.Cache.MinIdleConns, 10)
	}
	if cfg.Cache.TLSEnabled != true {
		t.Errorf("Cache.TLSEnabled = %v, want %v", cfg.Cache.TLSEnabled, true)
	}

	// Storage
	if cfg.Storage.S3Enabled != true {
		t.Errorf("Storage.S3Enabled = %v, want %v", cfg.Storage.S3Enabled, true)
	}
	if cfg.Storage.Endpoint != "https://s3.amazonaws.com" {
		t.Errorf("Storage.Endpoint = %q, want %q", cfg.Storage.Endpoint, "https://s3.amazonaws.com")
	}
	if cfg.Storage.PublicEndpoint != "https://cdn.example.com" {
		t.Errorf("Storage.PublicEndpoint = %q, want %q", cfg.Storage.PublicEndpoint, "https://cdn.example.com")
	}
	if cfg.Storage.Region != "us-east-1" {
		t.Errorf("Storage.Region = %q, want %q", cfg.Storage.Region, "us-east-1")
	}
	if cfg.Storage.Bucket != "my-bucket" {
		t.Errorf("Storage.Bucket = %q, want %q", cfg.Storage.Bucket, "my-bucket")
	}
	if cfg.Storage.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("Storage.AccessKeyID = %q, want %q", cfg.Storage.AccessKeyID, "AKIAIOSFODNN7EXAMPLE")
	}
	if cfg.Storage.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("Storage.SecretAccessKey = %q, want %q", cfg.Storage.SecretAccessKey, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	}
	if cfg.Storage.UsePathStyle != false {
		t.Errorf("Storage.UsePathStyle = %v, want %v", cfg.Storage.UsePathStyle, false)
	}
	if cfg.Storage.PresignedURLExpiry != 1800*time.Second {
		t.Errorf("Storage.PresignedURLExpiry = %v, want %v", cfg.Storage.PresignedURLExpiry, 1800*time.Second)
	}
}

func TestLoadMissingRequiredEnv(t *testing.T) {
	// DB_PASSWORDを設定しない(required)

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error when required env is missing")
	}
}

func TestStorageConfigGetPublicEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       string
		publicEndpoint string
		want           string
	}{
		{
			name:           "PublicEndpointが設定されている場合",
			endpoint:       "http://localhost:9000",
			publicEndpoint: "https://cdn.example.com",
			want:           "https://cdn.example.com",
		},
		{
			name:           "PublicEndpointが未設定の場合",
			endpoint:       "http://localhost:9000",
			publicEndpoint: "",
			want:           "http://localhost:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StorageConfig{
				Endpoint:       tt.endpoint,
				PublicEndpoint: tt.publicEndpoint,
			}
			if got := s.GetPublicEndpoint(); got != tt.want {
				t.Errorf("GetPublicEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}
