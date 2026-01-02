package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

// CacheConfig はRedis/Valkey接続設定を保持する
type CacheConfig struct {
	Host           string        `env:"CACHE_HOST" envDefault:"localhost"`
	Port           int           `env:"CACHE_PORT" envDefault:"6379"`
	Password       string        `env:"CACHE_PASSWORD"`
	Database       int           `env:"CACHE_DATABASE" envDefault:"0"`
	MaxRetries     int           `env:"CACHE_MAX_RETRIES" envDefault:"3"`
	ConnectTimeout time.Duration `env:"CACHE_CONNECT_TIMEOUT" envDefault:"5s"`
	ReadTimeout    time.Duration `env:"CACHE_READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout   time.Duration `env:"CACHE_WRITE_TIMEOUT" envDefault:"5s"`
	PoolSize       int           `env:"CACHE_POOL_SIZE" envDefault:"10"`
	MinIdleConns   int           `env:"CACHE_MIN_IDLE_CONNS" envDefault:"5"`
	TLSEnabled     bool          `env:"CACHE_TLS_ENABLED" envDefault:"false"`
}

// LoadCacheConfig はCache設定を読み込む
func LoadCacheConfig() (*CacheConfig, error) {
	cfg := &CacheConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
