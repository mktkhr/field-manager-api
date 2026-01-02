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
	Priority        int32    // 優先度（高いほど先に処理）
	AffectedH3Cells []string // 影響を受けたH3セル(nil=全範囲再計算)
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
	var job *entity.ClusterJob
	if len(input.AffectedH3Cells) > 0 {
		// 差分更新用ジョブ
		job = entity.NewClusterJobWithAffectedCells(input.Priority, input.AffectedH3Cells)
		if err := u.jobRepo.CreateWithAffectedCells(ctx, job); err != nil {
			return EnqueueJobOutput{}, err
		}
		u.logger.Info("差分更新用クラスタージョブをエンキューしました",
			slog.String("job_id", job.ID.String()),
			slog.Int("priority", int(input.Priority)),
			slog.Int("affected_cells", len(input.AffectedH3Cells)))
	} else {
		// 全範囲再計算用ジョブ
		job = entity.NewClusterJob(input.Priority)
		if err := u.jobRepo.Create(ctx, job); err != nil {
			return EnqueueJobOutput{}, err
		}
		u.logger.Info("全範囲再計算用クラスタージョブをエンキューしました",
			slog.String("job_id", job.ID.String()),
			slog.Int("priority", int(input.Priority)))
	}

	return EnqueueJobOutput{Enqueued: true}, nil
}

// ClusterJobEnqueuerAdapter はimport機能から使用するアダプタ
// import機能のConsumer側で定義されるClusterJobEnqueuerインターフェースに対応
type ClusterJobEnqueuerAdapter struct {
	usecase *EnqueueJobUseCase
}

// NewClusterJobEnqueuer はClusterJobEnqueuerAdapterを作成する
func NewClusterJobEnqueuer(usecase *EnqueueJobUseCase) *ClusterJobEnqueuerAdapter {
	return &ClusterJobEnqueuerAdapter{
		usecase: usecase,
	}
}

// Enqueue はジョブをエンキューする(全範囲再計算)
func (a *ClusterJobEnqueuerAdapter) Enqueue(ctx context.Context, priority int32) error {
	_, err := a.usecase.Execute(ctx, EnqueueJobInput{
		Priority: priority,
	})
	return err
}

// EnqueueWithAffectedCells は影響セル情報付きでジョブをエンキューする(差分更新)
func (a *ClusterJobEnqueuerAdapter) EnqueueWithAffectedCells(ctx context.Context, priority int32, affectedCells []string) error {
	_, err := a.usecase.Execute(ctx, EnqueueJobInput{
		Priority:        priority,
		AffectedH3Cells: affectedCells,
	})
	return err
}
