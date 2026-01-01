package port

import (
	"context"

	"github.com/google/uuid"
)

// WorkflowInput はStep Functionsワークフローの入力
type WorkflowInput struct {
	ImportJobID uuid.UUID `json:"import_job_id"`
	CityCode    string    `json:"city_code"`
}

// WorkflowExecution はStep Functionsワークフローの実行情報
type WorkflowExecution struct {
	ExecutionArn string `json:"execution_arn"`
	StartDate    string `json:"start_date"`
}

// StepFunctionsClient はStep Functions操作のインターフェース
type StepFunctionsClient interface {
	// StartExecution はワークフローの実行を開始する
	StartExecution(ctx context.Context, input WorkflowInput) (*WorkflowExecution, error)

	// GetExecutionStatus はワークフローの実行状態を取得する
	GetExecutionStatus(ctx context.Context, executionArn string) (string, error)

	// StopExecution はワークフローの実行を停止する
	StopExecution(ctx context.Context, executionArn string, cause string) error
}
