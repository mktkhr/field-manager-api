// Package repository はクラスタリング機能のリポジトリ実装を提供する
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
)

// clusterJobPostgresRepository はClusterJobRepositoryのPostgreSQL実装
type clusterJobPostgresRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewClusterJobPostgresRepository はClusterJobRepositoryのPostgreSQL実装を作成する
func NewClusterJobPostgresRepository(pool *pgxpool.Pool) repository.ClusterJobRepository {
	return &clusterJobPostgresRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Create は新しいクラスタージョブを作成する
func (r *clusterJobPostgresRepository) Create(ctx context.Context, job *entity.ClusterJob) error {
	_, err := r.queries.CreateClusterJob(ctx, &sqlc.CreateClusterJobParams{
		ID:       job.ID,
		Priority: job.Priority,
	})
	if err != nil {
		return fmt.Errorf("クラスタージョブの作成に失敗しました: %w", err)
	}
	return nil
}

// FindByID はIDでクラスタージョブを取得する
func (r *clusterJobPostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ClusterJob, error) {
	result, err := r.queries.GetClusterJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("クラスタージョブの取得に失敗しました: %w", err)
	}

	return convertToClusterJobEntity(result), nil
}

// FindPendingJobs は保留中のジョブを優先度順に取得する(排他ロック付き)
func (r *clusterJobPostgresRepository) FindPendingJobs(ctx context.Context, limit int32) ([]*entity.ClusterJob, error) {
	results, err := r.queries.GetPendingClusterJobs(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("保留中ジョブの取得に失敗しました: %w", err)
	}

	jobs := make([]*entity.ClusterJob, 0, len(results))
	for _, result := range results {
		jobs = append(jobs, convertToClusterJobEntity(result))
	}

	return jobs, nil
}

// UpdateToProcessing はジョブを処理中に更新する
func (r *clusterJobPostgresRepository) UpdateToProcessing(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.UpdateClusterJobToProcessing(ctx, id); err != nil {
		return fmt.Errorf("ジョブの処理中への更新に失敗しました: %w", err)
	}
	return nil
}

// UpdateToCompleted はジョブを完了に更新する
func (r *clusterJobPostgresRepository) UpdateToCompleted(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.UpdateClusterJobToCompleted(ctx, id); err != nil {
		return fmt.Errorf("ジョブの完了への更新に失敗しました: %w", err)
	}
	return nil
}

// UpdateToFailed はジョブを失敗に更新する
func (r *clusterJobPostgresRepository) UpdateToFailed(ctx context.Context, id uuid.UUID, errorMessage string) error {
	if err := r.queries.UpdateClusterJobToFailed(ctx, &sqlc.UpdateClusterJobToFailedParams{
		ID:           id,
		ErrorMessage: &errorMessage,
	}); err != nil {
		return fmt.Errorf("ジョブの失敗への更新に失敗しました: %w", err)
	}
	return nil
}

// HasPendingOrProcessingJob は保留中または処理中のジョブがあるか確認する
func (r *clusterJobPostgresRepository) HasPendingOrProcessingJob(ctx context.Context) (bool, error) {
	hasJob, err := r.queries.HasPendingOrProcessingJob(ctx)
	if err != nil {
		return false, fmt.Errorf("ジョブ存在確認に失敗しました: %w", err)
	}
	return hasJob, nil
}

// DeleteOldCompletedJobs は古い完了済みジョブを削除する
func (r *clusterJobPostgresRepository) DeleteOldCompletedJobs(ctx context.Context) error {
	if err := r.queries.DeleteOldCompletedJobs(ctx); err != nil {
		return fmt.Errorf("古い完了済みジョブの削除に失敗しました: %w", err)
	}
	return nil
}

// DeleteOldFailedJobs は古い失敗ジョブを削除する
func (r *clusterJobPostgresRepository) DeleteOldFailedJobs(ctx context.Context) error {
	if err := r.queries.DeleteOldFailedJobs(ctx); err != nil {
		return fmt.Errorf("古い失敗ジョブの削除に失敗しました: %w", err)
	}
	return nil
}

// convertToClusterJobEntity はSQLCの結果をエンティティに変換する
func convertToClusterJobEntity(job *sqlc.ClusterJob) *entity.ClusterJob {
	result := &entity.ClusterJob{
		ID:        job.ID,
		Status:    entity.JobStatus(job.Status),
		Priority:  job.Priority,
		CreatedAt: job.CreatedAt.Time,
	}

	if job.StartedAt.Valid {
		t := job.StartedAt.Time
		result.StartedAt = &t
	}

	if job.CompletedAt.Valid {
		t := job.CompletedAt.Time
		result.CompletedAt = &t
	}

	if job.ErrorMessage != nil {
		result.ErrorMessage = *job.ErrorMessage
	}

	return result
}
