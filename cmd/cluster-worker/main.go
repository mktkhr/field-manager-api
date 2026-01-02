// Package main はクラスター計算ワーカーのエントリポイント
//
// このワーカーは以下のモードで動作可能:
//   - RUN_ONCE=true: 1回実行して終了（Lambda/K8s Job向け）
//   - RUN_ONCE=false: デーモンモードでポーリング実行
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/application/usecase"
	clusterRepo "github.com/mktkhr/field-manager-api/internal/features/cluster/infrastructure/repository"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/cache"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/postgres"
	"github.com/mktkhr/field-manager-api/internal/logger"
	"github.com/mktkhr/field-manager-api/internal/utils"
)

const (
	defaultBatchSize    = 10
	defaultPollInterval = 60 * time.Second
)

func main() {
	// 設定読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// ログ設定
	logger.Setup(cfg.Logger)

	slog.Info("クラスターワーカーを起動しています...")

	// コンテキスト設定（シグナルハンドリング）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("シャットダウンシグナルを受信しました")
		cancel()
	}()

	// DB接続
	pool, err := postgres.CreateConnectionPool(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("DB接続に失敗しました: %v", err)
	}
	defer pool.Close()

	// Redis接続
	redisClient := cache.NewRedisClient(cfg.Cache)
	cacheClient := cache.NewClient(redisClient)
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("Redis接続のクローズに失敗しました", slog.String("error", err.Error()))
		}
	}()

	// リポジトリ作成
	clusterRepository := clusterRepo.NewClusterPostgresRepository(pool)
	clusterCacheRepository := clusterRepo.NewClusterCacheRedisRepository(cacheClient)
	clusterJobRepository := clusterRepo.NewClusterJobPostgresRepository(pool)

	// ユースケース作成
	calculateUC := usecase.NewCalculateClustersUseCase(
		clusterRepository,
		clusterCacheRepository,
		slog.Default(),
	)

	processJobsUC := usecase.NewProcessJobsUseCase(
		clusterJobRepository,
		calculateUC,
		slog.Default(),
	)

	// 環境変数から設定を読み込み
	batchSize := getEnvInt("BATCH_SIZE", defaultBatchSize)
	runOnce := getEnvBool("RUN_ONCE", false)

	slog.Info("ワーカー設定",
		slog.Int("batch_size", batchSize),
		slog.Bool("run_once", runOnce))

	if runOnce {
		// 1回実行モード（Lambda/K8s Job向け）
		slog.Info("1回実行モードで起動します")
		if err := processJobsUC.Execute(ctx, usecase.ProcessJobsInput{
			BatchSize: utils.SafeIntToInt32(batchSize),
		}); err != nil {
			log.Fatalf("ジョブ処理に失敗しました: %v", err)
		}
		slog.Info("1回実行モードが完了しました")
		return
	}

	// デーモンモード（ポーリング）
	pollInterval := getEnvDuration("POLL_INTERVAL", defaultPollInterval)
	slog.Info("デーモンモードで起動します",
		slog.Duration("poll_interval", pollInterval))

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("ワーカーを停止します")
			return
		case <-ticker.C:
			if err := processJobsUC.Execute(ctx, usecase.ProcessJobsInput{
				BatchSize: utils.SafeIntToInt32(batchSize),
			}); err != nil {
				slog.Error("ジョブ処理に失敗しました",
					slog.String("error", err.Error()))
			}
		}
	}
}

// getEnvInt は環境変数から整数値を取得する
func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

// getEnvBool は環境変数からブール値を取得する
func getEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

// getEnvDuration は環境変数からDurationを取得する
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
