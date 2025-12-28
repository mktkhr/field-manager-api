package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v10"
	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/features/import/infrastructure/external"
)

// Event はStep Functionsからの入力イベント
type Event struct {
	CityCode    string    `json:"city_code"`
	ImportJobID uuid.UUID `json:"import_job_id"`
}

// Output はStep Functionsへの出力
type Output struct {
	S3Key       string    `json:"s3_key"`
	ImportJobID uuid.UUID `json:"import_job_id"`
	CityCode    string    `json:"city_code"`
}

// WagriConfig はwagri API設定
type WagriConfig struct {
	BaseURL string `env:"WAGRI_BASE_URL" envDefault:"https://api.wagri.net"`
	APIKey  string `env:"WAGRI_API_KEY"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	lambda.Start(handler)
}

func handler(ctx context.Context, event Event) (*Output, error) {
	slog.Info("wagri-fetcher開始", "city_code", event.CityCode, "import_job_id", event.ImportJobID)

	// AWS設定読み込み
	awsCfg, err := config.LoadAWSConfig()
	if err != nil {
		slog.Error("AWS設定の読み込みに失敗", "error", err)
		return nil, fmt.Errorf("AWS設定の読み込みに失敗: %w", err)
	}

	// wagri設定読み込み
	wagriCfg := &WagriConfig{}
	if err := env.Parse(wagriCfg); err != nil {
		slog.Error("wagri設定の読み込みに失敗", "error", err)
		return nil, fmt.Errorf("wagri設定の読み込みに失敗: %w", err)
	}

	// S3クライアント作成
	s3Client, err := external.NewS3Client(ctx, awsCfg)
	if err != nil {
		slog.Error("S3クライアントの作成に失敗", "error", err)
		return nil, fmt.Errorf("S3クライアントの作成に失敗: %w", err)
	}

	// wagriクライアント作成
	wagriClient := external.NewWagriClient(&external.WagriConfig{
		BaseURL: wagriCfg.BaseURL,
		APIKey:  wagriCfg.APIKey,
	})

	// wagri APIから圃場データを取得
	slog.Info("wagri API呼び出し開始", "city_code", event.CityCode)
	data, err := wagriClient.FetchFieldsByCityCodeToStream(ctx, event.CityCode)
	if err != nil {
		slog.Error("wagri API呼び出しに失敗", "error", err, "city_code", event.CityCode)
		return nil, fmt.Errorf("wagri API呼び出しに失敗: %w", err)
	}
	slog.Info("wagri API呼び出し完了", "size_bytes", len(data))

	// S3キー生成
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	s3Key := fmt.Sprintf("imports/%s/%s.json", event.CityCode, timestamp)

	// S3にアップロード
	slog.Info("S3アップロード開始", "s3_key", s3Key)
	if err := s3Client.Upload(ctx, s3Key, bytes.NewReader(data), "application/json"); err != nil {
		slog.Error("S3アップロードに失敗", "error", err, "s3_key", s3Key)
		return nil, fmt.Errorf("S3アップロードに失敗: %w", err)
	}
	slog.Info("S3アップロード完了", "s3_key", s3Key)

	output := &Output{
		S3Key:       s3Key,
		ImportJobID: event.ImportJobID,
		CityCode:    event.CityCode,
	}

	slog.Info("wagri-fetcher完了", "output", output)
	return output, nil
}
