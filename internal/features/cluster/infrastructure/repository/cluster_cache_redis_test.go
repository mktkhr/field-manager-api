package repository

import (
	"testing"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
)

// TestBuildCacheKey はbuildCacheKeyが正しいキーを生成することをテストする
func TestBuildCacheKey(t *testing.T) {
	tests := []struct {
		name       string
		resolution entity.Resolution
		wantKey    string
	}{
		{
			name:       "res3のキー",
			resolution: entity.Res3,
			wantKey:    "cluster:results:res3",
		},
		{
			name:       "res5のキー",
			resolution: entity.Res5,
			wantKey:    "cluster:results:res5",
		},
		{
			name:       "res7のキー",
			resolution: entity.Res7,
			wantKey:    "cluster:results:res7",
		},
		{
			name:       "res9のキー",
			resolution: entity.Res9,
			wantKey:    "cluster:results:res9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCacheKey(tt.resolution)
			if got != tt.wantKey {
				t.Errorf("buildCacheKey(%v) = %q, 期待値 %q", tt.resolution, got, tt.wantKey)
			}
		})
	}
}

// TestBuildCacheKey_UnknownResolution は未知の解像度でもキーが生成されることをテストする
func TestBuildCacheKey_UnknownResolution(t *testing.T) {
	// 未知の解像度でもパニックせずにキーを生成することを確認
	got := buildCacheKey(entity.Resolution(100))
	expected := "cluster:results:unknown"
	if got != expected {
		t.Errorf("buildCacheKey(100) = %q, 期待値 %q", got, expected)
	}
}

// TestClusterCacheKeyPrefix はキープレフィックスが正しいことをテストする
func TestClusterCacheKeyPrefix(t *testing.T) {
	expectedPrefix := "cluster:results:"
	if clusterCacheKeyPrefix != expectedPrefix {
		t.Errorf("clusterCacheKeyPrefix = %q, 期待値 %q", clusterCacheKeyPrefix, expectedPrefix)
	}
}

// TestClusterCacheTTL はキャッシュTTLが30分であることをテストする
func TestClusterCacheTTL(t *testing.T) {
	expectedMinutes := 30
	actualMinutes := int(clusterCacheTTL.Minutes())
	if actualMinutes != expectedMinutes {
		t.Errorf("clusterCacheTTL = %d分, 期待値 %d分", actualMinutes, expectedMinutes)
	}
}

// TestClusterCacheData はclusterCacheDataの構造体が正しくシリアライズ可能であることをテストする
func TestClusterCacheData(t *testing.T) {
	data := clusterCacheData{
		H3Index:      "871f1a4adffffff",
		FieldCount:   42,
		CenterLat:    35.681236,
		CenterLng:    139.767125,
		CalculatedAt: 1704067200, // 2024-01-01 00:00:00 UTC
	}

	if data.H3Index != "871f1a4adffffff" {
		t.Errorf("H3Index = %q, 期待値 871f1a4adffffff", data.H3Index)
	}

	if data.FieldCount != 42 {
		t.Errorf("FieldCount = %d, 期待値 42", data.FieldCount)
	}

	if data.CenterLat != 35.681236 {
		t.Errorf("CenterLat = %f, 期待値 35.681236", data.CenterLat)
	}

	if data.CenterLng != 139.767125 {
		t.Errorf("CenterLng = %f, 期待値 139.767125", data.CenterLng)
	}

	if data.CalculatedAt != 1704067200 {
		t.Errorf("CalculatedAt = %d, 期待値 1704067200", data.CalculatedAt)
	}
}

// TestClusterCacheData_ZeroValues はゼロ値でも正しく動作することをテストする
func TestClusterCacheData_ZeroValues(t *testing.T) {
	data := clusterCacheData{}

	if data.H3Index != "" {
		t.Errorf("H3Index = %q, 期待値 空文字列", data.H3Index)
	}

	if data.FieldCount != 0 {
		t.Errorf("FieldCount = %d, 期待値 0", data.FieldCount)
	}

	if data.CenterLat != 0 {
		t.Errorf("CenterLat = %f, 期待値 0", data.CenterLat)
	}

	if data.CenterLng != 0 {
		t.Errorf("CenterLng = %f, 期待値 0", data.CenterLng)
	}

	if data.CalculatedAt != 0 {
		t.Errorf("CalculatedAt = %d, 期待値 0", data.CalculatedAt)
	}
}
