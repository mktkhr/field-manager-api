package config

// DatabaseConfig はデータベース接続設定を保持する
type DatabaseConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"postgres"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME" envDefault:"field_manager_db"`
	SSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`
}
