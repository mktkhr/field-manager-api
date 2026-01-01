package config

import (
	"testing"
)

func TestLoadAWSConfig(t *testing.T) {
	// デフォルト値のテスト
	t.Setenv("AWS_REGION", "ap-northeast-1")
	t.Setenv("AWS_STEP_FUNCTIONS_ARN", "")
	t.Setenv("AWS_S3_BUCKET", "field-manager-imports")
	t.Setenv("LOCALSTACK_ENABLED", "false")
	t.Setenv("LOCALSTACK_URL", "http://localhost:4566")

	cfg, err := LoadAWSConfig()
	if err != nil {
		t.Fatalf("LoadAWSConfig()でエラー発生 = %v", err)
	}

	if cfg.Region != "ap-northeast-1" {
		t.Errorf("Region = %q, 期待値 %q", cfg.Region, "ap-northeast-1")
	}
	if cfg.S3Bucket != "field-manager-imports" {
		t.Errorf("S3Bucket = %q, 期待値 %q", cfg.S3Bucket, "field-manager-imports")
	}
	if cfg.LocalStackEnabled != false {
		t.Errorf("LocalStackEnabled = %v, 期待値 %v", cfg.LocalStackEnabled, false)
	}
	if cfg.LocalStackURL != "http://localhost:4566" {
		t.Errorf("LocalStackURL = %q, 期待値 %q", cfg.LocalStackURL, "http://localhost:4566")
	}
}

func TestLoadAWSConfigWithCustomValues(t *testing.T) {
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_STEP_FUNCTIONS_ARN", "arn:aws:states:us-east-1:123456789012:stateMachine:test")
	t.Setenv("AWS_S3_BUCKET", "custom-bucket")
	t.Setenv("LOCALSTACK_ENABLED", "true")
	t.Setenv("LOCALSTACK_URL", "http://localstack:4566")

	cfg, err := LoadAWSConfig()
	if err != nil {
		t.Fatalf("LoadAWSConfig()でエラー発生 = %v", err)
	}

	if cfg.Region != "us-east-1" {
		t.Errorf("Region = %q, 期待値 %q", cfg.Region, "us-east-1")
	}
	if cfg.StepFunctionsARN != "arn:aws:states:us-east-1:123456789012:stateMachine:test" {
		t.Errorf("StepFunctionsARN = %q, 期待値 %q", cfg.StepFunctionsARN, "arn:aws:states:us-east-1:123456789012:stateMachine:test")
	}
	if cfg.S3Bucket != "custom-bucket" {
		t.Errorf("S3Bucket = %q, 期待値 %q", cfg.S3Bucket, "custom-bucket")
	}
	if cfg.LocalStackEnabled != true {
		t.Errorf("LocalStackEnabled = %v, 期待値 %v", cfg.LocalStackEnabled, true)
	}
	if cfg.LocalStackURL != "http://localstack:4566" {
		t.Errorf("LocalStackURL = %q, 期待値 %q", cfg.LocalStackURL, "http://localstack:4566")
	}
}

func TestLoadAWSConfigDefaults(t *testing.T) {
	// 環境変数をクリア(デフォルト値のテスト)
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_STEP_FUNCTIONS_ARN", "")
	t.Setenv("AWS_S3_BUCKET", "")
	t.Setenv("LOCALSTACK_ENABLED", "")
	t.Setenv("LOCALSTACK_URL", "")

	cfg, err := LoadAWSConfig()
	if err != nil {
		t.Fatalf("LoadAWSConfig()でエラー発生 = %v", err)
	}

	// デフォルト値を確認
	if cfg.Region != "ap-northeast-1" {
		t.Errorf("Region = %q, want %q (デフォルト)", cfg.Region, "ap-northeast-1")
	}
	if cfg.S3Bucket != "field-manager-imports" {
		t.Errorf("S3Bucket = %q, want %q (デフォルト)", cfg.S3Bucket, "field-manager-imports")
	}
	if cfg.LocalStackEnabled != false {
		t.Errorf("LocalStackEnabled = %v, want %v (デフォルト)", cfg.LocalStackEnabled, false)
	}
	if cfg.LocalStackURL != "http://localhost:4566" {
		t.Errorf("LocalStackURL = %q, want %q (デフォルト)", cfg.LocalStackURL, "http://localhost:4566")
	}
}
