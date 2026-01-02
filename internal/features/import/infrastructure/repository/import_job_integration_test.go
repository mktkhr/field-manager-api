//go:build integration

package repository

import (
	"context"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	// テスト用DB接続設定
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5433")
	user := getEnvOrDefault("TEST_DB_USER", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "postgres")
	dbname := getEnvOrDefault("TEST_DB_NAME", "field_manager_db_test")

	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	testDB, err = pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("テスト用DB接続に失敗: %v", err)
	}
	defer testDB.Close()

	// 接続確認
	if err := testDB.Ping(ctx); err != nil {
		log.Fatalf("テスト用DBへのPingに失敗: %v", err)
	}

	os.Exit(m.Run())
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func cleanupImportJobs(t *testing.T, ctx context.Context) {
	t.Helper()
	_, err := testDB.Exec(ctx, "DELETE FROM import_jobs")
	if err != nil {
		t.Fatalf("import_jobsのクリーンアップに失敗: %v", err)
	}
}

func TestImportJobRepository_Create_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	job := &entity.ImportJob{
		CityCode: "163210",
	}

	err := repo.Create(ctx, job)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// IDが設定されていることを確認
	if job.ID == uuid.Nil {
		t.Error("Create() should set job.ID")
	}

	// ステータスがpendingになっていることを確認
	if job.Status != entity.ImportStatusPending {
		t.Errorf("Create() Status = %q, want %q", job.Status, entity.ImportStatusPending)
	}

	// CreatedAtが設定されていることを確認
	if job.CreatedAt.IsZero() {
		t.Error("Create() should set job.CreatedAt")
	}
}

func TestImportJobRepository_FindByID_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成
	job := &entity.ImportJob{
		CityCode: "131016",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 作成したジョブを取得
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.ID != job.ID {
		t.Errorf("FindByID() ID = %v, want %v", found.ID, job.ID)
	}
	if found.CityCode != "131016" {
		t.Errorf("FindByID() CityCode = %q, want %q", found.CityCode, "131016")
	}
	if found.Status != entity.ImportStatusPending {
		t.Errorf("FindByID() Status = %q, want %q", found.Status, entity.ImportStatusPending)
	}
}

func TestImportJobRepository_FindByID_NotFound_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// 存在しないIDで検索
	_, err := repo.FindByID(ctx, uuid.New())
	if err == nil {
		t.Error("FindByID() expected error for non-existent ID, got nil")
	}
}

func TestImportJobRepository_UpdateStatus_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// ステータスをprocessingに更新
	err := repo.UpdateStatus(ctx, job.ID, entity.ImportStatusProcessing)
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.Status != entity.ImportStatusProcessing {
		t.Errorf("UpdateStatus() Status = %q, want %q", found.Status, entity.ImportStatusProcessing)
	}

	// started_atが設定されていることを確認
	if found.StartedAt == nil {
		t.Error("UpdateStatus() to processing should set StartedAt")
	}
}

func TestImportJobRepository_UpdateStatus_ToCompleted_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成して処理中にする
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.UpdateStatus(ctx, job.ID, entity.ImportStatusProcessing); err != nil {
		t.Fatalf("UpdateStatus() to processing error = %v", err)
	}

	// ステータスをcompletedに更新
	err := repo.UpdateStatus(ctx, job.ID, entity.ImportStatusCompleted)
	if err != nil {
		t.Fatalf("UpdateStatus() to completed error = %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.Status != entity.ImportStatusCompleted {
		t.Errorf("UpdateStatus() Status = %q, want %q", found.Status, entity.ImportStatusCompleted)
	}

	// completed_atが設定されていることを確認
	if found.CompletedAt == nil {
		t.Error("UpdateStatus() to completed should set CompletedAt")
	}
}

func TestImportJobRepository_UpdateProgress_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 進捗を更新
	err := repo.UpdateProgress(ctx, job.ID, 500, 10, 5)
	if err != nil {
		t.Fatalf("UpdateProgress() error = %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.ProcessedRecords != 500 {
		t.Errorf("UpdateProgress() ProcessedRecords = %d, want %d", found.ProcessedRecords, 500)
	}
	if found.FailedRecords != 10 {
		t.Errorf("UpdateProgress() FailedRecords = %d, want %d", found.FailedRecords, 10)
	}
	if found.LastProcessedBatch != 5 {
		t.Errorf("UpdateProgress() LastProcessedBatch = %d, want %d", found.LastProcessedBatch, 5)
	}
}

func TestImportJobRepository_UpdateS3Key_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// S3キーを更新
	s3Key := "imports/163210/2024-01-01T00:00:00.json"
	err := repo.UpdateS3Key(ctx, job.ID, s3Key)
	if err != nil {
		t.Fatalf("UpdateS3Key() error = %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.S3Key == nil {
		t.Error("UpdateS3Key() S3Key should not be nil")
	} else if *found.S3Key != s3Key {
		t.Errorf("UpdateS3Key() S3Key = %q, want %q", *found.S3Key, s3Key)
	}
}

func TestImportJobRepository_UpdateExecutionArn_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 実行ARNを更新
	arn := "arn:aws:states:ap-northeast-1:123456789012:execution:wagri-import:test-123"
	err := repo.UpdateExecutionArn(ctx, job.ID, arn)
	if err != nil {
		t.Fatalf("UpdateExecutionArn() error = %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.ExecutionArn == nil {
		t.Error("UpdateExecutionArn() ExecutionArn should not be nil")
	} else if *found.ExecutionArn != arn {
		t.Errorf("UpdateExecutionArn() ExecutionArn = %q, want %q", *found.ExecutionArn, arn)
	}
}

func TestImportJobRepository_UpdateTotalRecords_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 総レコード数を更新
	var total int32 = 10000
	err := repo.UpdateTotalRecords(ctx, job.ID, total)
	if err != nil {
		t.Fatalf("UpdateTotalRecords() error = %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.TotalRecords == nil {
		t.Error("UpdateTotalRecords() TotalRecords should not be nil")
	} else if *found.TotalRecords != total {
		t.Errorf("UpdateTotalRecords() TotalRecords = %d, want %d", *found.TotalRecords, total)
	}
}

func TestImportJobRepository_UpdateError_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// エラー情報を更新
	errorMessage := "データの処理中にエラーが発生しました"
	failedIDs := []string{"field-001", "field-002", "field-003"}

	err := repo.UpdateError(ctx, job.ID, errorMessage, failedIDs)
	if err != nil {
		t.Fatalf("UpdateError() error = %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	// ステータスがfailedになっていることを確認
	if found.Status != entity.ImportStatusFailed {
		t.Errorf("UpdateError() Status = %q, want %q", found.Status, entity.ImportStatusFailed)
	}

	// エラーメッセージが設定されていることを確認
	if found.ErrorMessage == nil {
		t.Error("UpdateError() ErrorMessage should not be nil")
	} else if *found.ErrorMessage != errorMessage {
		t.Errorf("UpdateError() ErrorMessage = %q, want %q", *found.ErrorMessage, errorMessage)
	}

	// 失敗レコードIDが設定されていることを確認
	if len(found.FailedRecordIDs) != 3 {
		t.Errorf("UpdateError() len(FailedRecordIDs) = %d, want 3", len(found.FailedRecordIDs))
	}
	for i, id := range failedIDs {
		if found.FailedRecordIDs[i] != id {
			t.Errorf("UpdateError() FailedRecordIDs[%d] = %q, want %q", i, found.FailedRecordIDs[i], id)
		}
	}

	// completed_atが設定されていることを確認
	if found.CompletedAt == nil {
		t.Error("UpdateError() should set CompletedAt")
	}
}

func TestImportJobRepository_UpdateError_EmptyFailedIDs_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// ジョブを作成
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// エラー情報を更新(失敗IDなし)
	errorMessage := "接続エラー"
	var failedIDs []string

	err := repo.UpdateError(ctx, job.ID, errorMessage, failedIDs)
	if err != nil {
		t.Fatalf("UpdateError() error = %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.ErrorMessage == nil {
		t.Error("UpdateError() ErrorMessage should not be nil")
	} else if *found.ErrorMessage != errorMessage {
		t.Errorf("UpdateError() ErrorMessage = %q, want %q", *found.ErrorMessage, errorMessage)
	}

	if len(found.FailedRecordIDs) != 0 {
		t.Errorf("UpdateError() len(FailedRecordIDs) = %d, want 0", len(found.FailedRecordIDs))
	}
}

func TestImportJobRepository_FullWorkflow_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// 1. ジョブを作成
	job := &entity.ImportJob{
		CityCode: "163210",
	}
	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Logf("Created job: %v", job.ID)

	// 2. S3キーと実行ARNを設定
	s3Key := "imports/163210/test.json"
	arn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc"
	if err := repo.UpdateS3Key(ctx, job.ID, s3Key); err != nil {
		t.Fatalf("UpdateS3Key() error = %v", err)
	}
	if err := repo.UpdateExecutionArn(ctx, job.ID, arn); err != nil {
		t.Fatalf("UpdateExecutionArn() error = %v", err)
	}

	// 3. ステータスを処理中に変更し、総レコード数を設定
	if err := repo.UpdateStatus(ctx, job.ID, entity.ImportStatusProcessing); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if err := repo.UpdateTotalRecords(ctx, job.ID, 1000); err != nil {
		t.Fatalf("UpdateTotalRecords() error = %v", err)
	}

	// 4. 進捗を更新(バッチ1)
	if err := repo.UpdateProgress(ctx, job.ID, 200, 5, 1); err != nil {
		t.Fatalf("UpdateProgress() batch 1 error = %v", err)
	}

	// 5. 進捗を更新(バッチ2)
	if err := repo.UpdateProgress(ctx, job.ID, 400, 10, 2); err != nil {
		t.Fatalf("UpdateProgress() batch 2 error = %v", err)
	}

	// 6. 進捗を更新(最終バッチ)
	if err := repo.UpdateProgress(ctx, job.ID, 990, 10, 5); err != nil {
		t.Fatalf("UpdateProgress() final batch error = %v", err)
	}

	// 7. 部分完了としてマーク
	if err := repo.UpdateStatus(ctx, job.ID, entity.ImportStatusPartiallyCompleted); err != nil {
		t.Fatalf("UpdateStatus() to partially_completed error = %v", err)
	}

	// 8. 最終状態を確認
	found, err := repo.FindByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.Status != entity.ImportStatusPartiallyCompleted {
		t.Errorf("Final Status = %q, want %q", found.Status, entity.ImportStatusPartiallyCompleted)
	}
	if found.ProcessedRecords != 990 {
		t.Errorf("Final ProcessedRecords = %d, want %d", found.ProcessedRecords, 990)
	}
	if found.FailedRecords != 10 {
		t.Errorf("Final FailedRecords = %d, want %d", found.FailedRecords, 10)
	}
	if found.LastProcessedBatch != 5 {
		t.Errorf("Final LastProcessedBatch = %d, want %d", found.LastProcessedBatch, 5)
	}
	if found.StartedAt == nil {
		t.Error("Final StartedAt should not be nil")
	}
	if found.CompletedAt == nil {
		t.Error("Final CompletedAt should not be nil")
	}
}

func TestImportJobRepository_MultipleJobs_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupImportJobs(t, ctx)

	repo := NewImportJobRepository(testDB, slog.Default())

	// 複数のジョブを作成
	cityCodes := []string{"163210", "131016", "271004"}
	jobs := make([]*entity.ImportJob, len(cityCodes))

	for i, cityCode := range cityCodes {
		job := &entity.ImportJob{
			CityCode: cityCode,
		}
		if err := repo.Create(ctx, job); err != nil {
			t.Fatalf("Create() for %s error = %v", cityCode, err)
		}
		jobs[i] = job
	}

	// 各ジョブを個別に取得して確認
	for i, job := range jobs {
		found, err := repo.FindByID(ctx, job.ID)
		if err != nil {
			t.Fatalf("FindByID() for job %d error = %v", i, err)
		}
		if found.CityCode != cityCodes[i] {
			t.Errorf("Job %d CityCode = %q, want %q", i, found.CityCode, cityCodes[i])
		}
	}
}

func TestNewImportJobRepository_Integration(t *testing.T) {
	repo := NewImportJobRepository(testDB, slog.Default())
	if repo == nil {
		t.Error("NewImportJobRepository() returned nil")
	}
}

func TestImportJobRepository_Create_Error_Integration(t *testing.T) {
	// キャンセル済みコンテキストでエラーを発生させる
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	repo := NewImportJobRepository(testDB, slog.Default())

	job := &entity.ImportJob{
		CityCode: "163210",
	}

	err := repo.Create(ctx, job)
	if err == nil {
		t.Error("Create() with cancelled context should return error")
	}
}

func TestImportJobRepository_FindByID_Error_Integration(t *testing.T) {
	// キャンセル済みコンテキストでエラーを発生させる
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewImportJobRepository(testDB, slog.Default())

	_, err := repo.FindByID(ctx, uuid.New())
	if err == nil {
		t.Error("FindByID() with cancelled context should return error")
	}
}

func TestImportJobRepository_UpdateStatus_Error_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewImportJobRepository(testDB, slog.Default())

	err := repo.UpdateStatus(ctx, uuid.New(), entity.ImportStatusProcessing)
	if err == nil {
		t.Error("UpdateStatus() with cancelled context should return error")
	}
}

func TestImportJobRepository_UpdateProgress_Error_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewImportJobRepository(testDB, slog.Default())

	err := repo.UpdateProgress(ctx, uuid.New(), 100, 10, 1)
	if err == nil {
		t.Error("UpdateProgress() with cancelled context should return error")
	}
}

func TestImportJobRepository_UpdateS3Key_Error_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewImportJobRepository(testDB, slog.Default())

	err := repo.UpdateS3Key(ctx, uuid.New(), "test-key")
	if err == nil {
		t.Error("UpdateS3Key() with cancelled context should return error")
	}
}

func TestImportJobRepository_UpdateExecutionArn_Error_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewImportJobRepository(testDB, slog.Default())

	err := repo.UpdateExecutionArn(ctx, uuid.New(), "test-arn")
	if err == nil {
		t.Error("UpdateExecutionArn() with cancelled context should return error")
	}
}

func TestImportJobRepository_UpdateTotalRecords_Error_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewImportJobRepository(testDB, slog.Default())

	err := repo.UpdateTotalRecords(ctx, uuid.New(), 1000)
	if err == nil {
		t.Error("UpdateTotalRecords() with cancelled context should return error")
	}
}

func TestImportJobRepository_UpdateError_Error_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewImportJobRepository(testDB, slog.Default())

	err := repo.UpdateError(ctx, uuid.New(), "error", []string{"id1"})
	if err == nil {
		t.Error("UpdateError() with cancelled context should return error")
	}
}
