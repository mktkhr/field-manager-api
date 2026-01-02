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

// CalculateClustersInput はクラスター計算ユースケースの入力
type CalculateClustersInput struct {
	AffectedH3Cells []string // 影響を受けたH3セル(nil or empty = 全範囲再計算)
}

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
// 入力のAffectedH3Cellsがnilまたは空の場合は全範囲再計算、それ以外は差分更新
func (u *CalculateClustersUseCase) Execute(ctx context.Context, input CalculateClustersInput) error {
	if len(input.AffectedH3Cells) == 0 {
		return u.executeFullRecalculation(ctx)
	}
	return u.executeDifferentialRecalculation(ctx, input.AffectedH3Cells)
}

// executeFullRecalculation は全範囲でクラスター計算を実行する
func (u *CalculateClustersUseCase) executeFullRecalculation(ctx context.Context) error {
	u.logger.Info("全範囲クラスター計算を開始します")

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

	u.logger.Info("全範囲クラスター計算が完了しました")
	return nil
}

// executeDifferentialRecalculation は差分でクラスター計算を実行する
func (u *CalculateClustersUseCase) executeDifferentialRecalculation(ctx context.Context, affectedH3Cells []string) error {
	u.logger.Info("差分クラスター計算を開始します",
		slog.Int("affected_cells", len(affectedH3Cells)))

	// 解像度ごとにH3セルを分類
	cellsByResolution := u.classifyH3CellsByResolution(affectedH3Cells)

	// 各解像度で差分更新
	for _, resolution := range entity.AllResolutions {
		cells := cellsByResolution[resolution]
		if len(cells) == 0 {
			continue
		}

		if err := u.calculateForResolutionDifferential(ctx, resolution, cells); err != nil {
			return fmt.Errorf("解像度%sの差分計算に失敗しました: %w", resolution.String(), err)
		}
	}

	// キャッシュをクリア
	if err := u.cacheRepo.DeleteClusters(ctx); err != nil {
		u.logger.Warn("キャッシュのクリアに失敗しました",
			slog.String("error", err.Error()))
	}

	u.logger.Info("差分クラスター計算が完了しました")
	return nil
}

// classifyH3CellsByResolution はH3セルを解像度ごとに分類する
func (u *CalculateClustersUseCase) classifyH3CellsByResolution(cells []string) map[entity.Resolution][]string {
	result := make(map[entity.Resolution][]string)
	for _, cell := range cells {
		resolution := u.detectResolution(cell)
		if resolution >= 0 {
			result[resolution] = append(result[resolution], cell)
		}
	}
	return result
}

// detectResolution はH3インデックスから解像度を検出する
// H3インデックスの長さから解像度を推定(簡易実装)
func (u *CalculateClustersUseCase) detectResolution(h3Index string) entity.Resolution {
	// H3インデックスの長さと解像度の対応:
	// res3: 約8文字 (例: 831f8bff)
	// res5: 約10文字 (例: 851f8b3ff)
	// res7: 約12文字 (例: 871f8b3c7)
	// res9: 約14文字 (例: 891f8b3c5af)
	length := len(h3Index)
	switch {
	case length <= 9:
		return entity.Res3
	case length <= 11:
		return entity.Res5
	case length <= 13:
		return entity.Res7
	default:
		return entity.Res9
	}
}

// calculateForResolutionDifferential は指定解像度で差分クラスター計算を実行する
func (u *CalculateClustersUseCase) calculateForResolutionDifferential(ctx context.Context, resolution entity.Resolution, h3Cells []string) error {
	u.logger.Info("解像度別の差分クラスター計算を開始します",
		slog.String("resolution", resolution.String()),
		slog.Int("cells", len(h3Cells)))

	// 1. 対象セルの既存クラスター結果を削除
	if err := u.clusterRepo.DeleteClustersByH3Indexes(ctx, resolution, h3Cells); err != nil {
		return fmt.Errorf("既存クラスター結果の削除に失敗しました: %w", err)
	}

	// 2. 対象セルのみ再集計
	aggregated, err := u.clusterRepo.AggregateByH3ForCells(ctx, resolution, h3Cells)
	if err != nil {
		return fmt.Errorf("差分集計に失敗しました: %w", err)
	}

	if len(aggregated) == 0 {
		u.logger.Info("差分集計結果が0件でした",
			slog.String("resolution", resolution.String()))
		return nil
	}

	// 3. 集計結果をClusterエンティティに変換
	clusters, err := infraRepo.ConvertAggregatedToClusters(resolution, aggregated)
	if err != nil {
		return fmt.Errorf("変換に失敗しました: %w", err)
	}

	// 4. cluster_resultsテーブルに保存(UPSERT)
	if err := u.clusterRepo.SaveClusters(ctx, clusters); err != nil {
		return fmt.Errorf("保存に失敗しました: %w", err)
	}

	u.logger.Info("解像度別の差分クラスター計算が完了しました",
		slog.String("resolution", resolution.String()),
		slog.Int("cluster_count", len(clusters)))

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
