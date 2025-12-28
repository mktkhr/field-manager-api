package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() failed: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	client := NewClient(redisClient)
	return mr, client
}

func TestNewClient(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	if client == nil {
		t.Error("NewClient() should not return nil")
	}
	if client.client == nil {
		t.Error("NewClient() should set internal redis client")
	}
}

func TestClient_SetAndGet(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("値の設定と取得", func(t *testing.T) {
		err := client.Set(ctx, "test-key", "test-value", time.Minute)
		if err != nil {
			t.Errorf("Set() error = %v", err)
		}

		value, err := client.Get(ctx, "test-key")
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if value != "test-value" {
			t.Errorf("Get() = %v, want %v", value, "test-value")
		}
	})

	t.Run("存在しないキーの取得", func(t *testing.T) {
		_, err := client.Get(ctx, "non-existent-key")
		if err == nil {
			t.Error("Get() should return error for non-existent key")
		}
	})
}

func TestClient_Delete(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	// 値を設定
	err := client.Set(ctx, "delete-key", "value", time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 削除
	err = client.Delete(ctx, "delete-key")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// 削除後は取得できない
	_, err = client.Get(ctx, "delete-key")
	if err == nil {
		t.Error("Get() should return error after Delete()")
	}
}

func TestClient_Ping(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	err := client.Ping(ctx)
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestClient_Close(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestClient_SetWithTTL(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	err := client.Set(ctx, "ttl-key", "ttl-value", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 値が存在することを確認
	value, err := client.Get(ctx, "ttl-key")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if value != "ttl-value" {
		t.Errorf("Get() = %v, want %v", value, "ttl-value")
	}

	// TTL経過後
	mr.FastForward(200 * time.Millisecond)

	_, err = client.Get(ctx, "ttl-key")
	if err == nil {
		t.Error("Get() should return error after TTL expired")
	}
}
