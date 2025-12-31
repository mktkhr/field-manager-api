package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

// mockImportJobQuery はImportJobQueryのモック実装
type mockImportJobQuery struct {
	job *entity.ImportJob
	err error
}

func (m *mockImportJobQuery) FindByID(ctx context.Context, id uuid.UUID) (*entity.ImportJob, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.job, nil
}

func (m *mockImportJobQuery) List(ctx context.Context, limit, offset int32) ([]*entity.ImportJob, error) {
	return nil, nil
}

func (m *mockImportJobQuery) ListByCityCode(ctx context.Context, cityCode string, limit, offset int32) ([]*entity.ImportJob, error) {
	return nil, nil
}

func (m *mockImportJobQuery) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockImportJobQuery) CountByStatus(ctx context.Context, status entity.ImportStatus) (int64, error) {
	return 0, nil
}

// TestGetImportStatusUseCase_Execute はExecuteメソッドが正常系、存在しないジョブ、DBエラーを正しく処理することをテストする
func TestGetImportStatusUseCase_Execute(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	total := int32(100)

	tests := []struct {
		name      string
		mockQuery *mockImportJobQuery
		wantErr   bool
		errType   string
	}{
		// 正常系: 存在するジョブIDで正しくステータスを取得できる
		{
			name: "success",
			mockQuery: &mockImportJobQuery{
				job: &entity.ImportJob{
					ID:               uuid.New(),
					CityCode:         "163210",
					Status:           entity.ImportStatusCompleted,
					TotalRecords:     &total,
					ProcessedRecords: 100,
					FailedRecords:    0,
					CreatedAt:        now,
					StartedAt:        &startedAt,
					CompletedAt:      &now,
				},
			},
			wantErr: false,
		},
		// 異常系: 存在しないジョブIDの場合はNOT_FOUNDエラーを返す
		{
			name: "not found",
			mockQuery: &mockImportJobQuery{
				job: nil,
				err: nil,
			},
			wantErr: true,
			errType: "NOT_FOUND",
		},
		// 異常系: DBエラーの場合はINTERNAL_ERRORを返す
		{
			name: "internal error",
			mockQuery: &mockImportJobQuery{
				job: nil,
				err: errors.New("database error"),
			},
			wantErr: true,
			errType: "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewGetImportStatusUseCase(tt.mockQuery)

			output, err := uc.Execute(context.Background(), uuid.New())

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}

			if output == nil {
				t.Error("Execute() returned nil output")
				return
			}

			if output.CityCode != tt.mockQuery.job.CityCode {
				t.Errorf("CityCode = %q, want %q", output.CityCode, tt.mockQuery.job.CityCode)
			}
			if output.Status != tt.mockQuery.job.Status {
				t.Errorf("Status = %q, want %q", output.Status, tt.mockQuery.job.Status)
			}
		})
	}
}
