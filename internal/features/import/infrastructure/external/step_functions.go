package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	appConfig "github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/features/import/application/port"
)

// stepFunctionsClient はStepFunctionsClientの実装
type stepFunctionsClient struct {
	client          *sfn.Client
	stateMachineArn string
}

// NewStepFunctionsClient は新しいStepFunctionsClientを作成する
func NewStepFunctionsClient(ctx context.Context, cfg *appConfig.AWSConfig) (port.StepFunctionsClient, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	if cfg.LocalStackEnabled {
		opts = append(opts, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               cfg.LocalStackURL,
					HostnameImmutable: true,
				}, nil
			}),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &stepFunctionsClient{
		client:          sfn.NewFromConfig(awsCfg),
		stateMachineArn: cfg.StepFunctionsARN,
	}, nil
}

// StartExecution はワークフローの実行を開始する
func (c *stepFunctionsClient) StartExecution(ctx context.Context, input port.WorkflowInput) (*port.WorkflowExecution, error) {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("入力のJSON変換に失敗: %w", err)
	}

	name := fmt.Sprintf("import-%s-%d", input.CityCode, time.Now().UnixNano())

	output, err := c.client.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: aws.String(c.stateMachineArn),
		Name:            aws.String(name),
		Input:           aws.String(string(inputJSON)),
	})
	if err != nil {
		return nil, fmt.Errorf("ワークフロー開始に失敗: %w", err)
	}

	return &port.WorkflowExecution{
		ExecutionArn: *output.ExecutionArn,
		StartDate:    output.StartDate.Format(time.RFC3339),
	}, nil
}

// GetExecutionStatus はワークフローの実行状態を取得する
func (c *stepFunctionsClient) GetExecutionStatus(ctx context.Context, executionArn string) (string, error) {
	output, err := c.client.DescribeExecution(ctx, &sfn.DescribeExecutionInput{
		ExecutionArn: aws.String(executionArn),
	})
	if err != nil {
		return "", fmt.Errorf("実行状態の取得に失敗: %w", err)
	}

	return string(output.Status), nil
}

// StopExecution はワークフローの実行を停止する
func (c *stepFunctionsClient) StopExecution(ctx context.Context, executionArn string, cause string) error {
	_, err := c.client.StopExecution(ctx, &sfn.StopExecutionInput{
		ExecutionArn: aws.String(executionArn),
		Cause:        aws.String(cause),
	})
	if err != nil {
		return fmt.Errorf("実行停止に失敗: %w", err)
	}

	return nil
}
