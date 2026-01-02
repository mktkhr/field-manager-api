package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestResolution_IsValid はIsValidメソッドが有効な解像度を正しく判定することをテストする
func TestResolution_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		resolution Resolution
		want       bool
	}{
		// 正常系: 有効な解像度
		{"res3は有効", Res3, true},
		{"res5は有効", Res5, true},
		{"res7は有効", Res7, true},
		{"res9は有効", Res9, true},
		// 異常系: 無効な解像度
		{"res0は無効", Resolution(0), false},
		{"res1は無効", Resolution(1), false},
		{"res2は無効", Resolution(2), false},
		{"res4は無効", Resolution(4), false},
		{"res6は無効", Resolution(6), false},
		{"res8は無効", Resolution(8), false},
		{"res10は無効", Resolution(10), false},
		{"負の値は無効", Resolution(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resolution.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestResolution_String はStringメソッドが解像度を文字列として正しく返すことをテストする
func TestResolution_String(t *testing.T) {
	tests := []struct {
		name       string
		resolution Resolution
		want       string
	}{
		{"res3の文字列表現", Res3, "res3"},
		{"res5の文字列表現", Res5, "res5"},
		{"res7の文字列表現", Res7, "res7"},
		{"res9の文字列表現", Res9, "res9"},
		{"未知の解像度の文字列表現", Resolution(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resolution.String()
			if got != tt.want {
				t.Errorf("String() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestAllResolutions はAllResolutionsが全てのサポートされている解像度を含むことをテストする
func TestAllResolutions(t *testing.T) {
	// 期待する解像度のリスト
	expected := []Resolution{Res3, Res5, Res7, Res9}

	if len(AllResolutions) != len(expected) {
		t.Errorf("AllResolutionsの長さ = %d, 期待値 %d", len(AllResolutions), len(expected))
	}

	for i, res := range expected {
		if AllResolutions[i] != res {
			t.Errorf("AllResolutions[%d] = %v, 期待値 %v", i, AllResolutions[i], res)
		}
	}
}

// TestNewCluster はNewClusterが正しい値でClusterを生成することをテストする
func TestNewCluster(t *testing.T) {
	resolution := Res7
	h3Index := "871f1a4adffffff"
	fieldCount := int32(42)
	centerLat := 35.681236
	centerLng := 139.767125

	before := time.Now()
	cluster := NewCluster(resolution, h3Index, fieldCount, centerLat, centerLng)
	after := time.Now()

	require.NotNil(t, cluster, "Clusterがnilです")

	// IDが設定されていることを確認
	if cluster.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("IDがゼロ値です")
	}

	if cluster.Resolution != resolution {
		t.Errorf("Resolution = %v, 期待値 %v", cluster.Resolution, resolution)
	}

	if cluster.H3Index != h3Index {
		t.Errorf("H3Index = %q, 期待値 %q", cluster.H3Index, h3Index)
	}

	if cluster.FieldCount != fieldCount {
		t.Errorf("FieldCount = %d, 期待値 %d", cluster.FieldCount, fieldCount)
	}

	if cluster.CenterLat != centerLat {
		t.Errorf("CenterLat = %f, 期待値 %f", cluster.CenterLat, centerLat)
	}

	if cluster.CenterLng != centerLng {
		t.Errorf("CenterLng = %f, 期待値 %f", cluster.CenterLng, centerLng)
	}

	// CalculatedAtが適切な時間範囲内であることを確認
	if cluster.CalculatedAt.Before(before) || cluster.CalculatedAt.After(after) {
		t.Errorf("CalculatedAt = %v, 期待範囲 %v - %v", cluster.CalculatedAt, before, after)
	}
}

// TestCluster_ToResult はToResultメソッドがClusterResultに正しく変換することをテストする
func TestCluster_ToResult(t *testing.T) {
	cluster := &Cluster{
		H3Index:    "871f1a4adffffff",
		CenterLat:  35.681236,
		CenterLng:  139.767125,
		FieldCount: 42,
	}

	result := cluster.ToResult()

	require.NotNil(t, result, "ClusterResultがnilです")

	if result.H3Index != cluster.H3Index {
		t.Errorf("H3Index = %q, 期待値 %q", result.H3Index, cluster.H3Index)
	}

	if result.Lat != cluster.CenterLat {
		t.Errorf("Lat = %f, 期待値 %f", result.Lat, cluster.CenterLat)
	}

	if result.Lng != cluster.CenterLng {
		t.Errorf("Lng = %f, 期待値 %f", result.Lng, cluster.CenterLng)
	}

	if result.Count != cluster.FieldCount {
		t.Errorf("Count = %d, 期待値 %d", result.Count, cluster.FieldCount)
	}
}

// TestCluster_ToResult_ZeroValues はToResultメソッドがゼロ値でも正しく動作することをテストする
func TestCluster_ToResult_ZeroValues(t *testing.T) {
	cluster := &Cluster{
		H3Index:    "",
		CenterLat:  0,
		CenterLng:  0,
		FieldCount: 0,
	}

	result := cluster.ToResult()

	require.NotNil(t, result, "ClusterResultがnilです")

	if result.H3Index != "" {
		t.Errorf("H3Index = %q, 期待値 空文字列", result.H3Index)
	}

	if result.Lat != 0 {
		t.Errorf("Lat = %f, 期待値 0", result.Lat)
	}

	if result.Lng != 0 {
		t.Errorf("Lng = %f, 期待値 0", result.Lng)
	}

	if result.Count != 0 {
		t.Errorf("Count = %d, 期待値 0", result.Count)
	}
}

// TestClusterResult はClusterResultの構造体フィールドが正しくアクセスできることをテストする
func TestClusterResult(t *testing.T) {
	result := &ClusterResult{
		H3Index: "871f1a4adffffff",
		Lat:     35.681236,
		Lng:     139.767125,
		Count:   100,
	}

	if result.H3Index != "871f1a4adffffff" {
		t.Errorf("H3Index = %q, 期待値 871f1a4adffffff", result.H3Index)
	}

	if result.Lat != 35.681236 {
		t.Errorf("Lat = %f, 期待値 35.681236", result.Lat)
	}

	if result.Lng != 139.767125 {
		t.Errorf("Lng = %f, 期待値 139.767125", result.Lng)
	}

	if result.Count != 100 {
		t.Errorf("Count = %d, 期待値 100", result.Count)
	}
}
