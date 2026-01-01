package repository

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/repository"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
)

// jsonMarshal はテスト時にモック可能なJSON Marshal関数
var jsonMarshal = json.Marshal

// importJobRepository はImportJobRepositoryの実装
type importJobRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewImportJobRepository は新しいImportJobRepositoryを作成する
func NewImportJobRepository(db *pgxpool.Pool) repository.ImportJobRepository {
	return &importJobRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// FindByID はIDでインポートジョブを取得する
func (r *importJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ImportJob, error) {
	row, err := r.queries.GetImportJob(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toEntity(row), nil
}

// Create はインポートジョブを作成する
func (r *importJobRepository) Create(ctx context.Context, job *entity.ImportJob) error {
	row, err := r.queries.CreateImportJob(ctx, job.CityCode)
	if err != nil {
		return err
	}
	job.ID = row.ID
	job.Status = entity.ImportStatus(row.Status)
	if row.CreatedAt.Valid {
		job.CreatedAt = row.CreatedAt.Time
	}
	return nil
}

// UpdateStatus はステータスを更新する
func (r *importJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ImportStatus) error {
	_, err := r.queries.UpdateImportJobStatus(ctx, &sqlc.UpdateImportJobStatusParams{
		ID:     id,
		Status: string(status),
	})
	return err
}

// UpdateProgress は進捗を更新する
func (r *importJobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, processed, failed, batch int32) error {
	_, err := r.queries.UpdateImportJobProgress(ctx, &sqlc.UpdateImportJobProgressParams{
		ID:                 id,
		ProcessedRecords:   processed,
		FailedRecords:      failed,
		LastProcessedBatch: batch,
	})
	return err
}

// UpdateS3Key はS3キーを更新する
func (r *importJobRepository) UpdateS3Key(ctx context.Context, id uuid.UUID, s3Key string) error {
	_, err := r.queries.UpdateImportJobS3Key(ctx, &sqlc.UpdateImportJobS3KeyParams{
		ID:    id,
		S3Key: &s3Key,
	})
	return err
}

// UpdateExecutionArn は実行ARNを更新する
func (r *importJobRepository) UpdateExecutionArn(ctx context.Context, id uuid.UUID, arn string) error {
	_, err := r.queries.UpdateImportJobExecutionArn(ctx, &sqlc.UpdateImportJobExecutionArnParams{
		ID:           id,
		ExecutionArn: &arn,
	})
	return err
}

// UpdateTotalRecords は総レコード数を更新する
func (r *importJobRepository) UpdateTotalRecords(ctx context.Context, id uuid.UUID, total int32) error {
	_, err := r.queries.UpdateImportJobTotalRecords(ctx, &sqlc.UpdateImportJobTotalRecordsParams{
		ID:           id,
		TotalRecords: &total,
	})
	return err
}

// UpdateError はエラー情報を更新する
func (r *importJobRepository) UpdateError(ctx context.Context, id uuid.UUID, message string, failedIDs []string) error {
	var failedIDsJSON json.RawMessage
	if len(failedIDs) > 0 {
		data, err := jsonMarshal(failedIDs)
		if err != nil {
			return err
		}
		failedIDsJSON = data
	}

	_, err := r.queries.UpdateImportJobError(ctx, &sqlc.UpdateImportJobErrorParams{
		ID:              id,
		ErrorMessage:    &message,
		FailedRecordIds: failedIDsJSON,
	})
	return err
}

// toEntity はSQLCモデルをエンティティに変換する
func (r *importJobRepository) toEntity(row *sqlc.ImportJob) *entity.ImportJob {
	if row == nil {
		return nil
	}

	job := &entity.ImportJob{
		ID:                 row.ID,
		CityCode:           row.CityCode,
		Status:             entity.ImportStatus(row.Status),
		TotalRecords:       row.TotalRecords,
		ProcessedRecords:   row.ProcessedRecords,
		FailedRecords:      row.FailedRecords,
		LastProcessedBatch: row.LastProcessedBatch,
		S3Key:              row.S3Key,
		ExecutionArn:       row.ExecutionArn,
		ErrorMessage:       row.ErrorMessage,
	}

	if row.CreatedAt.Valid {
		job.CreatedAt = row.CreatedAt.Time
	}
	if row.StartedAt.Valid {
		job.StartedAt = &row.StartedAt.Time
	}
	if row.CompletedAt.Valid {
		job.CompletedAt = &row.CompletedAt.Time
	}

	if len(row.FailedRecordIds) > 0 {
		var ids []string
		if err := json.Unmarshal(row.FailedRecordIds, &ids); err != nil {
			// DBに保存されたJSONは通常有効なはずだが、デバッグのためにログ出力
			slog.Warn("失敗レコードIDのパースに失敗", "job_id", row.ID, "error", err)
		} else {
			job.FailedRecordIDs = ids
		}
	}

	return job
}
