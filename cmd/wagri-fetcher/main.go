package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
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

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	lambda.Start(handler)
}

func handler(ctx context.Context, event Event) (*Output, error) {
	slog.Info("wagri-fetcher開始", "city_code", event.CityCode, "import_job_id", event.ImportJobID)

	// Storage設定読み込み
	storageCfg, err := config.LoadStorageConfig()
	if err != nil {
		slog.Error("Storage設定の読み込みに失敗", "error", err)
		return nil, fmt.Errorf("storage設定の読み込みに失敗: %w", err)
	}

	// Wagri設定読み込み
	wagriCfg, err := config.LoadWagriConfig()
	if err != nil {
		slog.Error("Wagri設定の読み込みに失敗", "error", err)
		return nil, fmt.Errorf("wagri設定の読み込みに失敗: %w", err)
	}

	// S3クライアント作成(StorageConfigで切り替え)
	s3Client, err := external.NewS3ClientFromStorageConfig(ctx, storageCfg)
	if err != nil {
		slog.Error("S3クライアントの作成に失敗", "error", err)
		return nil, fmt.Errorf("s3クライアントの作成に失敗: %w", err)
	}

	// Wagriクライアント作成
	wagriClient := external.NewWagriClient(wagriCfg)

	// Wagri APIから圃場データを取得
	slog.Info("Wagri API呼び出し開始", "city_code", event.CityCode)
	data, err := wagriClient.FetchFieldsByCityCodeToStream(ctx, event.CityCode)
	if err != nil {
		slog.Error("Wagri API呼び出しに失敗", "error", err, "city_code", event.CityCode)
		return nil, fmt.Errorf("wagri API呼び出しに失敗: %w", err)
	}
	slog.Info("Wagri API呼び出し完了", "size_bytes", len(data))

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
