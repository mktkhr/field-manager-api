package external

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/import/application/port"
)

// mockSFNAPI はStep Functions APIのモック実装
type mockSFNAPI struct {
	startExecutionFunc    func(ctx context.Context, params *sfn.StartExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error)
	describeExecutionFunc func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error)
	stopExecutionFunc     func(ctx context.Context, params *sfn.StopExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error)
}

func (m *mockSFNAPI) StartExecution(ctx context.Context, params *sfn.StartExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error) {
	if m.startExecutionFunc != nil {
		return m.startExecutionFunc(ctx, params, optFns...)
	}
	now := time.Now()
	return &sfn.StartExecutionOutput{
		ExecutionArn: params.Name,
		StartDate:    &now,
	}, nil
}

func (m *mockSFNAPI) DescribeExecution(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
	if m.describeExecutionFunc != nil {
		return m.describeExecutionFunc(ctx, params, optFns...)
	}
	return &sfn.DescribeExecutionOutput{
		Status: types.ExecutionStatusRunning,
	}, nil
}

func (m *mockSFNAPI) StopExecution(ctx context.Context, params *sfn.StopExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error) {
	if m.stopExecutionFunc != nil {
		return m.stopExecutionFunc(ctx, params, optFns...)
	}
	return &sfn.StopExecutionOutput{}, nil
}

func TestStepFunctionsClient_StartExecution(t *testing.T) {
	tests := []struct {
		name     string
		input    port.WorkflowInput
		mockFunc func(ctx context.Context, params *sfn.StartExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error)
		wantErr  bool
	}{
		{
			name: "success",
			input: port.WorkflowInput{
				ImportJobID: uuid.New(),
				CityCode:    "163210",
			},
			mockFunc: func(ctx context.Context, params *sfn.StartExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error) {
				now := time.Now()
				arn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123"
				return &sfn.StartExecutionOutput{
					ExecutionArn: &arn,
					StartDate:    &now,
				}, nil
			},
			wantErr: false,
		},
		{
			name: "API error",
			input: port.WorkflowInput{
				ImportJobID: uuid.New(),
				CityCode:    "163210",
			},
			mockFunc: func(ctx context.Context, params *sfn.StartExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error) {
				return nil, errors.New("Step Functions error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSFNAPI{
				startExecutionFunc: tt.mockFunc,
			}
			client := &stepFunctionsClient{
				api:             mock,
				stateMachineArn: "arn:aws:states:ap-northeast-1:123456789012:stateMachine:test",
			}

			result, err := client.StartExecution(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("StartExecution() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("StartExecution() error = %v", err)
				return
			}

			if result.ExecutionArn == "" {
				t.Error("ExecutionArn should not be empty")
			}
			if result.StartDate == "" {
				t.Error("StartDate should not be empty")
			}
		})
	}
}

func TestStepFunctionsClient_GetExecutionStatus(t *testing.T) {
	tests := []struct {
		name         string
		executionArn string
		mockFunc     func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error)
		wantStatus   string
		wantErr      bool
	}{
		{
			name:         "running",
			executionArn: "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123",
			mockFunc: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
				return &sfn.DescribeExecutionOutput{
					Status: types.ExecutionStatusRunning,
				}, nil
			},
			wantStatus: "RUNNING",
			wantErr:    false,
		},
		{
			name:         "succeeded",
			executionArn: "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123",
			mockFunc: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
				return &sfn.DescribeExecutionOutput{
					Status: types.ExecutionStatusSucceeded,
				}, nil
			},
			wantStatus: "SUCCEEDED",
			wantErr:    false,
		},
		{
			name:         "failed",
			executionArn: "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123",
			mockFunc: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
				return &sfn.DescribeExecutionOutput{
					Status: types.ExecutionStatusFailed,
				}, nil
			},
			wantStatus: "FAILED",
			wantErr:    false,
		},
		{
			name:         "API error",
			executionArn: "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123",
			mockFunc: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
				return nil, errors.New("Step Functions error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSFNAPI{
				describeExecutionFunc: tt.mockFunc,
			}
			client := &stepFunctionsClient{
				api:             mock,
				stateMachineArn: "arn:aws:states:ap-northeast-1:123456789012:stateMachine:test",
			}

			status, err := client.GetExecutionStatus(context.Background(), tt.executionArn)

			if tt.wantErr {
				if err == nil {
					t.Error("GetExecutionStatus() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetExecutionStatus() error = %v", err)
				return
			}

			if status != tt.wantStatus {
				t.Errorf("GetExecutionStatus() = %q, want %q", status, tt.wantStatus)
			}
		})
	}
}

func TestStepFunctionsClient_StopExecution(t *testing.T) {
	tests := []struct {
		name         string
		executionArn string
		cause        string
		mockFunc     func(ctx context.Context, params *sfn.StopExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error)
		wantErr      bool
	}{
		{
			name:         "success",
			executionArn: "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123",
			cause:        "User cancelled",
			mockFunc:     nil,
			wantErr:      false,
		},
		{
			name:         "API error",
			executionArn: "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123",
			cause:        "User cancelled",
			mockFunc: func(ctx context.Context, params *sfn.StopExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error) {
				return nil, errors.New("Step Functions error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSFNAPI{
				stopExecutionFunc: tt.mockFunc,
			}
			client := &stepFunctionsClient{
				api:             mock,
				stateMachineArn: "arn:aws:states:ap-northeast-1:123456789012:stateMachine:test",
			}

			err := client.StopExecution(context.Background(), tt.executionArn, tt.cause)

			if tt.wantErr {
				if err == nil {
					t.Error("StopExecution() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("StopExecution() error = %v", err)
			}
		})
	}
}

func TestStepFunctionsClient_StartExecutionValidatesParams(t *testing.T) {
	var capturedParams *sfn.StartExecutionInput

	mock := &mockSFNAPI{
		startExecutionFunc: func(ctx context.Context, params *sfn.StartExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error) {
			capturedParams = params
			now := time.Now()
			arn := "test-arn"
			return &sfn.StartExecutionOutput{
				ExecutionArn: &arn,
				StartDate:    &now,
			}, nil
		},
	}
	client := &stepFunctionsClient{
		api:             mock,
		stateMachineArn: "arn:aws:states:ap-northeast-1:123456789012:stateMachine:wagri-import",
	}

	input := port.WorkflowInput{
		ImportJobID: uuid.New(),
		CityCode:    "163210",
	}

	_, err := client.StartExecution(context.Background(), input)
	if err != nil {
		t.Errorf("StartExecution() error = %v", err)
		return
	}

	if capturedParams == nil {
		t.Error("StartExecution was not called")
		return
	}

	if *capturedParams.StateMachineArn != "arn:aws:states:ap-northeast-1:123456789012:stateMachine:wagri-import" {
		t.Errorf("StateMachineArn = %q, want %q", *capturedParams.StateMachineArn, "arn:aws:states:ap-northeast-1:123456789012:stateMachine:wagri-import")
	}

	if capturedParams.Name == nil || *capturedParams.Name == "" {
		t.Error("Name should not be empty")
	}

	if capturedParams.Input == nil || *capturedParams.Input == "" {
		t.Error("Input should not be empty")
	}
}

func TestStepFunctionsClient_GetExecutionStatusValidatesParams(t *testing.T) {
	var capturedParams *sfn.DescribeExecutionInput

	mock := &mockSFNAPI{
		describeExecutionFunc: func(ctx context.Context, params *sfn.DescribeExecutionInput, optFns ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
			capturedParams = params
			return &sfn.DescribeExecutionOutput{
				Status: types.ExecutionStatusRunning,
			}, nil
		},
	}
	client := &stepFunctionsClient{
		api:             mock,
		stateMachineArn: "arn:aws:states:ap-northeast-1:123456789012:stateMachine:test",
	}

	executionArn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123"
	_, err := client.GetExecutionStatus(context.Background(), executionArn)
	if err != nil {
		t.Errorf("GetExecutionStatus() error = %v", err)
		return
	}

	if capturedParams == nil {
		t.Error("DescribeExecution was not called")
		return
	}

	if *capturedParams.ExecutionArn != executionArn {
		t.Errorf("ExecutionArn = %q, want %q", *capturedParams.ExecutionArn, executionArn)
	}
}

func TestStepFunctionsClient_StopExecutionValidatesParams(t *testing.T) {
	var capturedParams *sfn.StopExecutionInput

	mock := &mockSFNAPI{
		stopExecutionFunc: func(ctx context.Context, params *sfn.StopExecutionInput, optFns ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error) {
			capturedParams = params
			return &sfn.StopExecutionOutput{}, nil
		},
	}
	client := &stepFunctionsClient{
		api:             mock,
		stateMachineArn: "arn:aws:states:ap-northeast-1:123456789012:stateMachine:test",
	}

	executionArn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123"
	cause := "Manual cancellation"

	err := client.StopExecution(context.Background(), executionArn, cause)
	if err != nil {
		t.Errorf("StopExecution() error = %v", err)
		return
	}

	if capturedParams == nil {
		t.Error("StopExecution was not called")
		return
	}

	if *capturedParams.ExecutionArn != executionArn {
		t.Errorf("ExecutionArn = %q, want %q", *capturedParams.ExecutionArn, executionArn)
	}
	if *capturedParams.Cause != cause {
		t.Errorf("Cause = %q, want %q", *capturedParams.Cause, cause)
	}
}
