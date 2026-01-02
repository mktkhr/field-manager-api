// Package usecase はクラスタリング機能のユースケースを提供する
package usecase

import (
	"context"
	"log/slog"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/internal/h3util"
)

// GetClustersInput はクラスター取得ユースケースの入力
type GetClustersInput struct {
	Zoom  float64 // ズームレベル(1.0-22.0)
	SWLat float64 // 南西端の緯度
	SWLng float64 // 南西端の経度
	NELat float64 // 北東端の緯度
	NELng float64 // 北東端の経度
}

// GetClustersOutput はクラスター取得ユースケースの出力
type GetClustersOutput struct {
	Clusters []*entity.ClusterResult
	IsStale  bool
}

// GetClustersUseCase はクラスター取得ユースケース
type GetClustersUseCase struct {
	clusterRepo repository.ClusterRepository
	cacheRepo   repository.ClusterCacheRepository
	jobRepo     repository.ClusterJobRepository
	logger      *slog.Logger
}

// NewGetClustersUseCase はGetClustersUseCaseを作成する
func NewGetClustersUseCase(
	clusterRepo repository.ClusterRepository,
	cacheRepo repository.ClusterCacheRepository,
	jobRepo repository.ClusterJobRepository,
	logger *slog.Logger,
) *GetClustersUseCase {
	return &GetClustersUseCase{
		clusterRepo: clusterRepo,
		cacheRepo:   cacheRepo,
		jobRepo:     jobRepo,
		logger:      logger,
	}
}

// Execute はクラスター取得を実行する
func (u *GetClustersUseCase) Execute(ctx context.Context, input GetClustersInput) (*GetClustersOutput, error) {
	// ズームレベルからH3解像度を決定
	resolution := h3util.ZoomToResolution(input.Zoom)

	// バウンディングボックスを作成
	bbox := h3util.NewBoundingBox(input.SWLat, input.SWLng, input.NELat, input.NELng)

	// キャッシュから取得を試みる
	clusters, err := u.cacheRepo.GetClusters(ctx, resolution)
	if err != nil {
		// キャッシュエラーはログに残して続行
		u.logger.Warn("キャッシュからの取得に失敗しました",
			slog.String("error", err.Error()),
			slog.String("resolution", resolution.String()))
	}

	// キャッシュミスの場合はDBから取得
	if clusters == nil {
		clusters, err = u.clusterRepo.GetClusters(ctx, resolution)
		if err != nil {
			return nil, err
		}

		// 取得したデータをキャッシュに保存
		if len(clusters) > 0 {
			if cacheErr := u.cacheRepo.SetClusters(ctx, resolution, clusters); cacheErr != nil {
				u.logger.Warn("キャッシュへの保存に失敗しました",
					slog.String("error", cacheErr.Error()),
					slog.String("resolution", resolution.String()))
			}
		}
	}

	// BoundingBox内のクラスターをフィルタリング
	filteredClusters := u.filterClustersByBBox(clusters, bbox)

	// staleフラグを確認(再計算中のジョブがあるか)
	isStale, err := u.jobRepo.HasPendingOrProcessingJob(ctx)
	if err != nil {
		// エラーの場合は安全側に倒してtrueとして扱う
		// (クライアントに「再計算中かもしれない」と伝える)
		u.logger.Warn("staleフラグの確認に失敗しました",
			slog.String("error", err.Error()))
		isStale = true
	}

	// レスポンス用に変換
	results := make([]*entity.ClusterResult, 0, len(filteredClusters))
	for _, cluster := range filteredClusters {
		results = append(results, cluster.ToResult())
	}

	return &GetClustersOutput{
		Clusters: results,
		IsStale:  isStale,
	}, nil
}

// filterClustersByBBox はBoundingBox内のクラスターをフィルタリングする
func (u *GetClustersUseCase) filterClustersByBBox(clusters []*entity.Cluster, bbox *h3util.BoundingBox) []*entity.Cluster {
	if bbox == nil || !bbox.IsValid() {
		return clusters
	}

	filtered := make([]*entity.Cluster, 0, len(clusters))
	for _, cluster := range clusters {
		if bbox.Contains(cluster.CenterLat, cluster.CenterLng) {
			filtered = append(filtered, cluster)
		}
	}

	return filtered
}
