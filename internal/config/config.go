package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config はアプリケーション全体の設定を保持する
type Config struct {
	Logger   LoggerConfig
	Database DatabaseConfig
	Cache    CacheConfig
	Storage  StorageConfig
}

// Load は環境変数から設定を読み込む
// .envファイルが存在する場合は先に読み込む
func Load() (*Config, error) {
	// 開発環境: .envファイルがあれば読み込み
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
