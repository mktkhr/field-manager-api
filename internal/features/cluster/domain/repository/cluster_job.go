// Package repository はクラスタリング機能のリポジトリインターフェースを定義する
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
)

// ClusterJobRepository はクラスタージョブのリポジトリインターフェース
type ClusterJobRepository interface {
	// Create は新しいクラスタージョブを作成する(全範囲再計算)
	Create(ctx context.Context, job *entity.ClusterJob) error

	// CreateWithAffectedCells は影響セル情報付きでクラスタージョブを作成する
	CreateWithAffectedCells(ctx context.Context, job *entity.ClusterJob) error

	// FindByID はIDでクラスタージョブを取得する
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ClusterJob, error)

	// FindPendingJobs は保留中のジョブを優先度順に取得する(排他ロック付き)
	FindPendingJobs(ctx context.Context, limit int32) ([]*entity.ClusterJob, error)

	// FindPendingJobsWithAffectedCells は影響セル情報付きで保留中のジョブを取得する
	FindPendingJobsWithAffectedCells(ctx context.Context, limit int32) ([]*entity.ClusterJob, error)

	// UpdateToProcessing はジョブを処理中に更新する
	UpdateToProcessing(ctx context.Context, id uuid.UUID) error

	// UpdateToCompleted はジョブを完了に更新する
	UpdateToCompleted(ctx context.Context, id uuid.UUID) error

	// UpdateToFailed はジョブを失敗に更新する
	UpdateToFailed(ctx context.Context, id uuid.UUID, errorMessage string) error

	// HasPendingOrProcessingJob は保留中または処理中のジョブがあるか確認する
	HasPendingOrProcessingJob(ctx context.Context) (bool, error)

	// DeleteOldCompletedJobs は古い完了済みジョブを削除する
	DeleteOldCompletedJobs(ctx context.Context) error

	// DeleteOldFailedJobs は古い失敗ジョブを削除する
	DeleteOldFailedJobs(ctx context.Context) error
}
