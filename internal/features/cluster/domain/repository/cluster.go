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

	// AggregateByH3 は指定解像度でfieldsテーブルを集計する
	AggregateByH3(ctx context.Context, resolution entity.Resolution) ([]*AggregatedCluster, error)
}
