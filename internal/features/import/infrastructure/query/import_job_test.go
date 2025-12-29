package query

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
)

func TestImportJobQuery_ToEntity(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now
	total := int32(100)
	s3Key := "imports/163210/test.json"
	executionArn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123"
	errorMessage := "Something went wrong"
	failedRecordIDs := []string{"id1", "id2", "id3"}
	failedRecordIDsJSON, _ := json.Marshal(failedRecordIDs)

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
			name: "minimal job",
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
			name: "full job with all fields",
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
			name: "completed job with error",
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
			q := &importJobQuery{}
			result := q.toEntity(tt.row)

			if tt.want == nil {
				if result != nil {
					t.Errorf("toEntity() = %v, 期待値 nil", result)
				}
				return
			}

			if result == nil {
				t.Error("toEntity()がnilです、非nilを期待")
				return
			}

			if result.Status != tt.want.Status {
				t.Errorf("Status = %q, 期待値 %q", result.Status, tt.want.Status)
			}
			if result.CityCode != tt.want.CityCode {
				t.Errorf("CityCode = %q, 期待値 %q", result.CityCode, tt.want.CityCode)
			}
			if result.ProcessedRecords != tt.want.ProcessedRecords {
				t.Errorf("ProcessedRecords = %d, 期待値 %d", result.ProcessedRecords, tt.want.ProcessedRecords)
			}
			if result.FailedRecords != tt.want.FailedRecords {
				t.Errorf("FailedRecords = %d, 期待値 %d", result.FailedRecords, tt.want.FailedRecords)
			}
		})
	}
}

func TestImportJobQuery_ToEntity_FieldMapping(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now
	total := int32(100)
	s3Key := "imports/163210/test.json"
	executionArn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123"
	errorMessage := "Something went wrong"
	failedRecordIDs := []string{"id1", "id2", "id3"}
	failedRecordIDsJSON, _ := json.Marshal(failedRecordIDs)

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

	q := &importJobQuery{}
	result := q.toEntity(row)

	// ID
	if result.ID != id {
		t.Errorf("ID = %v, 期待値 %v", result.ID, id)
	}

	// CityCode
	if result.CityCode != "163210" {
		t.Errorf("CityCode = %q, 期待値 %q", result.CityCode, "163210")
	}

	// Status
	if result.Status != entity.ImportStatusProcessing {
		t.Errorf("Status = %q, 期待値 %q", result.Status, entity.ImportStatusProcessing)
	}

	// TotalRecords
	if result.TotalRecords == nil || *result.TotalRecords != total {
		t.Errorf("TotalRecords = %v, 期待値 %v", result.TotalRecords, &total)
	}

	// ProcessedRecords
	if result.ProcessedRecords != 50 {
		t.Errorf("ProcessedRecords = %d, 期待値 %d", result.ProcessedRecords, 50)
	}

	// FailedRecords
	if result.FailedRecords != 5 {
		t.Errorf("FailedRecords = %d, 期待値 %d", result.FailedRecords, 5)
	}

	// LastProcessedBatch
	if result.LastProcessedBatch != 2 {
		t.Errorf("LastProcessedBatch = %d, 期待値 %d", result.LastProcessedBatch, 2)
	}

	// S3Key
	if result.S3Key == nil || *result.S3Key != s3Key {
		t.Errorf("S3Key = %v, 期待値 %v", result.S3Key, &s3Key)
	}

	// ExecutionArn
	if result.ExecutionArn == nil || *result.ExecutionArn != executionArn {
		t.Errorf("ExecutionArn = %v, 期待値 %v", result.ExecutionArn, &executionArn)
	}

	// ErrorMessage
	if result.ErrorMessage == nil || *result.ErrorMessage != errorMessage {
		t.Errorf("ErrorMessage = %v, 期待値 %v", result.ErrorMessage, &errorMessage)
	}

	// FailedRecordIDs
	if len(result.FailedRecordIDs) != 3 {
		t.Errorf("len(FailedRecordIDs) = %d, 期待値 %d", len(result.FailedRecordIDs), 3)
	}

	// CreatedAt
	if result.CreatedAt.IsZero() {
		t.Error("CreatedAtがゼロ値です")
	}

	// StartedAt
	if result.StartedAt == nil {
		t.Error("StartedAtがnilです")
	}

	// CompletedAt
	if result.CompletedAt == nil {
		t.Error("CompletedAtがnilです")
	}
}

func TestImportJobQuery_ToEntity_NullableFields(t *testing.T) {
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

	q := &importJobQuery{}
	result := q.toEntity(row)

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

func TestImportJobQuery_ToEntity_InvalidFailedRecordIDs(t *testing.T) {
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

	q := &importJobQuery{}
	result := q.toEntity(row)

	// 無効なJSONの場合、空のスライスになるはず
	if len(result.FailedRecordIDs) != 0 {
		t.Errorf("len(FailedRecordIDs) = %d, 無効なJSONに対して0を期待", len(result.FailedRecordIDs))
	}
}

func TestImportJobQuery_ToEntity_AllStatuses(t *testing.T) {
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

			q := &importJobQuery{}
			result := q.toEntity(row)

			if result.Status != s.expected {
				t.Errorf("Status = %q, 期待値 %q", result.Status, s.expected)
			}
		})
	}
}
