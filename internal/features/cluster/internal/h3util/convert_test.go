package h3util

import (
	"testing"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/stretchr/testify/require"
)

// TestConvertAggregatedToClusters はConvertAggregatedToClustersが集計結果を正しくClusterエンティティに変換することをテストする
func TestConvertAggregatedToClusters(t *testing.T) {
	tests := []struct {
		name       string
		resolution entity.Resolution
		aggregated []*repository.AggregatedCluster
		wantLen    int
	}{
		// 正常系: 有効な集計結果を変換
		{
			name:       "有効な集計結果を変換",
			resolution: entity.Res7,
			aggregated: []*repository.AggregatedCluster{
				{H3Index: "871f1a4adffffff", FieldCount: 10},
				{H3Index: "871f1a4aeffffff", FieldCount: 5},
			},
			wantLen: 2,
		},
		// 正常系: 空のスライス
		{
			name:       "空のスライスは空を返す",
			resolution: entity.Res7,
			aggregated: []*repository.AggregatedCluster{},
			wantLen:    0,
		},
		// 異常系: 無効なH3インデックスはスキップ
		{
			name:       "無効なH3インデックスはスキップ",
			resolution: entity.Res7,
			aggregated: []*repository.AggregatedCluster{
				{H3Index: "invalid", FieldCount: 10},
				{H3Index: "871f1a4adffffff", FieldCount: 5},
			},
			wantLen: 1,
		},
		// 異常系: 全て無効なH3インデックス
		{
			name:       "全て無効なH3インデックスの場合は空を返す",
			resolution: entity.Res7,
			aggregated: []*repository.AggregatedCluster{
				{H3Index: "invalid1", FieldCount: 10},
				{H3Index: "invalid2", FieldCount: 5},
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertAggregatedToClusters(tt.resolution, tt.aggregated)

			require.NoError(t, err, "ConvertAggregatedToClustersでエラーが発生")

			if len(result) != tt.wantLen {
				t.Errorf("結果の長さ = %d, 期待値 %d", len(result), tt.wantLen)
			}

			// 変換結果の検証
			for _, cluster := range result {
				if cluster.Resolution != tt.resolution {
					t.Errorf("Resolution = %v, 期待値 %v", cluster.Resolution, tt.resolution)
				}

				if cluster.H3Index == "" {
					t.Error("H3Indexが空文字列です")
				}

				// UUIDが設定されていることを確認
				if cluster.ID.String() == "00000000-0000-0000-0000-000000000000" {
					t.Error("IDがゼロ値です")
				}

				// CalculatedAtが設定されていることを確認
				if cluster.CalculatedAt.IsZero() {
					t.Error("CalculatedAtがゼロ値です")
				}
			}
		})
	}
}

// TestConvertAggregatedToClusters_FieldCount は変換時にFieldCountが正しく設定されることをテストする
func TestConvertAggregatedToClusters_FieldCount(t *testing.T) {
	aggregated := []*repository.AggregatedCluster{
		{H3Index: "871f1a4adffffff", FieldCount: 42},
	}

	result, err := ConvertAggregatedToClusters(entity.Res7, aggregated)

	require.NoError(t, err, "ConvertAggregatedToClustersでエラーが発生")
	require.Len(t, result, 1, "結果の長さが1ではありません")

	if result[0].FieldCount != 42 {
		t.Errorf("FieldCount = %d, 期待値 42", result[0].FieldCount)
	}
}

// TestConvertAggregatedToClusters_CenterCoordinates は変換時に中心座標が正しく計算されることをテストする
func TestConvertAggregatedToClusters_CenterCoordinates(t *testing.T) {
	aggregated := []*repository.AggregatedCluster{
		{H3Index: "871f1a4adffffff", FieldCount: 1},
	}

	result, err := ConvertAggregatedToClusters(entity.Res7, aggregated)

	require.NoError(t, err, "ConvertAggregatedToClustersでエラーが発生")
	require.Len(t, result, 1, "結果の長さが1ではありません")

	// 座標が有効な範囲内であることを確認
	if result[0].CenterLat < -90 || result[0].CenterLat > 90 {
		t.Errorf("CenterLat = %f, 有効範囲外です", result[0].CenterLat)
	}

	if result[0].CenterLng < -180 || result[0].CenterLng > 180 {
		t.Errorf("CenterLng = %f, 有効範囲外です", result[0].CenterLng)
	}
}

// TestConvertAggregatedToClusters_AllResolutions は全解像度で正しく変換されることをテストする
func TestConvertAggregatedToClusters_AllResolutions(t *testing.T) {
	resolutions := []entity.Resolution{entity.Res3, entity.Res5, entity.Res7, entity.Res9}

	for _, res := range resolutions {
		t.Run(res.String(), func(t *testing.T) {
			aggregated := []*repository.AggregatedCluster{
				{H3Index: "871f1a4adffffff", FieldCount: 1},
			}

			result, err := ConvertAggregatedToClusters(res, aggregated)

			require.NoError(t, err, "ConvertAggregatedToClustersでエラーが発生")
			require.Len(t, result, 1, "結果の長さが1ではありません")

			if result[0].Resolution != res {
				t.Errorf("Resolution = %v, 期待値 %v", result[0].Resolution, res)
			}
		})
	}
}

// TestConvertAggregatedToClusters_NilInput はnilの入力でもパニックしないことをテストする
func TestConvertAggregatedToClusters_NilInput(t *testing.T) {
	result, err := ConvertAggregatedToClusters(entity.Res7, nil)

	require.NoError(t, err, "ConvertAggregatedToClustersでエラーが発生")

	if result == nil {
		t.Error("結果がnilです、空のスライスを期待")
	}

	if len(result) != 0 {
		t.Errorf("結果の長さ = %d, 期待値 0", len(result))
	}
}
