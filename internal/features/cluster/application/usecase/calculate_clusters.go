// Package usecase はクラスタリング機能のユースケースを提供する
package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	infraRepo "github.com/mktkhr/field-manager-api/internal/features/cluster/infrastructure/repository"
)

// CalculateClustersUseCase はクラスター計算ユースケース
type CalculateClustersUseCase struct {
	clusterRepo repository.ClusterRepository
	cacheRepo   repository.ClusterCacheRepository
	logger      *slog.Logger
}

// NewCalculateClustersUseCase はCalculateClustersUseCaseを作成する
func NewCalculateClustersUseCase(
	clusterRepo repository.ClusterRepository,
	cacheRepo repository.ClusterCacheRepository,
	logger *slog.Logger,
) *CalculateClustersUseCase {
	return &CalculateClustersUseCase{
		clusterRepo: clusterRepo,
		cacheRepo:   cacheRepo,
		logger:      logger,
	}
}

// Execute はクラスター計算を実行する
// 全解像度でfieldsテーブルを集計し、cluster_resultsテーブルに保存する
func (u *CalculateClustersUseCase) Execute(ctx context.Context) error {
	u.logger.Info("クラスター計算を開始します")

	// 全解像度で処理
	for _, resolution := range entity.AllResolutions {
		if err := u.calculateForResolution(ctx, resolution); err != nil {
			return fmt.Errorf("解像度%sの計算に失敗しました: %w", resolution.String(), err)
		}
	}

	// キャッシュをクリア
	if err := u.cacheRepo.DeleteClusters(ctx); err != nil {
		u.logger.Warn("キャッシュのクリアに失敗しました",
			slog.String("error", err.Error()))
	}

	u.logger.Info("クラスター計算が完了しました")
	return nil
}

// calculateForResolution は指定解像度でクラスター計算を実行する
func (u *CalculateClustersUseCase) calculateForResolution(ctx context.Context, resolution entity.Resolution) error {
	u.logger.Info("解像度別のクラスター計算を開始します",
		slog.String("resolution", resolution.String()))

	// fieldsテーブルからH3インデックスで集計
	aggregated, err := u.clusterRepo.AggregateByH3(ctx, resolution)
	if err != nil {
		return fmt.Errorf("集計に失敗しました: %w", err)
	}

	if len(aggregated) == 0 {
		u.logger.Info("集計結果が0件でした",
			slog.String("resolution", resolution.String()))
		return nil
	}

	// 集計結果をClusterエンティティに変換
	clusters, err := infraRepo.ConvertAggregatedToClusters(resolution, aggregated)
	if err != nil {
		return fmt.Errorf("変換に失敗しました: %w", err)
	}

	// cluster_resultsテーブルに保存
	if err := u.clusterRepo.SaveClusters(ctx, clusters); err != nil {
		return fmt.Errorf("保存に失敗しました: %w", err)
	}

	u.logger.Info("解像度別のクラスター計算が完了しました",
		slog.String("resolution", resolution.String()),
		slog.Int("cluster_count", len(clusters)))

	return nil
}
