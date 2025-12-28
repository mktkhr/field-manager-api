package usecase

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/apperror"
	"github.com/mktkhr/field-manager-api/internal/features/import/application/port"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/repository"
)

const (
	// DefaultBatchSize はデフォルトのバッチサイズ
	DefaultBatchSize = 1000
)

// FieldRepository はField操作用のリポジトリインターフェース(Consumer側で定義)
type FieldRepository interface {
	// UpsertBatch は圃場をバッチでUPSERTする
	UpsertBatch(ctx context.Context, features []entity.WagriFeature) error
}

// ProcessImportInput はインポート処理の入力
type ProcessImportInput struct {
	ImportJobID uuid.UUID
	S3Key       string
	BatchSize   int
}

// ProcessImportUseCase はインポート処理のユースケース
type ProcessImportUseCase struct {
	importJobRepo repository.ImportJobRepository
	storageClient port.StorageClient
	fieldRepo     FieldRepository
	logger        *slog.Logger
}

// NewProcessImportUseCase は新しいProcessImportUseCaseを作成する
func NewProcessImportUseCase(
	importJobRepo repository.ImportJobRepository,
	storageClient port.StorageClient,
	fieldRepo FieldRepository,
	logger *slog.Logger,
) *ProcessImportUseCase {
	return &ProcessImportUseCase{
		importJobRepo: importJobRepo,
		storageClient: storageClient,
		fieldRepo:     fieldRepo,
		logger:        logger,
	}
}

// Execute はインポート処理を実行する
func (uc *ProcessImportUseCase) Execute(ctx context.Context, input ProcessImportInput) error {
	if input.BatchSize <= 0 {
		input.BatchSize = DefaultBatchSize
	}

	uc.logger.Info("インポート処理を開始", "import_job_id", input.ImportJobID, "s3_key", input.S3Key)

	// 1. S3からストリーミング読み取り
	reader, err := uc.storageClient.GetObjectStream(ctx, input.S3Key)
	if err != nil {
		uc.handleError(ctx, input.ImportJobID, "S3からの読み取りに失敗しました", err)
		return apperror.InternalErrorWithCause("S3からの読み取りに失敗しました", err)
	}
	defer reader.Close()

	// 2. JSONをストリーミングパース
	decoder := json.NewDecoder(reader)

	// "targetFeatures"配列の開始を探す
	if err := uc.seekToTargetFeatures(decoder); err != nil {
		uc.handleError(ctx, input.ImportJobID, "JSONのパースに失敗しました", err)
		return apperror.InternalErrorWithCause("JSONのパースに失敗しました", err)
	}

	// 3. バッチ処理
	var (
		batch          []entity.WagriFeature
		processedCount int32
		failedCount    int32
		batchNumber    int32
		failedIDs      []string
	)

	for decoder.More() {
		var feature entity.WagriFeature
		if err := decoder.Decode(&feature); err != nil {
			if err == io.EOF {
				break
			}
			uc.logger.Warn("Featureのパースに失敗", "error", err)
			failedCount++
			continue
		}

		batch = append(batch, feature)

		if len(batch) >= input.BatchSize {
			batchNumber++
			if err := uc.processBatch(ctx, batch); err != nil {
				uc.logger.Error("バッチ処理に失敗", "batch", batchNumber, "error", err)
				for _, f := range batch {
					failedIDs = append(failedIDs, f.Properties.ID)
				}
				failedCount += int32(len(batch))
			} else {
				processedCount += int32(len(batch))
			}

			// 進捗を更新
			if err := uc.importJobRepo.UpdateProgress(ctx, input.ImportJobID, processedCount, failedCount, batchNumber); err != nil {
				uc.logger.Warn("進捗の更新に失敗", "error", err)
			}

			batch = batch[:0]
		}
	}

	// 残りのバッチを処理
	if len(batch) > 0 {
		batchNumber++
		if err := uc.processBatch(ctx, batch); err != nil {
			uc.logger.Error("最終バッチ処理に失敗", "batch", batchNumber, "error", err)
			for _, f := range batch {
				failedIDs = append(failedIDs, f.Properties.ID)
			}
			failedCount += int32(len(batch))
		} else {
			processedCount += int32(len(batch))
		}
	}

	// 4. 最終ステータスを更新
	totalRecords := processedCount + failedCount
	if err := uc.importJobRepo.UpdateTotalRecords(ctx, input.ImportJobID, totalRecords); err != nil {
		uc.logger.Warn("総レコード数の更新に失敗", "error", err)
	}

	if err := uc.importJobRepo.UpdateProgress(ctx, input.ImportJobID, processedCount, failedCount, batchNumber); err != nil {
		uc.logger.Warn("最終進捗の更新に失敗", "error", err)
	}

	// 5. 完了ステータスを設定
	var finalStatus entity.ImportStatus
	if failedCount == 0 {
		finalStatus = entity.ImportStatusCompleted
	} else if processedCount == 0 {
		finalStatus = entity.ImportStatusFailed
	} else {
		finalStatus = entity.ImportStatusPartiallyCompleted
	}

	if finalStatus == entity.ImportStatusFailed {
		if err := uc.importJobRepo.UpdateError(ctx, input.ImportJobID, "一部または全てのレコードの処理に失敗しました", failedIDs); err != nil {
			uc.logger.Warn("エラー情報の更新に失敗", "error", err)
		}
	} else {
		if err := uc.importJobRepo.UpdateStatus(ctx, input.ImportJobID, finalStatus); err != nil {
			uc.logger.Warn("ステータスの更新に失敗", "error", err)
		}
	}

	uc.logger.Info("インポート処理が完了",
		"import_job_id", input.ImportJobID,
		"processed", processedCount,
		"failed", failedCount,
		"status", finalStatus,
	)

	return nil
}

// seekToTargetFeatures は"targetFeatures"配列の開始を探す
func (uc *ProcessImportUseCase) seekToTargetFeatures(decoder *json.Decoder) error {
	// オブジェクトの開始 '{'
	t, err := decoder.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return apperror.InternalError("予期しないJSONフォーマットです")
	}

	// "targetFeatures"キーを探す
	for decoder.More() {
		t, err := decoder.Token()
		if err != nil {
			return err
		}

		if key, ok := t.(string); ok && key == "targetFeatures" {
			// 配列の開始 '['
			t, err := decoder.Token()
			if err != nil {
				return err
			}
			if delim, ok := t.(json.Delim); !ok || delim != '[' {
				return apperror.InternalError("targetFeaturesが配列ではありません")
			}
			return nil
		}
	}

	return apperror.InternalError("targetFeaturesが見つかりません")
}

// processBatch はバッチを処理する
func (uc *ProcessImportUseCase) processBatch(ctx context.Context, batch []entity.WagriFeature) error {
	return uc.fieldRepo.UpsertBatch(ctx, batch)
}

// handleError はエラーを処理してジョブを失敗状態にする
func (uc *ProcessImportUseCase) handleError(ctx context.Context, jobID uuid.UUID, message string, err error) {
	uc.logger.Error(message, "import_job_id", jobID, "error", err)
	if updateErr := uc.importJobRepo.UpdateError(ctx, jobID, message+": "+err.Error(), nil); updateErr != nil {
		uc.logger.Warn("エラー情報の更新に失敗", "error", updateErr)
	}
}
