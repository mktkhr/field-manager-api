// Package repository はクラスタリング機能のリポジトリ実装を提供する
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/cache"
	"github.com/redis/go-redis/v9"
)

const (
	// clusterCacheKeyPrefix はクラスターキャッシュのキープレフィックス
	clusterCacheKeyPrefix = "cluster:results:"

	// clusterCacheTTL はクラスターキャッシュのTTL
	clusterCacheTTL = 30 * time.Minute
)

// clusterCacheData はキャッシュに保存するクラスターデータ
type clusterCacheData struct {
	H3Index      string  `json:"h3_index"`
	FieldCount   int32   `json:"field_count"`
	CenterLat    float64 `json:"center_lat"`
	CenterLng    float64 `json:"center_lng"`
	CalculatedAt int64   `json:"calculated_at"` // Unix timestamp
}

// clusterCacheRedisRepository はClusterCacheRepositoryのRedis実装
type clusterCacheRedisRepository struct {
	client *cache.Client
}

// NewClusterCacheRedisRepository はClusterCacheRepositoryのRedis実装を作成する
func NewClusterCacheRedisRepository(client *cache.Client) repository.ClusterCacheRepository {
	return &clusterCacheRedisRepository{
		client: client,
	}
}

// buildCacheKey はキャッシュキーを構築する
func buildCacheKey(resolution entity.Resolution) string {
	return fmt.Sprintf("%s%s", clusterCacheKeyPrefix, resolution.String())
}

// GetClusters はキャッシュから指定解像度のクラスター結果を取得する
func (r *clusterCacheRedisRepository) GetClusters(ctx context.Context, resolution entity.Resolution) ([]*entity.Cluster, error) {
	key := buildCacheKey(resolution)

	data, err := r.client.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			// キャッシュミス
			return nil, nil
		}
		return nil, fmt.Errorf("キャッシュからの取得に失敗しました: %w", err)
	}

	var cacheItems []clusterCacheData
	if err := json.Unmarshal([]byte(data), &cacheItems); err != nil {
		// 不正なデータの場合はキャッシュミスとして扱う
		return nil, nil
	}

	clusters := make([]*entity.Cluster, 0, len(cacheItems))
	for _, item := range cacheItems {
		clusters = append(clusters, &entity.Cluster{
			Resolution:   resolution,
			H3Index:      item.H3Index,
			FieldCount:   item.FieldCount,
			CenterLat:    item.CenterLat,
			CenterLng:    item.CenterLng,
			CalculatedAt: time.Unix(item.CalculatedAt, 0),
		})
	}

	return clusters, nil
}

// SetClusters はクラスター結果をキャッシュに保存する
func (r *clusterCacheRedisRepository) SetClusters(ctx context.Context, resolution entity.Resolution, clusters []*entity.Cluster) error {
	if len(clusters) == 0 {
		return nil
	}

	cacheItems := make([]clusterCacheData, 0, len(clusters))
	for _, cluster := range clusters {
		cacheItems = append(cacheItems, clusterCacheData{
			H3Index:      cluster.H3Index,
			FieldCount:   cluster.FieldCount,
			CenterLat:    cluster.CenterLat,
			CenterLng:    cluster.CenterLng,
			CalculatedAt: cluster.CalculatedAt.Unix(),
		})
	}

	data, err := json.Marshal(cacheItems)
	if err != nil {
		return fmt.Errorf("キャッシュデータのシリアライズに失敗しました: %w", err)
	}

	key := buildCacheKey(resolution)
	if err := r.client.Set(ctx, key, string(data), clusterCacheTTL); err != nil {
		return fmt.Errorf("キャッシュへの保存に失敗しました: %w", err)
	}

	return nil
}

// DeleteClusters は全解像度のクラスター結果をキャッシュから削除する
func (r *clusterCacheRedisRepository) DeleteClusters(ctx context.Context) error {
	for _, resolution := range entity.AllResolutions {
		key := buildCacheKey(resolution)
		if err := r.client.Delete(ctx, key); err != nil {
			// 削除失敗はログに残すが、処理は継続する
			continue
		}
	}
	return nil
}
