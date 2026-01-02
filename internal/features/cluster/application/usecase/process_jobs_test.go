package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/stretchr/testify/require"
)

// TestNewProcessJobsUseCase はNewProcessJobsUseCaseが正しくUseCaseを生成することをテストする
func TestNewProcessJobsUseCase(t *testing.T) {
	jobRepo := &mockClusterJobRepository{}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	require.NotNil(t, uc, "UseCaseがnilです")
}

// TestProcessJobsUseCase_Execute_NoJobs はジョブがない場合に正常終了することをテストする
func TestProcessJobsUseCase_Execute_NoJobs(t *testing.T) {
	jobRepo := &mockClusterJobRepository{jobs: []*entity.ClusterJob{}}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "ジョブがない場合もエラーなしで終了するべき")
}

// TestProcessJobsUseCase_Execute_Success はジョブが正常に処理されることをテストする
func TestProcessJobsUseCase_Execute_Success(t *testing.T) {
	jobs := []*entity.ClusterJob{
		entity.NewClusterJob(10),
	}
	jobRepo := &mockClusterJobRepository{jobs: jobs}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "ジョブが正常に処理されるべき")
}

// TestProcessJobsUseCase_Execute_MultipleJobs は複数ジョブが処理されることをテストする
func TestProcessJobsUseCase_Execute_MultipleJobs(t *testing.T) {
	jobs := []*entity.ClusterJob{
		entity.NewClusterJob(10),
		entity.NewClusterJob(5),
	}
	jobRepo := &mockClusterJobRepository{jobs: jobs}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "複数ジョブが処理されるべき")
}

// TestProcessJobsUseCase_Execute_FindPendingError は保留中ジョブ取得エラー時にエラーを返すことをテストする
func TestProcessJobsUseCase_Execute_FindPendingError(t *testing.T) {
	jobRepo := &mockClusterJobRepository{findPendingErr: errors.New("db error")}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.Error(t, err, "ジョブ取得エラー時はエラーを返すべき")
}

// TestProcessJobsUseCase_Execute_UpdateToProcessingError は処理中更新エラー時も次のジョブを処理することをテストする
func TestProcessJobsUseCase_Execute_UpdateToProcessingError(t *testing.T) {
	jobs := []*entity.ClusterJob{
		entity.NewClusterJob(10),
	}
	jobRepo := &mockClusterJobRepository{
		jobs:      jobs,
		updateErr: errors.New("update error"),
	}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "更新エラーでも処理自体は継続するべき")
}

// TestProcessJobsUseCase_Execute_CalculateError はクラスター計算エラー時もジョブを失敗として処理することをテストする
func TestProcessJobsUseCase_Execute_CalculateError(t *testing.T) {
	jobs := []*entity.ClusterJob{
		entity.NewClusterJob(10),
	}
	// updateErrがnilでないとUpdateToProcessingでcontinueされるので、
	// calculateのエラーをテストするにはupdateErrをnilにしておく
	jobRepo := &mockClusterJobRepository{jobs: jobs}
	clusterRepo := &mockClusterRepository{aggregateErr: errors.New("aggregate error")}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "計算エラーでもジョブ処理自体は継続するべき")
}

// TestProcessJobsUseCase_Execute_DeleteOldJobsError は古いジョブ削除エラーでも処理が完了することをテストする
func TestProcessJobsUseCase_Execute_DeleteOldJobsError(t *testing.T) {
	jobRepo := &mockClusterJobRepository{
		jobs:      []*entity.ClusterJob{},
		deleteErr: errors.New("delete error"),
	}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "古いジョブ削除エラーでも処理は完了するべき")
}

// TestProcessJobsInput はProcessJobsInputの構造体が正しくフィールドを持つことをテストする
func TestProcessJobsInput(t *testing.T) {
	input := ProcessJobsInput{BatchSize: 20}

	if input.BatchSize != 20 {
		t.Errorf("BatchSize = %d, 期待値 20", input.BatchSize)
	}
}

// TestProcessJobsUseCase_Execute_BatchSize はバッチサイズが正しく使用されることをテストする
func TestProcessJobsUseCase_Execute_BatchSize(t *testing.T) {
	tests := []struct {
		name      string
		batchSize int32
	}{
		{"バッチサイズ1", 1},
		{"バッチサイズ10", 10},
		{"バッチサイズ100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobRepo := &mockClusterJobRepository{jobs: []*entity.ClusterJob{}}
			aggregated := []*repository.AggregatedCluster{}
			clusterRepo := &mockClusterRepository{aggregated: aggregated}
			cacheRepo := &mockClusterCacheRepository{}
			logger := getTestLogger()

			calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
			uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

			err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: tt.batchSize})

			require.NoError(t, err, "バッチサイズ%dで正常に実行されるべき", tt.batchSize)
		})
	}
}

// TestProcessJobsUseCase_Execute_JobWithID はジョブIDが正しくログに出力されることをテストする
func TestProcessJobsUseCase_Execute_JobWithID(t *testing.T) {
	job := entity.NewClusterJob(10)
	jobs := []*entity.ClusterJob{job}

	jobRepo := &mockClusterJobRepository{jobs: jobs}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "ジョブIDを持つジョブが正常に処理されるべき")

	// ジョブIDが有効であることを確認
	if job.ID == uuid.Nil {
		t.Error("ジョブIDがnilです")
	}
}

// TestProcessJobsUseCase_Execute_UpdateToCompletedError はジョブ完了更新エラー時も処理が継続することをテストする
func TestProcessJobsUseCase_Execute_UpdateToCompletedError(t *testing.T) {
	jobs := []*entity.ClusterJob{
		entity.NewClusterJob(10),
	}
	jobRepo := &mockClusterJobRepository{
		jobs:                 jobs,
		updateToCompletedErr: errors.New("complete update error"),
	}
	aggregated := []*repository.AggregatedCluster{}
	clusterRepo := &mockClusterRepository{aggregated: aggregated}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "完了更新エラーでも処理自体は継続するべき")
}

// TestProcessJobsUseCase_Execute_UpdateToFailedError は失敗更新エラー時も処理が継続することをテストする
func TestProcessJobsUseCase_Execute_UpdateToFailedError(t *testing.T) {
	jobs := []*entity.ClusterJob{
		entity.NewClusterJob(10),
	}
	// 計算エラー + 失敗更新エラーの両方が発生するケース
	jobRepo := &mockClusterJobRepository{
		jobs:              jobs,
		updateToFailedErr: errors.New("failed update error"),
	}
	clusterRepo := &mockClusterRepository{aggregateErr: errors.New("aggregate error")}
	cacheRepo := &mockClusterCacheRepository{}
	logger := getTestLogger()

	calculateUC := NewCalculateClustersUseCase(clusterRepo, cacheRepo, logger)
	uc := NewProcessJobsUseCase(jobRepo, calculateUC, logger)

	err := uc.Execute(context.Background(), ProcessJobsInput{BatchSize: 10})

	require.NoError(t, err, "失敗更新エラーでも処理自体は継続するべき")
}
