package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/shared/types"
)

// mockStorageClient はStorageClientのモック実装
type mockStorageClient struct {
	data []byte
	err  error
}

func (m *mockStorageClient) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	return nil
}

func (m *mockStorageClient) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(bytes.NewReader(m.data)), nil
}

func (m *mockStorageClient) GetObjectStream(ctx context.Context, key string) (io.ReadCloser, error) {
	return m.Download(ctx, key)
}

func (m *mockStorageClient) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStorageClient) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

// mockFieldRepository はFieldRepositoryのモック実装
type mockFieldRepository struct {
	err error
}

func (m *mockFieldRepository) UpsertBatch(ctx context.Context, inputs []types.FieldBatchInput) error {
	return m.err
}

// testImportJobRepository はテスト用のImportJobRepositoryモック
type testImportJobRepository struct {
	job *entity.ImportJob
}

func (r *testImportJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ImportJob, error) {
	return r.job, nil
}

func (r *testImportJobRepository) Create(ctx context.Context, job *entity.ImportJob) error {
	r.job = job
	return nil
}

func (r *testImportJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ImportStatus) error {
	if r.job != nil {
		r.job.Status = status
	}
	return nil
}

func (r *testImportJobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, processed, failed, batch int32) error {
	if r.job != nil {
		r.job.ProcessedRecords = processed
		r.job.FailedRecords = failed
	}
	return nil
}

func (r *testImportJobRepository) UpdateS3Key(ctx context.Context, id uuid.UUID, s3Key string) error {
	return nil
}

func (r *testImportJobRepository) UpdateExecutionArn(ctx context.Context, id uuid.UUID, arn string) error {
	return nil
}

func (r *testImportJobRepository) UpdateTotalRecords(ctx context.Context, id uuid.UUID, total int32) error {
	if r.job != nil {
		r.job.TotalRecords = &total
	}
	return nil
}

func (r *testImportJobRepository) UpdateError(ctx context.Context, id uuid.UUID, message string, failedIDs []string) error {
	if r.job != nil {
		r.job.ErrorMessage = &message
		r.job.Status = entity.ImportStatusFailed
	}
	return nil
}

// TestProcessImportUseCase_Execute はExecuteメソッドが正常なJSON、S3エラー、無効なJSON、欠落フィールドを正しく処理することをテストする
func TestProcessImportUseCase_Execute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	validJSON := `{
		"targetFeatures": [
			{
				"type": "Feature",
				"geometry": {
					"type": "LinearPolygon",
					"coordinates": [[[139.0, 35.0], [139.1, 35.0], [139.05, 35.1]]]
				},
				"properties": {
					"ID": "test-id-001",
					"CityCode": "163210",
					"IssueYear": "2024",
					"EditYear": "2024",
					"PointLat": 35.05,
					"PointLng": 139.05,
					"FieldType": "1",
					"Number": 1,
					"SoilLargeCode": "",
					"SoilMiddleCode": "",
					"SoilSmallCode": "",
					"SoilSmallName": "",
					"History": "{}",
					"LastPolygonUuid": "uuid-123",
					"PinInfo": []
				}
			}
		]
	}`

	tests := []struct {
		name           string
		mockStorage    *mockStorageClient
		mockFieldRepo  *mockFieldRepository
		mockImportRepo *testImportJobRepository
		input          ProcessImportInput
		wantErr        bool
	}{
		// 正常系: 有効なJSONデータを正常に処理し、圃場データをUPSERTする
		{
			name: "success with valid JSON",
			mockStorage: &mockStorageClient{
				data: []byte(validJSON),
				err:  nil,
			},
			mockFieldRepo:  &mockFieldRepository{err: nil},
			mockImportRepo: &testImportJobRepository{job: entity.NewImportJob("163210")},
			input: ProcessImportInput{
				ImportJobID: uuid.New(),
				S3Key:       "imports/163210/test.json",
				BatchSize:   1000,
			},
			wantErr: false,
		},
		// 異常系: S3からのデータ取得に失敗した場合はエラーを返す
		{
			name: "S3 error",
			mockStorage: &mockStorageClient{
				data: nil,
				err:  errors.New("S3 error"),
			},
			mockFieldRepo:  &mockFieldRepository{err: nil},
			mockImportRepo: &testImportJobRepository{job: entity.NewImportJob("163210")},
			input: ProcessImportInput{
				ImportJobID: uuid.New(),
				S3Key:       "imports/163210/test.json",
				BatchSize:   1000,
			},
			wantErr: true,
		},
		// 異常系: 不正なJSON形式の場合はパースエラーを返す
		{
			name: "invalid JSON",
			mockStorage: &mockStorageClient{
				data: []byte(`{invalid json`),
				err:  nil,
			},
			mockFieldRepo:  &mockFieldRepository{err: nil},
			mockImportRepo: &testImportJobRepository{job: entity.NewImportJob("163210")},
			input: ProcessImportInput{
				ImportJobID: uuid.New(),
				S3Key:       "imports/163210/test.json",
				BatchSize:   1000,
			},
			wantErr: true,
		},
		// 異常系: targetFeaturesキーが存在しない場合はエラーを返す
		{
			name: "missing targetFeatures",
			mockStorage: &mockStorageClient{
				data: []byte(`{"otherKey": []}`),
				err:  nil,
			},
			mockFieldRepo:  &mockFieldRepository{err: nil},
			mockImportRepo: &testImportJobRepository{job: entity.NewImportJob("163210")},
			input: ProcessImportInput{
				ImportJobID: uuid.New(),
				S3Key:       "imports/163210/test.json",
				BatchSize:   1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewProcessImportUseCase(
				tt.mockImportRepo,
				tt.mockStorage,
				tt.mockFieldRepo,
				logger,
			)

			err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}
		})
	}
}

// TestProcessImportUseCase_DefaultBatchSize はBatchSizeが0の場合にデフォルト値が使用されることをテストする
func TestProcessImportUseCase_DefaultBatchSize(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	validJSON := `{"targetFeatures": []}`

	mockStorage := &mockStorageClient{
		data: []byte(validJSON),
		err:  nil,
	}
	mockFieldRepo := &mockFieldRepository{err: nil}
	mockImportRepo := &testImportJobRepository{job: entity.NewImportJob("163210")}

	uc := NewProcessImportUseCase(
		mockImportRepo,
		mockStorage,
		mockFieldRepo,
		logger,
	)

	// BatchSize = 0 should use default
	input := ProcessImportInput{
		ImportJobID: uuid.New(),
		S3Key:       "imports/163210/test.json",
		BatchSize:   0,
	}

	err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
