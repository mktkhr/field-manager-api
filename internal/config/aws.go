package config

import (
	"github.com/caarlos0/env/v10"
)

// AWSConfig はAWS関連の設定
type AWSConfig struct {
	Region            string `env:"AWS_REGION" envDefault:"ap-northeast-1"`
	StepFunctionsARN  string `env:"AWS_STEP_FUNCTIONS_ARN"`
	S3Bucket          string `env:"AWS_S3_BUCKET" envDefault:"field-manager-imports"`
	LocalStackEnabled bool   `env:"LOCALSTACK_ENABLED" envDefault:"false"`
	LocalStackURL     string `env:"LOCALSTACK_URL" envDefault:"http://localhost:4566"`
}

// LoadAWSConfig はAWS設定を読み込む
func LoadAWSConfig() (*AWSConfig, error) {
	cfg := &AWSConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
