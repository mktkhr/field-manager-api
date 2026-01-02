package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewEnqueueJobUseCase はNewEnqueueJobUseCaseが正しくUseCaseを生成することをテストする
func TestNewEnqueueJobUseCase(t *testing.T) {
	jobRepo := &mockClusterJobRepository{}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)

	require.NotNil(t, uc, "UseCaseがnilです")
}

// TestEnqueueJobUseCase_Execute_Success は正常にジョブをエンキューすることをテストする
func TestEnqueueJobUseCase_Execute_Success(t *testing.T) {
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)

	output, err := uc.Execute(context.Background(), EnqueueJobInput{Priority: 10})

	require.NoError(t, err, "Executeでエラーが発生")
	require.True(t, output.Enqueued, "ジョブがエンキューされるべき")
}

// TestEnqueueJobUseCase_Execute_SkipWhenPendingJobExists は保留中ジョブがある場合にスキップすることをテストする
func TestEnqueueJobUseCase_Execute_SkipWhenPendingJobExists(t *testing.T) {
	jobRepo := &mockClusterJobRepository{hasPendingJob: true}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)

	output, err := uc.Execute(context.Background(), EnqueueJobInput{Priority: 10})

	require.NoError(t, err, "既存ジョブがある場合はスキップしてエラーなしで終了するべき")
	require.False(t, output.Enqueued, "既存ジョブがある場合はEnqueuedがfalseになるべき")
}

// TestEnqueueJobUseCase_Execute_HasPendingJobError は保留中ジョブ確認でエラー時にエラーを返すことをテストする
func TestEnqueueJobUseCase_Execute_HasPendingJobError(t *testing.T) {
	jobRepo := &mockClusterJobRepository{hasPendingErr: errors.New("db error")}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)

	_, err := uc.Execute(context.Background(), EnqueueJobInput{Priority: 10})

	require.Error(t, err, "DB確認エラー時はエラーを返すべき")
}

// TestEnqueueJobUseCase_Execute_CreateError はジョブ作成エラー時にエラーを返すことをテストする
func TestEnqueueJobUseCase_Execute_CreateError(t *testing.T) {
	jobRepo := &mockClusterJobRepository{
		hasPendingJob: false,
		createErr:     errors.New("create error"),
	}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)

	_, err := uc.Execute(context.Background(), EnqueueJobInput{Priority: 10})

	require.Error(t, err, "ジョブ作成エラー時はエラーを返すべき")
}

// TestEnqueueJobUseCase_Execute_ZeroPriority は優先度0でもエンキューできることをテストする
func TestEnqueueJobUseCase_Execute_ZeroPriority(t *testing.T) {
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)

	output, err := uc.Execute(context.Background(), EnqueueJobInput{Priority: 0})

	require.NoError(t, err, "優先度0でもエンキューできるべき")
	require.True(t, output.Enqueued, "優先度0でもエンキューされるべき")
}

// TestEnqueueJobUseCase_Execute_NegativePriority は負の優先度でもエンキューできることをテストする
func TestEnqueueJobUseCase_Execute_NegativePriority(t *testing.T) {
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)

	output, err := uc.Execute(context.Background(), EnqueueJobInput{Priority: -5})

	require.NoError(t, err, "負の優先度でもエンキューできるべき")
	require.True(t, output.Enqueued, "負の優先度でもエンキューされるべき")
}

// TestEnqueueJobInput はEnqueueJobInputの構造体が正しくフィールドを持つことをテストする
func TestEnqueueJobInput(t *testing.T) {
	input := EnqueueJobInput{Priority: 10}

	if input.Priority != 10 {
		t.Errorf("Priority = %d, 期待値 10", input.Priority)
	}
}

// TestNewClusterJobEnqueuer はNewClusterJobEnqueuerが正しくアダプターを生成することをテストする
func TestNewClusterJobEnqueuer(t *testing.T) {
	jobRepo := &mockClusterJobRepository{}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)
	enqueuer := NewClusterJobEnqueuer(uc)

	require.NotNil(t, enqueuer, "ClusterJobEnqueuerがnilです")
}

// TestClusterJobEnqueuer_Enqueue はEnqueueがUseCaseを正しく呼び出すことをテストする
func TestClusterJobEnqueuer_Enqueue(t *testing.T) {
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)
	enqueuer := NewClusterJobEnqueuer(uc)

	err := enqueuer.Enqueue(context.Background(), 5)

	require.NoError(t, err, "Enqueueでエラーが発生")
}

// TestClusterJobEnqueuer_Enqueue_WithError はEnqueueがエラーを正しく伝播することをテストする
func TestClusterJobEnqueuer_Enqueue_WithError(t *testing.T) {
	jobRepo := &mockClusterJobRepository{
		hasPendingJob: false,
		createErr:     errors.New("create error"),
	}
	logger := getTestLogger()

	uc := NewEnqueueJobUseCase(jobRepo, logger)
	enqueuer := NewClusterJobEnqueuer(uc)

	err := enqueuer.Enqueue(context.Background(), 5)

	require.Error(t, err, "エラーが伝播されるべき")
}
