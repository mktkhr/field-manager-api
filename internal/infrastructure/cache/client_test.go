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
		t.Fatalf("miniredis.Run()が失敗しました: %v", err)
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
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close()でエラー発生 = %v", err)
		}
	}()

	if client == nil {
		t.Fatal("NewClient()はnilを返すべきではない")
	}
	if client.client == nil {
		t.Error("NewClient()は内部のRedisクライアントを設定すべき")
	}
}

func TestClient_SetAndGet(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close()でエラー発生 = %v", err)
		}
	}()

	ctx := context.Background()

	t.Run("値の設定と取得", func(t *testing.T) {
		err := client.Set(ctx, "test-key", "test-value", time.Minute)
		if err != nil {
			t.Errorf("Set()でエラー発生 = %v", err)
		}

		value, err := client.Get(ctx, "test-key")
		if err != nil {
			t.Errorf("Get()でエラー発生 = %v", err)
		}
		if value != "test-value" {
			t.Errorf("Get() = %v, 期待値 %v", value, "test-value")
		}
	})

	t.Run("存在しないキーの取得", func(t *testing.T) {
		_, err := client.Get(ctx, "non-existent-key")
		if err == nil {
			t.Error("Get()は存在しないキーに対してエラーを返すべき")
		}
	})
}

func TestClient_Delete(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close()でエラー発生 = %v", err)
		}
	}()

	ctx := context.Background()

	// 値を設定
	err := client.Set(ctx, "delete-key", "value", time.Minute)
	if err != nil {
		t.Fatalf("Set()でエラー発生 = %v", err)
	}

	// 削除
	err = client.Delete(ctx, "delete-key")
	if err != nil {
		t.Errorf("Delete()でエラー発生 = %v", err)
	}

	// 削除後は取得できない
	_, err = client.Get(ctx, "delete-key")
	if err == nil {
		t.Error("Get()はDelete()後にエラーを返すべき")
	}
}

func TestClient_Ping(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close()でエラー発生 = %v", err)
		}
	}()

	ctx := context.Background()

	err := client.Ping(ctx)
	if err != nil {
		t.Errorf("Ping()でエラー発生 = %v", err)
	}
}

func TestClient_Close(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	err := client.Close()
	if err != nil {
		t.Errorf("Close()でエラー発生 = %v", err)
	}
}

func TestClient_SetWithTTL(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close()でエラー発生 = %v", err)
		}
	}()

	ctx := context.Background()

	err := client.Set(ctx, "ttl-key", "ttl-value", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set()でエラー発生 = %v", err)
	}

	// 値が存在することを確認
	value, err := client.Get(ctx, "ttl-key")
	if err != nil {
		t.Errorf("Get()でエラー発生 = %v", err)
	}
	if value != "ttl-value" {
		t.Errorf("Get() = %v, 期待値 %v", value, "ttl-value")
	}

	// TTL経過後
	mr.FastForward(200 * time.Millisecond)

	_, err = client.Get(ctx, "ttl-key")
	if err == nil {
		t.Error("Get()はTTL経過後にエラーを返すべき")
	}
}
