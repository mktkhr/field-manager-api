package cache

import (
	"testing"
	"time"

	"github.com/mktkhr/field-manager-api/internal/config"
)

func TestNewRedisClient(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.CacheConfig
	}{
		{
			name: "デフォルト設定",
			cfg: config.CacheConfig{
				Host:           "localhost",
				Port:           6379,
				Password:       "",
				Database:       0,
				MaxRetries:     3,
				ConnectTimeout: 5 * time.Second,
				ReadTimeout:    5 * time.Second,
				WriteTimeout:   5 * time.Second,
				PoolSize:       10,
				MinIdleConns:   5,
				TLSEnabled:     false,
			},
		},
		{
			name: "カスタム設定",
			cfg: config.CacheConfig{
				Host:           "redis.example.com",
				Port:           6380,
				Password:       "secret",
				Database:       1,
				MaxRetries:     5,
				ConnectTimeout: 10 * time.Second,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				PoolSize:       20,
				MinIdleConns:   10,
				TLSEnabled:     false,
			},
		},
		{
			name: "TLS有効",
			cfg: config.CacheConfig{
				Host:           "secure-redis.example.com",
				Port:           6379,
				Password:       "password",
				Database:       0,
				MaxRetries:     3,
				ConnectTimeout: 5 * time.Second,
				ReadTimeout:    5 * time.Second,
				WriteTimeout:   5 * time.Second,
				PoolSize:       10,
				MinIdleConns:   5,
				TLSEnabled:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewRedisClient(tt.cfg)
			if client == nil {
				t.Error("NewRedisClient() should not return nil")
			}

			opts := client.Options()

			if opts.Password != tt.cfg.Password {
				t.Errorf("Password = %v, want %v", opts.Password, tt.cfg.Password)
			}
			if opts.DB != tt.cfg.Database {
				t.Errorf("DB = %v, want %v", opts.DB, tt.cfg.Database)
			}
			if opts.MaxRetries != tt.cfg.MaxRetries {
				t.Errorf("MaxRetries = %v, want %v", opts.MaxRetries, tt.cfg.MaxRetries)
			}
			if opts.DialTimeout != tt.cfg.ConnectTimeout {
				t.Errorf("DialTimeout = %v, want %v", opts.DialTimeout, tt.cfg.ConnectTimeout)
			}
			if opts.ReadTimeout != tt.cfg.ReadTimeout {
				t.Errorf("ReadTimeout = %v, want %v", opts.ReadTimeout, tt.cfg.ReadTimeout)
			}
			if opts.WriteTimeout != tt.cfg.WriteTimeout {
				t.Errorf("WriteTimeout = %v, want %v", opts.WriteTimeout, tt.cfg.WriteTimeout)
			}
			if opts.PoolSize != tt.cfg.PoolSize {
				t.Errorf("PoolSize = %v, want %v", opts.PoolSize, tt.cfg.PoolSize)
			}
			if opts.MinIdleConns != tt.cfg.MinIdleConns {
				t.Errorf("MinIdleConns = %v, want %v", opts.MinIdleConns, tt.cfg.MinIdleConns)
			}
			if tt.cfg.TLSEnabled && opts.TLSConfig == nil {
				t.Error("TLSConfig should not be nil when TLSEnabled is true")
			}
			if !tt.cfg.TLSEnabled && opts.TLSConfig != nil {
				t.Error("TLSConfig should be nil when TLSEnabled is false")
			}

			// クリーンアップ
			client.Close()
		})
	}
}

func TestNewCacheStore(t *testing.T) {
	cfg := config.CacheConfig{
		Host:           "localhost",
		Port:           6379,
		Password:       "",
		Database:       0,
		MaxRetries:     3,
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		PoolSize:       10,
		MinIdleConns:   5,
		TLSEnabled:     false,
	}

	store := NewCacheStore(cfg)
	if store == nil {
		t.Error("NewCacheStore() should not return nil")
	}

	// 型アサーションでClientであることを確認
	client, ok := store.(*Client)
	if !ok {
		t.Error("NewCacheStore() should return *Client")
	}
	if client.client == nil {
		t.Error("NewCacheStore() should initialize internal redis client")
	}

	// クリーンアップ
	store.Close()
}
