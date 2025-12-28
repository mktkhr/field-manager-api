// Package cache はキャッシュストア（Redis/Valkey）のクライアント実装を提供する
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client はRedis/Valkeyクライアントのラッパー構造体
type Client struct {
	client *redis.Client
}

// NewClient はRedisクライアントからCacheクライアントを作成する
func NewClient(redisClient *redis.Client) *Client {
	return &Client{
		client: redisClient,
	}
}

// Get はキーから値を取得する
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set はキーに値を設定する（TTL付き）
func (c *Client) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Delete はキーを削除する
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Ping は接続確認を行う
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close はRedisクライアントをクローズする
func (c *Client) Close() error {
	return c.client.Close()
}
