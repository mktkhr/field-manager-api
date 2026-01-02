package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/cache"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/postgres"
	"github.com/mktkhr/field-manager-api/internal/logger"
	"github.com/mktkhr/field-manager-api/internal/server"
)

func main() {
	// 設定読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// ログ設定
	logger.Setup(cfg.Logger)

	slog.Info("サーバーを起動しています...")
	slog.Info("設定読み込み完了",
		"database", fmt.Sprintf("%s:%d/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name),
		"cache", fmt.Sprintf("%s:%d", cfg.Cache.Host, cfg.Cache.Port),
		"storage", cfg.Storage.Endpoint,
	)

	ctx := context.Background()

	// データベース接続
	pool, err := postgres.CreateConnectionPool(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer pool.Close()

	// Redisクライアント作成
	redisClient := cache.NewRedisClient(cfg.Cache)
	cacheClient := cache.NewClient(redisClient)
	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			slog.Error("Redisクライアントのクローズに失敗", "error", closeErr)
		}
	}()

	// ハンドラー作成
	appLogger := slog.Default()
	handler := server.NewStrictServerHandler(pool, cacheClient, appLogger)

	// ルーターセットアップ
	router := server.SetupRouter(handler)

	slog.Info("サーバーがポート :8080 で起動しました")
	if err := router.Run(":8080"); err != nil {
		slog.Error("サーバー起動エラー", "error", err)
	}
}
