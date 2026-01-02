// Package repository はクラスタリング機能のリポジトリインターフェースを定義する
package repository

import (
	"context"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
)

// AggregatedCluster は集計されたクラスター情報
type AggregatedCluster struct {
	H3Index    string
	FieldCount int32
}

// ClusterRepository はクラスター結果のリポジトリインターフェース
type ClusterRepository interface {
	// GetClusters は指定解像度のクラスター結果を取得する
	GetClusters(ctx context.Context, resolution entity.Resolution) ([]*entity.Cluster, error)

	// SaveClusters は複数のクラスター結果を保存する
	SaveClusters(ctx context.Context, clusters []*entity.Cluster) error

	// DeleteClustersByResolution は指定解像度のクラスター結果を削除する
	DeleteClustersByResolution(ctx context.Context, resolution entity.Resolution) error

	// DeleteAllClusters は全てのクラスター結果を削除する
	DeleteAllClusters(ctx context.Context) error

	// AggregateByH3 は指定解像度でfieldsテーブルを集計する(全範囲)
	AggregateByH3(ctx context.Context, resolution entity.Resolution) ([]*AggregatedCluster, error)

	// AggregateByH3ForCells は指定H3セルのみfieldsテーブルを集計する(差分更新用)
	AggregateByH3ForCells(ctx context.Context, resolution entity.Resolution, h3Cells []string) ([]*AggregatedCluster, error)

	// DeleteClustersByH3Indexes は指定H3インデックスのクラスター結果を削除する
	DeleteClustersByH3Indexes(ctx context.Context, resolution entity.Resolution, h3Indexes []string) error
}
