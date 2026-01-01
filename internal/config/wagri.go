package config

import "github.com/caarlos0/env/v10"

// WagriConfig はWagri API関連の設定
type WagriConfig struct {
	BaseURL      string `env:"WAGRI_BASE_URL" envDefault:"https://api.wagri.net"`
	ClientID     string `env:"WAGRI_CLIENT_ID"`
	ClientSecret string `env:"WAGRI_CLIENT_SECRET"`
}

// LoadWagriConfig はWagri設定を読み込む
func LoadWagriConfig() (*WagriConfig, error) {
	cfg := &WagriConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
