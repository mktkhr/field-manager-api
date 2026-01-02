package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/stretchr/testify/require"
)

// TestNewCalculateClustersUseCase はNewCalculateClustersUseCaseが正しくUseCaseを生成することをテストする
func TestNewCalculateClustersUseCase(t *testing.T) {
	clusterRepo := &mockClusterRepository{}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

	require.NotNil(t, uc, "UseCaseがnilです")
}

// TestCalculateClustersUseCase_Execute_Success は正常にクラスター計算を実行することをテストする
func TestCalculateClustersUseCase_Execute_Success(t *testing.T) {
	aggregated := []*repository.AggregatedCluster{
		{H3Index: "871f1a4adffffff", FieldCount: 10},
		{H3Index: "871f1a4aeffffff", FieldCount: 5},
	}

	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

	// 全範囲再計算(空のAffectedH3Cells)
	err := uc.Execute(context.Background(), CalculateClustersInput{})

	require.NoError(t, err, "Executeでエラーが発生")
}

// TestCalculateClustersUseCase_Execute_EmptyAggregation は集計結果が0件でも正常に完了することをテストする
func TestCalculateClustersUseCase_Execute_EmptyAggregation(t *testing.T) {
	clusterRepo := &mockClusterRepository{aggregated: []*repository.AggregatedCluster{}}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

	err := uc.Execute(context.Background(), CalculateClustersInput{})

	require.NoError(t, err, "空の集計結果でもエラーにならないはず")
}

// TestCalculateClustersUseCase_Execute_AggregateError は集計エラー時にエラーを返すことをテストする
func TestCalculateClustersUseCase_Execute_AggregateError(t *testing.T) {
	clusterRepo := &mockClusterRepository{aggregateErr: errors.New("aggregate error")}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

	err := uc.Execute(context.Background(), CalculateClustersInput{})

	require.Error(t, err, "集計エラー時はエラーを返すべき")
}

// TestCalculateClustersUseCase_Execute_SaveError は保存エラー時にエラーを返すことをテストする
func TestCalculateClustersUseCase_Execute_SaveError(t *testing.T) {
	aggregated := []*repository.AggregatedCluster{
		{H3Index: "871f1a4adffffff", FieldCount: 10},
	}

	clusterRepo := &mockClusterRepository{
		aggregated: aggregated,
		saveErr:    errors.New("save error"),
	}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

	err := uc.Execute(context.Background(), CalculateClustersInput{})

	require.Error(t, err, "保存エラー時はエラーを返すべき")
}

// TestCalculateClustersUseCase_Execute_CacheDeleteError はキャッシュ削除エラーでも処理が完了することをテストする
func TestCalculateClustersUseCase_Execute_CacheDeleteError(t *testing.T) {
	aggregated := []*repository.AggregatedCluster{
		{H3Index: "871f1a4adffffff", FieldCount: 10},
	}

	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{deleteErr: errors.New("cache delete error")}
	logger := getTestLogger()

	uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

	err := uc.Execute(context.Background(), CalculateClustersInput{})

	require.NoError(t, err, "キャッシュ削除エラーでも処理は完了するべき")
}

// TestCalculateClustersUseCase_Execute_AllResolutions は全解像度で計算が実行されることをテストする
func TestCalculateClustersUseCase_Execute_AllResolutions(t *testing.T) {
	// 各解像度で異なる結果を返すモック
	aggregated := []*repository.AggregatedCluster{
		{H3Index: "871f1a4adffffff", FieldCount: 10},
	}

	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

	err := uc.Execute(context.Background(), CalculateClustersInput{})

	require.NoError(t, err, "全解像度での計算が正常に完了するべき")
}

// TestCalculateClustersUseCase_Execute_InvalidH3Index は無効なH3インデックスがスキップされることをテストする
func TestCalculateClustersUseCase_Execute_InvalidH3Index(t *testing.T) {
	aggregated := []*repository.AggregatedCluster{
		{H3Index: "invalid", FieldCount: 10},        // 無効なH3インデックス
		{H3Index: "871f1a4adffffff", FieldCount: 5}, // 有効なH3インデックス
	}

	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

	err := uc.Execute(context.Background(), CalculateClustersInput{})

	require.NoError(t, err, "無効なH3インデックスはスキップされて正常に完了するべき")
}

// TestCalculateClustersUseCase_calculateForResolution は個別解像度の計算が正しく動作することをテストする
func TestCalculateClustersUseCase_calculateForResolution(t *testing.T) {
	tests := []struct {
		name       string
		resolution entity.Resolution
	}{
		{"res3", entity.Res3},
		{"res5", entity.Res5},
		{"res7", entity.Res7},
		{"res9", entity.Res9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregated := []*repository.AggregatedCluster{
				{H3Index: "871f1a4adffffff", FieldCount: 10},
			}

			clusterRepo := &mockClusterRepository{aggregated: aggregated}
			cacheRepo := &mockClusterCacheRepository{}
			logger := getTestLogger()

			uc := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)

			err := uc.calculateForResolution(context.Background(), tt.resolution)

			require.NoError(t, err, "解像度%sの計算でエラーが発生", tt.resolution.String())
		})
	}
}
