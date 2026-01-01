package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/cache"
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
	pool, err := createDBPool(ctx, &cfg.Database)
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

// createDBPool はデータベース接続プールを作成する
func createDBPool(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// パスワードに特殊文字が含まれる場合に備えてurl.UserPasswordを使用
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
