// Package repository はクラスタリング機能のリポジトリインターフェースを定義する
package repository

import (
	"context"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
)

// ClusterCacheRepository はクラスター結果のキャッシュリポジトリインターフェース
type ClusterCacheRepository interface {
	// GetClusters はキャッシュから指定解像度のクラスター結果を取得する
	// キャッシュミスの場合はnilを返す
	GetClusters(ctx context.Context, resolution entity.Resolution) ([]*entity.Cluster, error)

	// SetClusters はクラスター結果をキャッシュに保存する
	SetClusters(ctx context.Context, resolution entity.Resolution, clusters []*entity.Cluster) error

	// DeleteClusters は全解像度のクラスター結果をキャッシュから削除する
	DeleteClusters(ctx context.Context) error
}
