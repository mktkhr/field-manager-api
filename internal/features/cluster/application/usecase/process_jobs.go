// Package usecase はクラスタリング機能のユースケースを提供する
package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
)

// ProcessJobsInput はジョブ処理ユースケースの入力
type ProcessJobsInput struct {
	BatchSize int32 // 1回に処理するジョブ数
}

// ProcessJobsUseCase はジョブ処理ユースケース
type ProcessJobsUseCase struct {
	jobRepo     repository.ClusterJobRepository
	calculateUC *CalculateClustersUseCase
	logger      *slog.Logger
}

// NewProcessJobsUseCase はProcessJobsUseCaseを作成する
func NewProcessJobsUseCase(
	jobRepo repository.ClusterJobRepository,
	calculateUC *CalculateClustersUseCase,
	logger *slog.Logger,
) *ProcessJobsUseCase {
	return &ProcessJobsUseCase{
		jobRepo:     jobRepo,
		calculateUC: calculateUC,
		logger:      logger,
	}
}

// Execute はジョブ処理を実行する
func (u *ProcessJobsUseCase) Execute(ctx context.Context, input ProcessJobsInput) error {
	u.logger.Info("ジョブ処理を開始します",
		slog.Int("batch_size", int(input.BatchSize)))

	// 保留中のジョブを取得（排他ロック付き）
	jobs, err := u.jobRepo.FindPendingJobs(ctx, input.BatchSize)
	if err != nil {
		return fmt.Errorf("ジョブの取得に失敗しました: %w", err)
	}

	if len(jobs) == 0 {
		u.logger.Info("処理対象のジョブがありません")
		return nil
	}

	u.logger.Info("処理対象のジョブを取得しました",
		slog.Int("job_count", len(jobs)))

	// 各ジョブを処理
	for _, job := range jobs {
		u.logger.Info("ジョブの処理を開始します",
			slog.String("job_id", job.ID.String()))

		// ジョブを処理中に更新
		if err := u.jobRepo.UpdateToProcessing(ctx, job.ID); err != nil {
			u.logger.Error("ジョブの処理中への更新に失敗しました",
				slog.String("job_id", job.ID.String()),
				slog.String("error", err.Error()))
			continue
		}

		// クラスター計算を実行
		if err := u.calculateUC.Execute(ctx); err != nil {
			u.logger.Error("クラスター計算に失敗しました",
				slog.String("job_id", job.ID.String()),
				slog.String("error", err.Error()))

			// ジョブを失敗に更新
			if updateErr := u.jobRepo.UpdateToFailed(ctx, job.ID, err.Error()); updateErr != nil {
				u.logger.Error("ジョブの失敗への更新に失敗しました",
					slog.String("job_id", job.ID.String()),
					slog.String("error", updateErr.Error()))
			}
			continue
		}

		// ジョブを完了に更新
		if err := u.jobRepo.UpdateToCompleted(ctx, job.ID); err != nil {
			u.logger.Error("ジョブの完了への更新に失敗しました",
				slog.String("job_id", job.ID.String()),
				slog.String("error", err.Error()))
			continue
		}

		u.logger.Info("ジョブの処理が完了しました",
			slog.String("job_id", job.ID.String()))
	}

	// 古いジョブを削除
	if err := u.jobRepo.DeleteOldCompletedJobs(ctx); err != nil {
		u.logger.Warn("古い完了済みジョブの削除に失敗しました",
			slog.String("error", err.Error()))
	}

	if err := u.jobRepo.DeleteOldFailedJobs(ctx); err != nil {
		u.logger.Warn("古い失敗ジョブの削除に失敗しました",
			slog.String("error", err.Error()))
	}

	u.logger.Info("ジョブ処理が完了しました")
	return nil
}
