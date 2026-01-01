package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
	"github.com/stretchr/testify/require"
)

// TestImportJobRepository_ToEntity はtoEntityメソッドがsqlc.ImportJobをEntity.ImportJobに正しく変換することをテストする
func TestImportJobRepository_ToEntity(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now
	total := int32(100)
	s3Key := "imports/163210/test.json"
	executionArn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123"
	errorMessage := "Something went wrong"
	failedRecordIDs := []string{"id1", "id2", "id3"}
	failedRecordIDsJSON, err := json.Marshal(failedRecordIDs)
	require.NoError(t, err, "json.Marshalでエラーが発生")

	tests := []struct {
		name string
		row  *sqlc.ImportJob
		want *entity.ImportJob
	}{
		{
			name: "nil row",
			row:  nil,
			want: nil,
		},
		{
			name: "minimal job (pending)",
			row: &sqlc.ImportJob{
				ID:               uuid.New(),
				CityCode:         "163210",
				Status:           "pending",
				ProcessedRecords: 0,
				FailedRecords:    0,
				CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.ImportJob{
				Status:           entity.ImportStatusPending,
				CityCode:         "163210",
				ProcessedRecords: 0,
				FailedRecords:    0,
			},
		},
		{
			name: "processing job",
			row: &sqlc.ImportJob{
				ID:                 uuid.New(),
				CityCode:           "163210",
				Status:             "processing",
				TotalRecords:       &total,
				ProcessedRecords:   50,
				FailedRecords:      5,
				LastProcessedBatch: 2,
				S3Key:              &s3Key,
				ExecutionArn:       &executionArn,
				CreatedAt:          pgtype.Timestamptz{Time: now, Valid: true},
				StartedAt:          pgtype.Timestamptz{Time: startedAt, Valid: true},
			},
			want: &entity.ImportJob{
				Status:             entity.ImportStatusProcessing,
				CityCode:           "163210",
				ProcessedRecords:   50,
				FailedRecords:      5,
				LastProcessedBatch: 2,
			},
		},
		{
			name: "completed job",
			row: &sqlc.ImportJob{
				ID:               uuid.New(),
				CityCode:         "163210",
				Status:           "completed",
				TotalRecords:     &total,
				ProcessedRecords: 100,
				FailedRecords:    0,
				S3Key:            &s3Key,
				ExecutionArn:     &executionArn,
				CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
				StartedAt:        pgtype.Timestamptz{Time: startedAt, Valid: true},
				CompletedAt:      pgtype.Timestamptz{Time: completedAt, Valid: true},
			},
			want: &entity.ImportJob{
				Status:           entity.ImportStatusCompleted,
				CityCode:         "163210",
				ProcessedRecords: 100,
				FailedRecords:    0,
			},
		},
		{
			name: "failed job with error",
			row: &sqlc.ImportJob{
				ID:               uuid.New(),
				CityCode:         "163210",
				Status:           "failed",
				TotalRecords:     &total,
				ProcessedRecords: 80,
				FailedRecords:    20,
				S3Key:            &s3Key,
				ExecutionArn:     &executionArn,
				ErrorMessage:     &errorMessage,
				FailedRecordIds:  failedRecordIDsJSON,
				CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
				StartedAt:        pgtype.Timestamptz{Time: startedAt, Valid: true},
				CompletedAt:      pgtype.Timestamptz{Time: completedAt, Valid: true},
			},
			want: &entity.ImportJob{
				Status:           entity.ImportStatusFailed,
				CityCode:         "163210",
				ProcessedRecords: 80,
				FailedRecords:    20,
			},
		},
		{
			name: "partially completed job",
			row: &sqlc.ImportJob{
				ID:               uuid.New(),
				CityCode:         "163210",
				Status:           "partially_completed",
				TotalRecords:     &total,
				ProcessedRecords: 90,
				FailedRecords:    10,
				S3Key:            &s3Key,
				ExecutionArn:     &executionArn,
				CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
				StartedAt:        pgtype.Timestamptz{Time: startedAt, Valid: true},
				CompletedAt:      pgtype.Timestamptz{Time: completedAt, Valid: true},
			},
			want: &entity.ImportJob{
				Status:           entity.ImportStatusPartiallyCompleted,
				CityCode:         "163210",
				ProcessedRecords: 90,
				FailedRecords:    10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &importJobRepository{}
			result := r.toEntity(tt.row)

			if tt.want == nil {
				if result != nil {
					t.Errorf("toEntity() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Error("toEntity() = nil, want non-nil")
				return
			}

			if result.Status != tt.want.Status {
				t.Errorf("Status = %q, want %q", result.Status, tt.want.Status)
			}
			if result.CityCode != tt.want.CityCode {
				t.Errorf("CityCode = %q, want %q", result.CityCode, tt.want.CityCode)
			}
			if result.ProcessedRecords != tt.want.ProcessedRecords {
				t.Errorf("ProcessedRecords = %d, want %d", result.ProcessedRecords, tt.want.ProcessedRecords)
			}
			if result.FailedRecords != tt.want.FailedRecords {
				t.Errorf("FailedRecords = %d, want %d", result.FailedRecords, tt.want.FailedRecords)
			}
		})
	}
}

// TestImportJobRepository_ToEntity_FullFieldMapping はtoEntityメソッドが全フィールドを正しくマッピングすることをテストする
func TestImportJobRepository_ToEntity_FullFieldMapping(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now
	total := int32(100)
	s3Key := "imports/163210/test.json"
	executionArn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123"
	errorMessage := "Something went wrong"
	failedRecordIDs := []string{"id1", "id2", "id3"}
	failedRecordIDsJSON, err := json.Marshal(failedRecordIDs)
	require.NoError(t, err, "json.Marshalでエラーが発生")

	id := uuid.New()
	row := &sqlc.ImportJob{
		ID:                 id,
		CityCode:           "163210",
		Status:             "processing",
		TotalRecords:       &total,
		ProcessedRecords:   50,
		FailedRecords:      5,
		LastProcessedBatch: 2,
		S3Key:              &s3Key,
		ExecutionArn:       &executionArn,
		ErrorMessage:       &errorMessage,
		FailedRecordIds:    failedRecordIDsJSON,
		CreatedAt:          pgtype.Timestamptz{Time: now, Valid: true},
		StartedAt:          pgtype.Timestamptz{Time: startedAt, Valid: true},
		CompletedAt:        pgtype.Timestamptz{Time: completedAt, Valid: true},
	}

	r := &importJobRepository{}
	result := r.toEntity(row)

	// ID
	if result.ID != id {
		t.Errorf("ID = %v, want %v", result.ID, id)
	}

	// CityCode
	if result.CityCode != "163210" {
		t.Errorf("CityCode = %q, want %q", result.CityCode, "163210")
	}

	// Status
	if result.Status != entity.ImportStatusProcessing {
		t.Errorf("Status = %q, want %q", result.Status, entity.ImportStatusProcessing)
	}

	// TotalRecords
	if result.TotalRecords == nil || *result.TotalRecords != total {
		t.Errorf("TotalRecords = %v, want %v", result.TotalRecords, &total)
	}

	// ProcessedRecords
	if result.ProcessedRecords != 50 {
		t.Errorf("ProcessedRecords = %d, want %d", result.ProcessedRecords, 50)
	}

	// FailedRecords
	if result.FailedRecords != 5 {
		t.Errorf("FailedRecords = %d, want %d", result.FailedRecords, 5)
	}

	// LastProcessedBatch
	if result.LastProcessedBatch != 2 {
		t.Errorf("LastProcessedBatch = %d, want %d", result.LastProcessedBatch, 2)
	}

	// S3Key
	if result.S3Key == nil || *result.S3Key != s3Key {
		t.Errorf("S3Key = %v, want %v", result.S3Key, &s3Key)
	}

	// ExecutionArn
	if result.ExecutionArn == nil || *result.ExecutionArn != executionArn {
		t.Errorf("ExecutionArn = %v, want %v", result.ExecutionArn, &executionArn)
	}

	// ErrorMessage
	if result.ErrorMessage == nil || *result.ErrorMessage != errorMessage {
		t.Errorf("ErrorMessage = %v, want %v", result.ErrorMessage, &errorMessage)
	}

	// FailedRecordIDs
	if len(result.FailedRecordIDs) != 3 {
		t.Errorf("len(FailedRecordIDs) = %d, want %d", len(result.FailedRecordIDs), 3)
	}
	for i, id := range []string{"id1", "id2", "id3"} {
		if result.FailedRecordIDs[i] != id {
			t.Errorf("FailedRecordIDs[%d] = %q, want %q", i, result.FailedRecordIDs[i], id)
		}
	}

	// CreatedAt
	if result.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// StartedAt
	if result.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	} else if !result.StartedAt.Equal(startedAt) {
		t.Errorf("StartedAt = %v, want %v", result.StartedAt, startedAt)
	}

	// CompletedAt
	if result.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	} else if !result.CompletedAt.Equal(completedAt) {
		t.Errorf("CompletedAt = %v, want %v", result.CompletedAt, completedAt)
	}
}

// TestImportJobRepository_ToEntity_NullableFields はtoEntityメソッドがNULL値を正しくnilに変換することをテストする
func TestImportJobRepository_ToEntity_NullableFields(t *testing.T) {
	now := time.Now()

	row := &sqlc.ImportJob{
		ID:               uuid.New(),
		CityCode:         "163210",
		Status:           "pending",
		TotalRecords:     nil,
		ProcessedRecords: 0,
		FailedRecords:    0,
		S3Key:            nil,
		ExecutionArn:     nil,
		ErrorMessage:     nil,
		FailedRecordIds:  nil,
		CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
		StartedAt:        pgtype.Timestamptz{Valid: false},
		CompletedAt:      pgtype.Timestamptz{Valid: false},
	}

	r := &importJobRepository{}
	result := r.toEntity(row)

	if result.TotalRecords != nil {
		t.Errorf("TotalRecords = %v, want nil", result.TotalRecords)
	}
	if result.S3Key != nil {
		t.Errorf("S3Key = %v, want nil", result.S3Key)
	}
	if result.ExecutionArn != nil {
		t.Errorf("ExecutionArn = %v, want nil", result.ExecutionArn)
	}
	if result.ErrorMessage != nil {
		t.Errorf("ErrorMessage = %v, want nil", result.ErrorMessage)
	}
	if len(result.FailedRecordIDs) != 0 {
		t.Errorf("len(FailedRecordIDs) = %d, want 0", len(result.FailedRecordIDs))
	}
	if result.StartedAt != nil {
		t.Errorf("StartedAt = %v, want nil", result.StartedAt)
	}
	if result.CompletedAt != nil {
		t.Errorf("CompletedAt = %v, want nil", result.CompletedAt)
	}
}

// TestImportJobRepository_ToEntity_InvalidFailedRecordIDs はtoEntityメソッドが無効なJSONを空のスライスに変換することをテストする
func TestImportJobRepository_ToEntity_InvalidFailedRecordIDs(t *testing.T) {
	now := time.Now()

	// 無効なJSONを設定
	invalidJSON := []byte(`{invalid json}`)

	row := &sqlc.ImportJob{
		ID:               uuid.New(),
		CityCode:         "163210",
		Status:           "failed",
		ProcessedRecords: 0,
		FailedRecords:    0,
		FailedRecordIds:  invalidJSON,
		CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
	}

	r := &importJobRepository{}
	result := r.toEntity(row)

	// 無効なJSONの場合、空のスライスになるはず
	if len(result.FailedRecordIDs) != 0 {
		t.Errorf("len(FailedRecordIDs) = %d, want 0 for invalid JSON", len(result.FailedRecordIDs))
	}
}

// TestImportJobRepository_ToEntity_EmptyFailedRecordIDs はtoEntityメソッドが空のJSON配列を空のスライスに変換することをテストする
func TestImportJobRepository_ToEntity_EmptyFailedRecordIDs(t *testing.T) {
	now := time.Now()

	// 空の配列
	emptyJSON := []byte(`[]`)

	row := &sqlc.ImportJob{
		ID:               uuid.New(),
		CityCode:         "163210",
		Status:           "completed",
		ProcessedRecords: 100,
		FailedRecords:    0,
		FailedRecordIds:  emptyJSON,
		CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
	}

	r := &importJobRepository{}
	result := r.toEntity(row)

	if len(result.FailedRecordIDs) != 0 {
		t.Errorf("len(FailedRecordIDs) = %d, want 0 for empty array", len(result.FailedRecordIDs))
	}
}

// TestImportJobRepository_ToEntity_AllStatuses はtoEntityメソッドが全てのステータスを正しく変換することをテストする
func TestImportJobRepository_ToEntity_AllStatuses(t *testing.T) {
	now := time.Now()

	statuses := []struct {
		dbStatus string
		expected entity.ImportStatus
	}{
		{"pending", entity.ImportStatusPending},
		{"processing", entity.ImportStatusProcessing},
		{"completed", entity.ImportStatusCompleted},
		{"failed", entity.ImportStatusFailed},
		{"partially_completed", entity.ImportStatusPartiallyCompleted},
	}

	for _, s := range statuses {
		t.Run(s.dbStatus, func(t *testing.T) {
			row := &sqlc.ImportJob{
				ID:        uuid.New(),
				CityCode:  "163210",
				Status:    s.dbStatus,
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			}

			r := &importJobRepository{}
			result := r.toEntity(row)

			if result.Status != s.expected {
				t.Errorf("Status = %q, want %q", result.Status, s.expected)
			}
		})
	}
}

// TestImportJobRepository_ToEntity_TimestampConversion はtoEntityメソッドがタイムスタンプを正しく変換することをテストする
func TestImportJobRepository_ToEntity_TimestampConversion(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	startedAt := now.Add(-2 * time.Hour).Truncate(time.Microsecond)
	completedAt := now.Add(-1 * time.Hour).Truncate(time.Microsecond)

	row := &sqlc.ImportJob{
		ID:          uuid.New(),
		CityCode:    "163210",
		Status:      "completed",
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		StartedAt:   pgtype.Timestamptz{Time: startedAt, Valid: true},
		CompletedAt: pgtype.Timestamptz{Time: completedAt, Valid: true},
	}

	r := &importJobRepository{}
	result := r.toEntity(row)

	// CreatedAtの時刻が正しく変換されていることを確認
	if !result.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", result.CreatedAt, now)
	}

	// StartedAtの時刻が正しく変換されていることを確認
	if result.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	} else if !result.StartedAt.Equal(startedAt) {
		t.Errorf("StartedAt = %v, want %v", *result.StartedAt, startedAt)
	}

	// CompletedAtの時刻が正しく変換されていることを確認
	if result.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	} else if !result.CompletedAt.Equal(completedAt) {
		t.Errorf("CompletedAt = %v, want %v", *result.CompletedAt, completedAt)
	}
}

// TestImportJobRepository_ToEntity_LargeFailedRecordIDs はtoEntityメソッドが大量の失敗レコードIDを正しく処理することをテストする
func TestImportJobRepository_ToEntity_LargeFailedRecordIDs(t *testing.T) {
	now := time.Now()

	// 大量のfailed record IDsを生成
	failedIDs := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		failedIDs[i] = uuid.New().String()
	}
	failedIDsJSON, err := json.Marshal(failedIDs)
	require.NoError(t, err, "json.Marshalでエラーが発生")

	row := &sqlc.ImportJob{
		ID:               uuid.New(),
		CityCode:         "163210",
		Status:           "partially_completed",
		ProcessedRecords: 9000,
		FailedRecords:    1000,
		FailedRecordIds:  failedIDsJSON,
		CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
	}

	r := &importJobRepository{}
	result := r.toEntity(row)

	if len(result.FailedRecordIDs) != 1000 {
		t.Errorf("len(FailedRecordIDs) = %d, want 1000", len(result.FailedRecordIDs))
	}

	// 最初と最後のIDが正しく変換されていることを確認
	if result.FailedRecordIDs[0] != failedIDs[0] {
		t.Errorf("FailedRecordIDs[0] = %q, want %q", result.FailedRecordIDs[0], failedIDs[0])
	}
	if result.FailedRecordIDs[999] != failedIDs[999] {
		t.Errorf("FailedRecordIDs[999] = %q, want %q", result.FailedRecordIDs[999], failedIDs[999])
	}
}

// TestImportJobRepository_ToEntity_ProgressCalculation はtoEntityメソッドで変換後のエンティティが進捗を正しく計算することをテストする
func TestImportJobRepository_ToEntity_ProgressCalculation(t *testing.T) {
	now := time.Now()
	total := int32(1000)

	testCases := []struct {
		name             string
		processedRecords int32
		failedRecords    int32
		expectedProgress float64
	}{
		{
			name:             "0% progress",
			processedRecords: 0,
			failedRecords:    0,
			expectedProgress: 0,
		},
		{
			name:             "50% progress (processed only)",
			processedRecords: 500,
			failedRecords:    0,
			expectedProgress: 50,
		},
		{
			name:             "100% progress (all processed)",
			processedRecords: 1000,
			failedRecords:    0,
			expectedProgress: 100,
		},
		{
			name:             "100% progress (with failures)",
			processedRecords: 800,
			failedRecords:    200,
			expectedProgress: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			row := &sqlc.ImportJob{
				ID:               uuid.New(),
				CityCode:         "163210",
				Status:           "processing",
				TotalRecords:     &total,
				ProcessedRecords: tc.processedRecords,
				FailedRecords:    tc.failedRecords,
				CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
			}

			r := &importJobRepository{}
			result := r.toEntity(row)

			// Progress()メソッドを使用して進捗を計算
			progress := result.Progress()
			if progress != tc.expectedProgress {
				t.Errorf("Progress() = %v, want %v", progress, tc.expectedProgress)
			}
		})
	}
}

// TestImportJobRepository_UpdateError_MarshalError はUpdateErrorメソッドがJSONマーシャルエラーを正しく処理することをテストする
func TestImportJobRepository_UpdateError_MarshalError(t *testing.T) {
	// jsonMarshalをモックしてエラーを返す
	originalMarshal := jsonMarshal
	defer func() { jsonMarshal = originalMarshal }()

	marshalError := errors.New("mock marshal error")
	jsonMarshal = func(v interface{}) ([]byte, error) {
		return nil, marshalError
	}

	repo := &importJobRepository{}

	err := repo.UpdateError(context.Background(), uuid.New(), "error message", []string{"id1", "id2"})
	if err == nil {
		t.Error("UpdateError() should return error when json.Marshal fails")
	}
	if !errors.Is(err, marshalError) {
		t.Errorf("UpdateError() error = %v, want %v", err, marshalError)
	}
}
