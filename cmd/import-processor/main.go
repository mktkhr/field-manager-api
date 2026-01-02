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
	clusterUsecase "github.com/mktkhr/field-manager-api/internal/features/cluster/application/usecase"
	clusterRepo "github.com/mktkhr/field-manager-api/internal/features/cluster/infrastructure/repository"
	fieldRepo "github.com/mktkhr/field-manager-api/internal/features/field/infrastructure/repository"
	importExternal "github.com/mktkhr/field-manager-api/internal/features/import/infrastructure/external"
	importRepo "github.com/mktkhr/field-manager-api/internal/features/import/infrastructure/repository"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/cache"

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

	cacheCfg, err := config.LoadCacheConfig()
	if err != nil {
		return fmt.Errorf("cache設定の読み込みに失敗: %w", err)
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

	// Redisクライアント作成
	redisClient := cache.NewRedisClient(*cacheCfg)
	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			logger.Error("Redisクライアントのクローズに失敗", slog.String("error", closeErr.Error()))
		}
	}()

	// リポジトリ作成
	importJobRepository := importRepo.NewImportJobRepository(pool, logger)
	fieldRepository := fieldRepo.NewFieldRepository(pool, logger)
	clusterJobRepository := clusterRepo.NewClusterJobPostgresRepository(pool)

	// クラスタージョブエンキューアー作成
	enqueueJobUC := clusterUsecase.NewEnqueueJobUseCase(clusterJobRepository, logger)
	clusterJobEnqueuer := clusterUsecase.NewClusterJobEnqueuer(enqueueJobUC)

	// ユースケース作成
	processImportUC := usecase.NewProcessImportUseCase(
		importJobRepository,
		s3Client,
		fieldRepository,
		clusterJobEnqueuer,
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
	// パスワードに特殊文字が含まれる場合に備えてurl.UserPasswordを使用
	// url.UserPasswordはuserinfoセクションに適したエスケープを行う
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}
	q := u.Query()
	q.Set("sslmode", cfg.SSLMode)
	u.RawQuery = q.Encode()

	return pgxpool.New(ctx, u.String())
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
		if _, err := fmt.Sscanf(v, "%d", &i); err != nil {
			slog.Warn("環境変数のパースに失敗。デフォルト値を使用します", "key", key, "value", v, "default", defaultValue, "error", err)
		} else {
			return i
		}
	}
	return defaultValue
}
