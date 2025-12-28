package query

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	appQuery "github.com/mktkhr/field-manager-api/internal/features/import/application/query"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
)

// importJobQuery はImportJobQueryの実装
type importJobQuery struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewImportJobQuery は新しいImportJobQueryを作成する
func NewImportJobQuery(db *pgxpool.Pool) appQuery.ImportJobQuery {
	return &importJobQuery{
		db:      db,
		queries: sqlc.New(db),
	}
}

// FindByID はIDでインポートジョブを取得する
func (q *importJobQuery) FindByID(ctx context.Context, id uuid.UUID) (*entity.ImportJob, error) {
	row, err := q.queries.GetImportJob(ctx, id)
	if err != nil {
		return nil, err
	}
	return q.toEntity(row), nil
}

// List はインポートジョブ一覧を取得する
func (q *importJobQuery) List(ctx context.Context, limit, offset int32) ([]*entity.ImportJob, error) {
	rows, err := q.queries.ListImportJobs(ctx, &sqlc.ListImportJobsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	jobs := make([]*entity.ImportJob, len(rows))
	for i, row := range rows {
		jobs[i] = q.toEntity(row)
	}
	return jobs, nil
}

// ListByCityCode は市区町村コードでインポートジョブ一覧を取得する
func (q *importJobQuery) ListByCityCode(ctx context.Context, cityCode string, limit, offset int32) ([]*entity.ImportJob, error) {
	rows, err := q.queries.ListImportJobsByCityCode(ctx, &sqlc.ListImportJobsByCityCodeParams{
		CityCode: cityCode,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	jobs := make([]*entity.ImportJob, len(rows))
	for i, row := range rows {
		jobs[i] = q.toEntity(row)
	}
	return jobs, nil
}

// Count はインポートジョブの総数を取得する
func (q *importJobQuery) Count(ctx context.Context) (int64, error) {
	return q.queries.CountImportJobs(ctx)
}

// CountByStatus はステータス別のインポートジョブ数を取得する
func (q *importJobQuery) CountByStatus(ctx context.Context, status entity.ImportStatus) (int64, error) {
	return q.queries.CountImportJobsByStatus(ctx, string(status))
}

// toEntity はSQLCモデルをエンティティに変換する
func (q *importJobQuery) toEntity(row *sqlc.ImportJob) *entity.ImportJob {
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
		if err := json.Unmarshal(row.FailedRecordIds, &ids); err == nil {
			job.FailedRecordIDs = ids
		}
	}

	return job
}
