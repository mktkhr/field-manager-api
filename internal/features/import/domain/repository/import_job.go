package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

// ImportJobRepository はインポートジョブのリポジトリインターフェース
type ImportJobRepository interface {
	// FindByID はIDでインポートジョブを取得する
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ImportJob, error)

	// Create はインポートジョブを作成する
	Create(ctx context.Context, job *entity.ImportJob) error

	// UpdateStatus はステータスを更新する
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ImportStatus) error

	// UpdateProgress は進捗を更新する
	UpdateProgress(ctx context.Context, id uuid.UUID, processed, failed, batch int32) error

	// UpdateS3Key はS3キーを更新する
	UpdateS3Key(ctx context.Context, id uuid.UUID, s3Key string) error

	// UpdateExecutionArn は実行ARNを更新する
	UpdateExecutionArn(ctx context.Context, id uuid.UUID, arn string) error

	// UpdateTotalRecords は総レコード数を更新する
	UpdateTotalRecords(ctx context.Context, id uuid.UUID, total int32) error

	// UpdateError はエラー情報を更新する
	UpdateError(ctx context.Context, id uuid.UUID, message string, failedIDs []string) error
}
