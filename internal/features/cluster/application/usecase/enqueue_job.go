// Package usecase はクラスタリング機能のユースケースを提供する
package usecase

import (
	"context"
	"log/slog"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
)

// EnqueueJobInput はジョブエンキューユースケースの入力
type EnqueueJobInput struct {
	Priority int32 // 優先度（高いほど先に処理）
}

// EnqueueJobOutput はジョブエンキューユースケースの出力
type EnqueueJobOutput struct {
	Enqueued bool // ジョブがエンキューされたかどうか
}

// EnqueueJobUseCase はジョブエンキューユースケース
type EnqueueJobUseCase struct {
	jobRepo repository.ClusterJobRepository
	logger  *slog.Logger
}

// NewEnqueueJobUseCase はEnqueueJobUseCaseを作成する
func NewEnqueueJobUseCase(
	jobRepo repository.ClusterJobRepository,
	logger *slog.Logger,
) *EnqueueJobUseCase {
	return &EnqueueJobUseCase{
		jobRepo: jobRepo,
		logger:  logger,
	}
}

// Execute はジョブエンキューを実行する
// 既に保留中または処理中のジョブがある場合はスキップする
func (u *EnqueueJobUseCase) Execute(ctx context.Context, input EnqueueJobInput) (EnqueueJobOutput, error) {
	// 既にペンディングまたは処理中のジョブがあるか確認
	hasJob, err := u.jobRepo.HasPendingOrProcessingJob(ctx)
	if err != nil {
		return EnqueueJobOutput{}, err
	}

	if hasJob {
		u.logger.Info("既に保留中または処理中のジョブがあるためスキップします")
		return EnqueueJobOutput{Enqueued: false}, nil
	}

	// 新しいジョブを作成
	job := entity.NewClusterJob(input.Priority)

	if err := u.jobRepo.Create(ctx, job); err != nil {
		return EnqueueJobOutput{}, err
	}

	u.logger.Info("クラスタージョブをエンキューしました",
		slog.String("job_id", job.ID.String()),
		slog.Int("priority", int(input.Priority)))

	return EnqueueJobOutput{Enqueued: true}, nil
}

// ClusterJobEnqueuer はimport機能から使用するインターフェース
// import機能のConsumer側で定義されるインターフェースに対応
type ClusterJobEnqueuer interface {
	Enqueue(ctx context.Context, priority int32) error
}

// enqueueJobAdapter はClusterJobEnqueuerの実装
type enqueueJobAdapter struct {
	usecase *EnqueueJobUseCase
}

// NewClusterJobEnqueuer はClusterJobEnqueuerを作成する
func NewClusterJobEnqueuer(usecase *EnqueueJobUseCase) ClusterJobEnqueuer {
	return &enqueueJobAdapter{
		usecase: usecase,
	}
}

// Enqueue はジョブをエンキューする
func (a *enqueueJobAdapter) Enqueue(ctx context.Context, priority int32) error {
	_, err := a.usecase.Execute(ctx, EnqueueJobInput{
		Priority: priority,
	})
	return err
}
