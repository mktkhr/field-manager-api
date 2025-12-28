package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/apperror"
	"github.com/mktkhr/field-manager-api/internal/features/import/application/query"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

// GetImportStatusOutput はインポートステータス取得の出力
type GetImportStatusOutput struct {
	ID               uuid.UUID
	CityCode         string
	Status           entity.ImportStatus
	TotalRecords     *int32
	ProcessedRecords int32
	FailedRecords    int32
	Progress         float64
	ErrorMessage     *string
	CreatedAt        string
	StartedAt        *string
	CompletedAt      *string
}

// GetImportStatusUseCase はインポートステータス取得のユースケース
type GetImportStatusUseCase struct {
	importJobQuery query.ImportJobQuery
}

// NewGetImportStatusUseCase は新しいGetImportStatusUseCaseを作成する
func NewGetImportStatusUseCase(importJobQuery query.ImportJobQuery) *GetImportStatusUseCase {
	return &GetImportStatusUseCase{
		importJobQuery: importJobQuery,
	}
}

// Execute はインポートステータスを取得する
func (uc *GetImportStatusUseCase) Execute(ctx context.Context, id uuid.UUID) (*GetImportStatusOutput, error) {
	job, err := uc.importJobQuery.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.InternalErrorWithCause("インポートジョブの取得に失敗しました", err)
	}

	if job == nil {
		return nil, apperror.NotFoundError("インポートジョブが見つかりません")
	}

	output := &GetImportStatusOutput{
		ID:               job.ID,
		CityCode:         job.CityCode,
		Status:           job.Status,
		TotalRecords:     job.TotalRecords,
		ProcessedRecords: job.ProcessedRecords,
		FailedRecords:    job.FailedRecords,
		Progress:         job.Progress(),
		ErrorMessage:     job.ErrorMessage,
		CreatedAt:        job.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if job.StartedAt != nil {
		s := job.StartedAt.Format("2006-01-02T15:04:05Z07:00")
		output.StartedAt = &s
	}

	if job.CompletedAt != nil {
		s := job.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		output.CompletedAt = &s
	}

	return output, nil
}
