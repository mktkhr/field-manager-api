// Package port はアプリケーション層の外部依存インターフェースを定義する
package port

import (
	"context"
	"time"
)

// CacheStore はキャッシュストアの操作を定義するインターフェース
type CacheStore interface {
	// Get はキーから値を取得する
	Get(ctx context.Context, key string) (string, error)

	// Set はキーに値を設定する（TTL付き）
	Set(ctx context.Context, key string, value string, expiration time.Duration) error

	// Delete はキーを削除する
	Delete(ctx context.Context, key string) error

	// Ping は接続確認を行う
	Ping(ctx context.Context) error

	// Close はクライアントをクローズする
	Close() error
}
