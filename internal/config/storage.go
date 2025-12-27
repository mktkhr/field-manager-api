package config

import "time"

// StorageConfig はS3/RustFS接続設定を保持する
type StorageConfig struct {
	// S3を有効にするか(true: AWS S3、false: Minio/RustFS)
	S3Enabled bool `env:"STORAGE_S3_ENABLED" envDefault:"false"`

	// S3互換エンドポイント(バックエンド用)
	// AWS S3を使用する場合は空を指定
	Endpoint string `env:"STORAGE_ENDPOINT" envDefault:"http://localhost:9000"`

	// ブラウザからアクセス可能なエンドポイント(presigned URL用)
	// 未設定の場合はEndpointと同じ値を使用
	PublicEndpoint string `env:"STORAGE_PUBLIC_ENDPOINT"`

	Region          string `env:"STORAGE_REGION" envDefault:"ap-northeast-1"`
	Bucket          string `env:"STORAGE_BUCKET" envDefault:"pts-soa-bucket"`
	AccessKeyID     string `env:"STORAGE_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"STORAGE_SECRET_ACCESS_KEY"`

	// RustFS/MinIOの場合はtrue(パス形式URL)
	UsePathStyle bool `env:"STORAGE_USE_PATH_STYLE" envDefault:"true"`

	// Pre-signed URLの有効期限
	PresignedURLExpiry time.Duration `env:"STORAGE_PRESIGNED_URL_EXPIRY" envDefault:"900s"`
}

// GetPublicEndpoint はPublicEndpointが未設定の場合Endpointを返す
func (s *StorageConfig) GetPublicEndpoint() string {
	if s.PublicEndpoint == "" {
		return s.Endpoint
	}
	return s.PublicEndpoint
}
