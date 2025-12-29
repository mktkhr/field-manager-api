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
				t.Fatal("NewRedisClient()はnilを返すべきではない")
			}

			opts := client.Options()

			if opts.Password != tt.cfg.Password {
				t.Errorf("Password = %v, 期待値 %v", opts.Password, tt.cfg.Password)
			}
			if opts.DB != tt.cfg.Database {
				t.Errorf("DB = %v, 期待値 %v", opts.DB, tt.cfg.Database)
			}
			if opts.MaxRetries != tt.cfg.MaxRetries {
				t.Errorf("MaxRetries = %v, 期待値 %v", opts.MaxRetries, tt.cfg.MaxRetries)
			}
			if opts.DialTimeout != tt.cfg.ConnectTimeout {
				t.Errorf("DialTimeout = %v, 期待値 %v", opts.DialTimeout, tt.cfg.ConnectTimeout)
			}
			if opts.ReadTimeout != tt.cfg.ReadTimeout {
				t.Errorf("ReadTimeout = %v, 期待値 %v", opts.ReadTimeout, tt.cfg.ReadTimeout)
			}
			if opts.WriteTimeout != tt.cfg.WriteTimeout {
				t.Errorf("WriteTimeout = %v, 期待値 %v", opts.WriteTimeout, tt.cfg.WriteTimeout)
			}
			if opts.PoolSize != tt.cfg.PoolSize {
				t.Errorf("PoolSize = %v, 期待値 %v", opts.PoolSize, tt.cfg.PoolSize)
			}
			if opts.MinIdleConns != tt.cfg.MinIdleConns {
				t.Errorf("MinIdleConns = %v, 期待値 %v", opts.MinIdleConns, tt.cfg.MinIdleConns)
			}
			if tt.cfg.TLSEnabled && opts.TLSConfig == nil {
				t.Error("TLSEnabledがtrueの場合、TLSConfigはnilであってはならない")
			}
			if !tt.cfg.TLSEnabled && opts.TLSConfig != nil {
				t.Error("TLSEnabledがfalseの場合、TLSConfigはnilであるべき")
			}

			// クリーンアップ
			if err := client.Close(); err != nil {
				t.Errorf("Close()でエラー発生 = %v", err)
			}
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
		t.Error("NewCacheStore()はnilを返すべきではない")
	}

	// 型アサーションでClientであることを確認
	client, ok := store.(*Client)
	if !ok {
		t.Error("NewCacheStore()は*Clientを返すべき")
	}
	if client.client == nil {
		t.Error("NewCacheStore()は内部のRedisクライアントを初期化すべき")
	}

	// クリーンアップ
	if err := store.Close(); err != nil {
		t.Errorf("Close()でエラー発生 = %v", err)
	}
}
