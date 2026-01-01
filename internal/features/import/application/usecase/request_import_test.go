package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/import/application/port"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

// mockImportJobRepository はImportJobRepositoryのモック実装
type mockImportJobRepository struct {
	createErr        error
	updateStatusErr  error
	updateArnErr     error
	createdJob       *entity.ImportJob
	updatedJobStatus entity.ImportStatus
}

func (m *mockImportJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.ImportJob, error) {
	return nil, nil
}

func (m *mockImportJobRepository) Create(ctx context.Context, job *entity.ImportJob) error {
	if m.createErr != nil {
		return m.createErr
	}
	job.ID = uuid.New()
	m.createdJob = job
	return nil
}

func (m *mockImportJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ImportStatus) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	m.updatedJobStatus = status
	return nil
}

func (m *mockImportJobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, processed, failed, batch int32) error {
	return nil
}

func (m *mockImportJobRepository) UpdateS3Key(ctx context.Context, id uuid.UUID, s3Key string) error {
	return nil
}

func (m *mockImportJobRepository) UpdateExecutionArn(ctx context.Context, id uuid.UUID, arn string) error {
	if m.updateArnErr != nil {
		return m.updateArnErr
	}
	return nil
}

func (m *mockImportJobRepository) UpdateTotalRecords(ctx context.Context, id uuid.UUID, total int32) error {
	return nil
}

func (m *mockImportJobRepository) UpdateError(ctx context.Context, id uuid.UUID, message string, failedIDs []string) error {
	return nil
}

// mockStepFunctionsClient はStepFunctionsClientのモック実装
type mockStepFunctionsClient struct {
	executionArn string
	err          error
}

func (m *mockStepFunctionsClient) StartExecution(ctx context.Context, input port.WorkflowInput) (*port.WorkflowExecution, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &port.WorkflowExecution{
		ExecutionArn: m.executionArn,
		StartDate:    "",
	}, nil
}

func (m *mockStepFunctionsClient) GetExecutionStatus(ctx context.Context, executionArn string) (string, error) {
	return "", nil
}

func (m *mockStepFunctionsClient) StopExecution(ctx context.Context, executionArn string, cause string) error {
	return nil
}

// TestRequestImportUseCase_Execute はExecuteメソッドが正常系、空CityCode、DB作成エラー、Step Functionsエラーを正しく処理することをテストする
func TestRequestImportUseCase_Execute(t *testing.T) {
	tests := []struct {
		name       string
		input      RequestImportInput
		mockRepo   *mockImportJobRepository
		mockSfn    *mockStepFunctionsClient
		wantErr    bool
		wantErrMsg string
	}{
		// 正常系: 有効な市区町村コードでインポートジョブが作成され、Step Functionsが正常に開始される
		{
			name:  "success",
			input: RequestImportInput{CityCode: "163210"},
			mockRepo: &mockImportJobRepository{
				createErr:       nil,
				updateStatusErr: nil,
				updateArnErr:    nil,
			},
			mockSfn: &mockStepFunctionsClient{
				executionArn: "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123",
				err:          nil,
			},
			wantErr: false,
		},
		// 異常系: 市区町村コードが空の場合はバリデーションエラーを返す
		{
			name:       "empty city code",
			input:      RequestImportInput{CityCode: ""},
			mockRepo:   &mockImportJobRepository{},
			mockSfn:    &mockStepFunctionsClient{},
			wantErr:    true,
			wantErrMsg: "市区町村コードは必須です",
		},
		// 異常系: DBへのジョブ作成が失敗した場合はエラーを返す
		{
			name:  "create job error",
			input: RequestImportInput{CityCode: "163210"},
			mockRepo: &mockImportJobRepository{
				createErr: errors.New("database error"),
			},
			mockSfn: &mockStepFunctionsClient{},
			wantErr: true,
		},
		// 異常系: Step Functions実行開始が失敗した場合はエラーを返す
		{
			name:  "start execution error",
			input: RequestImportInput{CityCode: "163210"},
			mockRepo: &mockImportJobRepository{
				createErr: nil,
			},
			mockSfn: &mockStepFunctionsClient{
				err: errors.New("step functions error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewRequestImportUseCase(tt.mockRepo, tt.mockSfn)

			output, err := uc.Execute(context.Background(), tt.input)

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

			if output.ImportJobID == uuid.Nil {
				t.Error("ImportJobID should not be nil UUID")
			}

			if output.ExecutionArn != tt.mockSfn.executionArn {
				t.Errorf("ExecutionArn = %q, want %q", output.ExecutionArn, tt.mockSfn.executionArn)
			}
		})
	}
}
