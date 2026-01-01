package usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/apperror"
	"github.com/mktkhr/field-manager-api/internal/features/import/application/port"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/repository"
)

// RequestImportInput はインポートリクエストの入力
type RequestImportInput struct {
	CityCode string
}

// RequestImportOutput はインポートリクエストの出力
type RequestImportOutput struct {
	ImportJobID  uuid.UUID
	ExecutionArn string
}

// RequestImportUseCase はインポートリクエストのユースケース
type RequestImportUseCase struct {
	importJobRepo repository.ImportJobRepository
	sfnClient     port.StepFunctionsClient
}

// NewRequestImportUseCase は新しいRequestImportUseCaseを作成する
func NewRequestImportUseCase(
	importJobRepo repository.ImportJobRepository,
	sfnClient port.StepFunctionsClient,
) *RequestImportUseCase {
	return &RequestImportUseCase{
		importJobRepo: importJobRepo,
		sfnClient:     sfnClient,
	}
}

// Execute はインポートリクエストを実行する
func (uc *RequestImportUseCase) Execute(ctx context.Context, input RequestImportInput) (*RequestImportOutput, error) {
	// 1. 市区町村コードのバリデーション
	if input.CityCode == "" {
		return nil, apperror.BadRequestError("市区町村コードは必須です")
	}

	// 2. インポートジョブを作成
	job := entity.NewImportJob(input.CityCode)

	if err := uc.importJobRepo.Create(ctx, job); err != nil {
		return nil, apperror.InternalErrorWithCause("インポートジョブの作成に失敗しました", err)
	}

	// 3. Step Functionsワークフローを開始
	workflowInput := port.WorkflowInput{
		ImportJobID: job.ID,
		CityCode:    input.CityCode,
	}

	execution, err := uc.sfnClient.StartExecution(ctx, workflowInput)
	if err != nil {
		// ワークフロー開始失敗時はジョブを失敗状態に更新
		if updateErr := uc.importJobRepo.UpdateStatus(ctx, job.ID, entity.ImportStatusFailed); updateErr != nil {
			slog.Warn("ジョブステータスの更新に失敗", "job_id", job.ID, "error", updateErr)
		}
		return nil, apperror.InternalErrorWithCause("ワークフローの開始に失敗しました", err)
	}

	// 4. 実行ARNを保存
	if err := uc.importJobRepo.UpdateExecutionArn(ctx, job.ID, execution.ExecutionArn); err != nil {
		return nil, apperror.InternalErrorWithCause("実行ARNの保存に失敗しました", err)
	}

	// 5. ステータスをprocessingに更新
	if err := uc.importJobRepo.UpdateStatus(ctx, job.ID, entity.ImportStatusProcessing); err != nil {
		return nil, apperror.InternalErrorWithCause("ステータスの更新に失敗しました", err)
	}

	return &RequestImportOutput{
		ImportJobID:  job.ID,
		ExecutionArn: execution.ExecutionArn,
	}, nil
}
