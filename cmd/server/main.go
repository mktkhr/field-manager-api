package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/logger"
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

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			slog.Error("レスポンス書き込みエラー", "error", err)
		}
	})

	slog.Info("サーバーがポート :8080 で起動しました")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("サーバー起動エラー", "error", err)
	}
}
