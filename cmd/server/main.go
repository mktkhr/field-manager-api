package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mktkhr/field-manager-api/internal/config"
)

func main() {
	fmt.Println("サーバーを起動しています...")

	// 設定読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	log.Printf("データベース接続先: %s:%d/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	log.Printf("キャッシュ接続先: %s:%d", cfg.Cache.Host, cfg.Cache.Port)
	log.Printf("ストレージエンドポイント: %s", cfg.Storage.Endpoint)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("サーバーがポート :8080 で起動しました")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
