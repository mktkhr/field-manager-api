// Package repository はクラスタリング機能のリポジトリ実装を提供する
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
	"github.com/mktkhr/field-manager-api/internal/utils"
)

// clusterPostgresRepository はClusterRepositoryのPostgreSQL実装
type clusterPostgresRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewClusterPostgresRepository はClusterRepositoryのPostgreSQL実装を作成する
func NewClusterPostgresRepository(pool *pgxpool.Pool) repository.ClusterRepository {
	return &clusterPostgresRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// GetClusters は指定解像度のクラスター結果を取得する
func (r *clusterPostgresRepository) GetClusters(ctx context.Context, resolution entity.Resolution) ([]*entity.Cluster, error) {
	results, err := r.queries.GetClusterResults(ctx, utils.SafeIntToInt32(int(resolution)))
	if err != nil {
		return nil, fmt.Errorf("クラスター結果の取得に失敗しました: %w", err)
	}

	clusters := make([]*entity.Cluster, 0, len(results))
	for _, result := range results {
		clusters = append(clusters, &entity.Cluster{
			ID:           result.ID,
			Resolution:   entity.Resolution(result.Resolution),
			H3Index:      result.H3Index,
			FieldCount:   result.FieldCount,
			CenterLat:    result.CenterLat,
			CenterLng:    result.CenterLng,
			CalculatedAt: result.CalculatedAt.Time,
		})
	}

	return clusters, nil
}

// SaveClusters は複数のクラスター結果を保存する
func (r *clusterPostgresRepository) SaveClusters(ctx context.Context, clusters []*entity.Cluster) error {
	if len(clusters) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗しました: %w", err)
	}
	defer func() {
		// コミット後のロールバックは"tx is closed"エラーになるため無視
		_ = tx.Rollback(ctx)
	}()

	queries := r.queries.WithTx(tx)
	for _, cluster := range clusters {
		err := queries.UpsertClusterResult(ctx, &sqlc.UpsertClusterResultParams{
			ID:         cluster.ID,
			Resolution: utils.SafeIntToInt32(int(cluster.Resolution)),
			H3Index:    cluster.H3Index,
			FieldCount: cluster.FieldCount,
			CenterLat:  cluster.CenterLat,
			CenterLng:  cluster.CenterLng,
		})
		if err != nil {
			return fmt.Errorf("クラスター結果の保存に失敗しました (H3Index: %s): %w", cluster.H3Index, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("トランザクションコミットに失敗しました: %w", err)
	}
	return nil
}

// DeleteClustersByResolution は指定解像度のクラスター結果を削除する
func (r *clusterPostgresRepository) DeleteClustersByResolution(ctx context.Context, resolution entity.Resolution) error {
	if err := r.queries.DeleteClusterResultsByResolution(ctx, utils.SafeIntToInt32(int(resolution))); err != nil {
		return fmt.Errorf("解像度%dのクラスター結果の削除に失敗しました: %w", resolution, err)
	}
	return nil
}

// DeleteAllClusters は全てのクラスター結果を削除する
func (r *clusterPostgresRepository) DeleteAllClusters(ctx context.Context) error {
	if err := r.queries.DeleteAllClusterResults(ctx); err != nil {
		return fmt.Errorf("クラスター結果の全削除に失敗しました: %w", err)
	}
	return nil
}

// AggregateByH3 は指定解像度でfieldsテーブルを集計する(全範囲)
func (r *clusterPostgresRepository) AggregateByH3(ctx context.Context, resolution entity.Resolution) ([]*repository.AggregatedCluster, error) {
	switch resolution {
	case entity.Res3:
		return r.aggregateRes3(ctx)
	case entity.Res5:
		return r.aggregateRes5(ctx)
	case entity.Res7:
		return r.aggregateRes7(ctx)
	case entity.Res9:
		return r.aggregateRes9(ctx)
	default:
		return nil, fmt.Errorf("未対応の解像度です: %d", resolution)
	}
}

// AggregateByH3ForCells は指定H3セルのみfieldsテーブルを集計する(差分更新用)
func (r *clusterPostgresRepository) AggregateByH3ForCells(ctx context.Context, resolution entity.Resolution, h3Cells []string) ([]*repository.AggregatedCluster, error) {
	switch resolution {
	case entity.Res3:
		return r.aggregateRes3ForCells(ctx, h3Cells)
	case entity.Res5:
		return r.aggregateRes5ForCells(ctx, h3Cells)
	case entity.Res7:
		return r.aggregateRes7ForCells(ctx, h3Cells)
	case entity.Res9:
		return r.aggregateRes9ForCells(ctx, h3Cells)
	default:
		return nil, fmt.Errorf("未対応の解像度です: %d", resolution)
	}
}

// DeleteClustersByH3Indexes は指定H3インデックスのクラスター結果を削除する
func (r *clusterPostgresRepository) DeleteClustersByH3Indexes(ctx context.Context, resolution entity.Resolution, h3Indexes []string) error {
	if len(h3Indexes) == 0 {
		return nil
	}
	if err := r.queries.DeleteClusterResultsByH3Indexes(ctx, &sqlc.DeleteClusterResultsByH3IndexesParams{
		Resolution: utils.SafeIntToInt32(int(resolution)),
		H3Indexes:  h3Indexes,
	}); err != nil {
		return fmt.Errorf("H3インデックス指定でのクラスター結果の削除に失敗しました: %w", err)
	}
	return nil
}

func (r *clusterPostgresRepository) aggregateRes3(ctx context.Context) ([]*repository.AggregatedCluster, error) {
	rows, err := r.queries.AggregateClustersByRes3(ctx)
	if err != nil {
		return nil, fmt.Errorf("解像度3での集計に失敗しました: %w", err)
	}

	result := make([]*repository.AggregatedCluster, 0, len(rows))
	for _, row := range rows {
		if row.H3Index == nil {
			continue
		}
		result = append(result, &repository.AggregatedCluster{
			H3Index:    *row.H3Index,
			FieldCount: row.FieldCount,
		})
	}
	return result, nil
}

func (r *clusterPostgresRepository) aggregateRes5(ctx context.Context) ([]*repository.AggregatedCluster, error) {
	rows, err := r.queries.AggregateClustersByRes5(ctx)
	if err != nil {
		return nil, fmt.Errorf("解像度5での集計に失敗しました: %w", err)
	}

	result := make([]*repository.AggregatedCluster, 0, len(rows))
	for _, row := range rows {
		if row.H3Index == nil {
			continue
		}
		result = append(result, &repository.AggregatedCluster{
			H3Index:    *row.H3Index,
			FieldCount: row.FieldCount,
		})
	}
	return result, nil
}

func (r *clusterPostgresRepository) aggregateRes7(ctx context.Context) ([]*repository.AggregatedCluster, error) {
	rows, err := r.queries.AggregateClustersByRes7(ctx)
	if err != nil {
		return nil, fmt.Errorf("解像度7での集計に失敗しました: %w", err)
	}

	result := make([]*repository.AggregatedCluster, 0, len(rows))
	for _, row := range rows {
		if row.H3Index == nil {
			continue
		}
		result = append(result, &repository.AggregatedCluster{
			H3Index:    *row.H3Index,
			FieldCount: row.FieldCount,
		})
	}
	return result, nil
}

func (r *clusterPostgresRepository) aggregateRes9(ctx context.Context) ([]*repository.AggregatedCluster, error) {
	rows, err := r.queries.AggregateClustersByRes9(ctx)
	if err != nil {
		return nil, fmt.Errorf("解像度9での集計に失敗しました: %w", err)
	}

	result := make([]*repository.AggregatedCluster, 0, len(rows))
	for _, row := range rows {
		if row.H3Index == nil {
			continue
		}
		result = append(result, &repository.AggregatedCluster{
			H3Index:    *row.H3Index,
			FieldCount: row.FieldCount,
		})
	}
	return result, nil
}

func (r *clusterPostgresRepository) aggregateRes3ForCells(ctx context.Context, h3Cells []string) ([]*repository.AggregatedCluster, error) {
	rows, err := r.queries.AggregateClustersByRes3ForCells(ctx, h3Cells)
	if err != nil {
		return nil, fmt.Errorf("解像度3での差分集計に失敗しました: %w", err)
	}

	result := make([]*repository.AggregatedCluster, 0, len(rows))
	for _, row := range rows {
		if row.H3Index == nil {
			continue
		}
		result = append(result, &repository.AggregatedCluster{
			H3Index:    *row.H3Index,
			FieldCount: row.FieldCount,
		})
	}
	return result, nil
}

func (r *clusterPostgresRepository) aggregateRes5ForCells(ctx context.Context, h3Cells []string) ([]*repository.AggregatedCluster, error) {
	rows, err := r.queries.AggregateClustersByRes5ForCells(ctx, h3Cells)
	if err != nil {
		return nil, fmt.Errorf("解像度5での差分集計に失敗しました: %w", err)
	}

	result := make([]*repository.AggregatedCluster, 0, len(rows))
	for _, row := range rows {
		if row.H3Index == nil {
			continue
		}
		result = append(result, &repository.AggregatedCluster{
			H3Index:    *row.H3Index,
			FieldCount: row.FieldCount,
		})
	}
	return result, nil
}

func (r *clusterPostgresRepository) aggregateRes7ForCells(ctx context.Context, h3Cells []string) ([]*repository.AggregatedCluster, error) {
	rows, err := r.queries.AggregateClustersByRes7ForCells(ctx, h3Cells)
	if err != nil {
		return nil, fmt.Errorf("解像度7での差分集計に失敗しました: %w", err)
	}

	result := make([]*repository.AggregatedCluster, 0, len(rows))
	for _, row := range rows {
		if row.H3Index == nil {
			continue
		}
		result = append(result, &repository.AggregatedCluster{
			H3Index:    *row.H3Index,
			FieldCount: row.FieldCount,
		})
	}
	return result, nil
}

func (r *clusterPostgresRepository) aggregateRes9ForCells(ctx context.Context, h3Cells []string) ([]*repository.AggregatedCluster, error) {
	rows, err := r.queries.AggregateClustersByRes9ForCells(ctx, h3Cells)
	if err != nil {
		return nil, fmt.Errorf("解像度9での差分集計に失敗しました: %w", err)
	}

	result := make([]*repository.AggregatedCluster, 0, len(rows))
	for _, row := range rows {
		if row.H3Index == nil {
			continue
		}
		result = append(result, &repository.AggregatedCluster{
			H3Index:    *row.H3Index,
			FieldCount: row.FieldCount,
		})
	}
	return result, nil
}
