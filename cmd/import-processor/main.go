package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/config"
	fieldRepo "github.com/mktkhr/field-manager-api/internal/features/field/infrastructure/repository"
	importExternal "github.com/mktkhr/field-manager-api/internal/features/import/infrastructure/external"
	importRepo "github.com/mktkhr/field-manager-api/internal/features/import/infrastructure/repository"

	"github.com/mktkhr/field-manager-api/internal/features/import/application/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// コマンドライン引数のパース
	importJobID := flag.String("import-job-id", "", "インポートジョブID (必須)")
	s3Key := flag.String("s3-key", "", "S3キー (必須)")
	batchSize := flag.Int("batch-size", usecase.DefaultBatchSize, "バッチサイズ")
	flag.Parse()

	if *importJobID == "" || *s3Key == "" {
		slog.Error("必須パラメータが不足しています")
		flag.Usage()
		os.Exit(1)
	}

	jobID, err := uuid.Parse(*importJobID)
	if err != nil {
		slog.Error("インポートジョブIDのパースに失敗", "error", err, "import_job_id", *importJobID)
		os.Exit(1)
	}

	slog.Info("import-processor開始", "import_job_id", jobID, "s3_key", *s3Key, "batch_size", *batchSize)

	ctx := context.Background()

	if err := run(ctx, jobID, *s3Key, *batchSize, logger); err != nil {
		slog.Error("処理に失敗", "error", err)
		os.Exit(1)
	}

	slog.Info("import-processor完了")
}

func run(ctx context.Context, importJobID uuid.UUID, s3Key string, batchSize int, logger *slog.Logger) error {
	// 設定読み込み
	dbCfg, err := loadDatabaseConfig()
	if err != nil {
		return fmt.Errorf("データベース設定の読み込みに失敗: %w", err)
	}

	storageCfg, err := config.LoadStorageConfig()
	if err != nil {
		return fmt.Errorf("storage設定の読み込みに失敗: %w", err)
	}

	// データベース接続
	pool, err := createDBPool(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("データベース接続に失敗: %w", err)
	}
	defer pool.Close()

	// S3クライアント作成
	s3Client, err := importExternal.NewS3ClientFromStorageConfig(ctx, storageCfg)
	if err != nil {
		return fmt.Errorf("s3クライアントの作成に失敗: %w", err)
	}

	// リポジトリ作成
	importJobRepository := importRepo.NewImportJobRepository(pool)
	fieldRepository := fieldRepo.NewFieldRepository(pool)

	// ユースケース作成
	processImportUC := usecase.NewProcessImportUseCase(
		importJobRepository,
		s3Client,
		fieldRepository,
		logger,
	)

	// 処理実行
	input := usecase.ProcessImportInput{
		ImportJobID: importJobID,
		S3Key:       s3Key,
		BatchSize:   batchSize,
	}

	return processImportUC.Execute(ctx, input)
}

// loadDatabaseConfig はデータベース設定を読み込む
func loadDatabaseConfig() (*config.DatabaseConfig, error) {
	cfg := &config.DatabaseConfig{}
	// 簡易的に環境変数から読み込み
	cfg.Host = getEnvOrDefault("DB_HOST", "localhost")
	cfg.Port = getEnvIntOrDefault("DB_PORT", 5432)
	cfg.User = getEnvOrDefault("DB_USER", "postgres")
	cfg.Password = os.Getenv("DB_PASSWORD")
	cfg.Name = getEnvOrDefault("DB_NAME", "field_manager_db")
	cfg.SSLMode = getEnvOrDefault("DB_SSL_MODE", "disable")
	return cfg, nil
}

// createDBPool はデータベース接続プールを作成する
func createDBPool(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// パスワードに特殊文字が含まれる場合に備えてURLエンコードを使用
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(cfg.User),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)
	return pgxpool.New(ctx, connString)
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return defaultValue
}
