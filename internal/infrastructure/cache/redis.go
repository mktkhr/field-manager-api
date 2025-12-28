package cache

import (
	"crypto/tls"
	"fmt"

	"github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/features/shared/application/port"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient はCacheConfigからRedisクライアントを作成する
func NewRedisClient(cfg config.CacheConfig) *redis.Client {
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.ConnectTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	}

	if cfg.TLSEnabled {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return redis.NewClient(opts)
}

// NewCacheStore はCacheConfigからCacheStoreインターフェースを作成する
func NewCacheStore(cfg config.CacheConfig) port.CacheStore {
	redisClient := NewRedisClient(cfg)
	return NewClient(redisClient)
}
