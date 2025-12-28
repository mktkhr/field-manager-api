package config

// LoggerConfig はログ出力の設定を保持する
type LoggerConfig struct {
	// Level はログレベルを指定する(DEBUG, INFO, WARN, ERROR)
	Level string `env:"LOG_LEVEL" envDefault:"INFO"`
	// Environment は実行環境を指定する(development, production)
	// developmentの場合は色付きテキスト形式、productionの場合はJSON形式で出力する
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

// IsProduction は本番環境かどうかを判定する
func (c *LoggerConfig) IsProduction() bool {
	return c.Environment == "production"
}
