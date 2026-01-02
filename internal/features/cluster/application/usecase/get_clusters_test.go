package usecase

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/stretchr/testify/require"
)

// mockClusterRepository はClusterRepositoryのモック実装
type mockClusterRepository struct {
	clusters     []*entity.Cluster
	aggregated   []*repository.AggregatedCluster
	getErr       error
	saveErr      error
	aggregateErr error
	deleteErr    error
}

func (m *mockClusterRepository) GetClusters(_ context.Context, _ entity.Resolution) ([]*entity.Cluster, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.clusters, nil
}

func (m *mockClusterRepository) SaveClusters(_ context.Context, _ []*entity.Cluster) error {
	return m.saveErr
}

func (m *mockClusterRepository) DeleteClustersByResolution(_ context.Context, _ entity.Resolution) error {
	return m.deleteErr
}

func (m *mockClusterRepository) DeleteAllClusters(_ context.Context) error {
	return m.deleteErr
}

func (m *mockClusterRepository) AggregateByH3(_ context.Context, _ entity.Resolution) ([]*repository.AggregatedCluster, error) {
	if m.aggregateErr != nil {
		return nil, m.aggregateErr
	}
	return m.aggregated, nil
}

func (m *mockClusterRepository) AggregateByH3ForCells(_ context.Context, _ entity.Resolution, _ []string) ([]*repository.AggregatedCluster, error) {
	if m.aggregateErr != nil {
		return nil, m.aggregateErr
	}
	return m.aggregated, nil
}

func (m *mockClusterRepository) DeleteClustersByH3Indexes(_ context.Context, _ entity.Resolution, _ []string) error {
	return m.deleteErr
}

// mockClusterCacheRepository はClusterCacheRepositoryのモック実装
type mockClusterCacheRepository struct {
	clusters  []*entity.Cluster
	getErr    error
	setErr    error
	deleteErr error
}

func (m *mockClusterCacheRepository) GetClusters(_ context.Context, _ entity.Resolution) ([]*entity.Cluster, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.clusters, nil
}

func (m *mockClusterCacheRepository) SetClusters(_ context.Context, _ entity.Resolution, _ []*entity.Cluster) error {
	return m.setErr
}

func (m *mockClusterCacheRepository) DeleteClusters(_ context.Context) error {
	return m.deleteErr
}

// mockClusterJobRepository はClusterJobRepositoryのモック実装
type mockClusterJobRepository struct {
	jobs                 []*entity.ClusterJob
	hasPendingJob        bool
	createErr            error
	findByIDErr          error
	findPendingErr       error
	updateErr            error
	updateToCompletedErr error
	updateToFailedErr    error
	hasPendingErr        error
	deleteErr            error
}

func (m *mockClusterJobRepository) Create(_ context.Context, _ *entity.ClusterJob) error {
	return m.createErr
}

func (m *mockClusterJobRepository) CreateWithAffectedCells(_ context.Context, _ *entity.ClusterJob) error {
	return m.createErr
}

func (m *mockClusterJobRepository) FindByID(_ context.Context, _ uuid.UUID) (*entity.ClusterJob, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	if len(m.jobs) > 0 {
		return m.jobs[0], nil
	}
	return nil, nil
}

func (m *mockClusterJobRepository) FindPendingJobs(_ context.Context, _ int32) ([]*entity.ClusterJob, error) {
	if m.findPendingErr != nil {
		return nil, m.findPendingErr
	}
	return m.jobs, nil
}

func (m *mockClusterJobRepository) FindPendingJobsWithAffectedCells(_ context.Context, _ int32) ([]*entity.ClusterJob, error) {
	if m.findPendingErr != nil {
		return nil, m.findPendingErr
	}
	return m.jobs, nil
}

func (m *mockClusterJobRepository) UpdateToProcessing(_ context.Context, _ uuid.UUID) error {
	return m.updateErr
}

func (m *mockClusterJobRepository) UpdateToCompleted(_ context.Context, _ uuid.UUID) error {
	if m.updateToCompletedErr != nil {
		return m.updateToCompletedErr
	}
	return m.updateErr
}

func (m *mockClusterJobRepository) UpdateToFailed(_ context.Context, _ uuid.UUID, _ string) error {
	if m.updateToFailedErr != nil {
		return m.updateToFailedErr
	}
	return m.updateErr
}

func (m *mockClusterJobRepository) HasPendingOrProcessingJob(_ context.Context) (bool, error) {
	if m.hasPendingErr != nil {
		return false, m.hasPendingErr
	}
	return m.hasPendingJob, nil
}

func (m *mockClusterJobRepository) DeleteOldCompletedJobs(_ context.Context) error {
	return m.deleteErr
}

func (m *mockClusterJobRepository) DeleteOldFailedJobs(_ context.Context) error {
	return m.deleteErr
}

// getTestLogger はテスト用のロガーを返す
func getTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// TestNewGetClustersUseCase はNewGetClustersUseCaseが正しくUseCaseを生成することをテストする
func TestNewGetClustersUseCase(t *testing.T) {
	clusterRepo := &mockClusterRepository{}
	cacheRepo := &mockClusterCacheRepository{}
	jobRepo := &mockClusterJobRepository{}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	require.NotNil(t, uc, "UseCaseがnilです")
}

// TestGetClustersUseCase_Execute_CacheHit はキャッシュヒット時に正常にクラスターを返すことをテストする
func TestGetClustersUseCase_Execute_CacheHit(t *testing.T) {
	clusters := []*entity.Cluster{
		{
			ID:           uuid.New(),
			Resolution:   entity.Res7,
			H3Index:      "871f1a4adffffff",
			FieldCount:   10,
			CenterLat:    35.681236,
			CenterLng:    139.767125,
			CalculatedAt: time.Now(),
		},
	}

	clusterRepo := &mockClusterRepository{}
	cacheRepo := &mockClusterCacheRepository{clusters: clusters}
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err, "Executeでエラーが発生")
	require.NotNil(t, output, "出力がnilです")
	require.Len(t, output.Clusters, 1, "クラスター数が期待値と異なります")
	require.False(t, output.IsStale, "IsStaleがtrueです")
}

// TestGetClustersUseCase_Execute_CacheMiss はキャッシュミス時にDBからクラスターを取得することをテストする
func TestGetClustersUseCase_Execute_CacheMiss(t *testing.T) {
	clusters := []*entity.Cluster{
		{
			ID:           uuid.New(),
			Resolution:   entity.Res7,
			H3Index:      "871f1a4adffffff",
			FieldCount:   5,
			CenterLat:    35.5,
			CenterLng:    139.5,
			CalculatedAt: time.Now(),
		},
	}

	clusterRepo := &mockClusterRepository{clusters: clusters}
	cacheRepo := &mockClusterCacheRepository{clusters: nil} // キャッシュミス
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err, "Executeでエラーが発生")
	require.NotNil(t, output, "出力がnilです")
	require.Len(t, output.Clusters, 1, "クラスター数が期待値と異なります")
}

// TestGetClustersUseCase_Execute_CacheError はキャッシュエラー時もDBから取得することをテストする
func TestGetClustersUseCase_Execute_CacheError(t *testing.T) {
	clusters := []*entity.Cluster{
		{
			ID:           uuid.New(),
			Resolution:   entity.Res7,
			H3Index:      "871f1a4adffffff",
			FieldCount:   5,
			CenterLat:    35.5,
			CenterLng:    139.5,
			CalculatedAt: time.Now(),
		},
	}

	clusterRepo := &mockClusterRepository{clusters: clusters}
	cacheRepo := &mockClusterCacheRepository{getErr: errors.New("cache error")}
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err, "キャッシュエラーでもDBから取得できるはず")
	require.NotNil(t, output, "出力がnilです")
}

// TestGetClustersUseCase_Execute_DBError はDBエラー時にエラーを返すことをテストする
func TestGetClustersUseCase_Execute_DBError(t *testing.T) {
	clusterRepo := &mockClusterRepository{getErr: errors.New("db error")}
	cacheRepo := &mockClusterCacheRepository{clusters: nil}
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.Error(t, err, "DBエラー時はエラーを返すべき")
	require.Nil(t, output, "エラー時は出力がnilであるべき")
}

// TestGetClustersUseCase_Execute_IsStale は再計算中のジョブがある場合にIsStaleがtrueになることをテストする
func TestGetClustersUseCase_Execute_IsStale(t *testing.T) {
	clusters := []*entity.Cluster{
		{
			ID:           uuid.New(),
			Resolution:   entity.Res7,
			H3Index:      "871f1a4adffffff",
			FieldCount:   10,
			CenterLat:    35.5,
			CenterLng:    139.5,
			CalculatedAt: time.Now(),
		},
	}

	clusterRepo := &mockClusterRepository{}
	cacheRepo := &mockClusterCacheRepository{clusters: clusters}
	jobRepo := &mockClusterJobRepository{hasPendingJob: true} // 再計算中
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err, "Executeでエラーが発生")
	require.NotNil(t, output, "出力がnilです")
	require.True(t, output.IsStale, "再計算中はIsStaleがtrueであるべき")
}

// TestGetClustersUseCase_Execute_BBoxFilter はBoundingBoxでフィルタリングされることをテストする
func TestGetClustersUseCase_Execute_BBoxFilter(t *testing.T) {
	clusters := []*entity.Cluster{
		{
			ID:           uuid.New(),
			Resolution:   entity.Res7,
			H3Index:      "871f1a4adffffff",
			FieldCount:   10,
			CenterLat:    35.5,
			CenterLng:    139.5,
			CalculatedAt: time.Now(),
		},
		{
			ID:           uuid.New(),
			Resolution:   entity.Res7,
			H3Index:      "871f1a4aeffffff",
			FieldCount:   5,
			CenterLat:    40.0, // BBox外
			CenterLng:    139.5,
			CalculatedAt: time.Now(),
		},
	}

	clusterRepo := &mockClusterRepository{}
	cacheRepo := &mockClusterCacheRepository{clusters: clusters}
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err, "Executeでエラーが発生")
	require.NotNil(t, output, "出力がnilです")
	require.Len(t, output.Clusters, 1, "BBox外のクラスターはフィルタリングされるべき")
}

// TestGetClustersUseCase_Execute_EmptyClusters は空のクラスターでも正常に動作することをテストする
func TestGetClustersUseCase_Execute_EmptyClusters(t *testing.T) {
	clusterRepo := &mockClusterRepository{clusters: []*entity.Cluster{}}
	cacheRepo := &mockClusterCacheRepository{clusters: nil}
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err, "空のクラスターでもエラーにならないはず")
	require.NotNil(t, output, "出力がnilです")
	require.Len(t, output.Clusters, 0, "クラスター数は0であるべき")
}

// TestGetClustersUseCase_Execute_ZoomToResolution は各ズームレベルで正しい解像度が使用されることをテストする
func TestGetClustersUseCase_Execute_ZoomToResolution(t *testing.T) {
	tests := []struct {
		name string
		zoom float64
	}{
		{"zoom 3はRes3", 3.0},
		{"zoom 8はRes5", 8.0},
		{"zoom 12はRes7", 12.0},
		{"zoom 18はRes9", 18.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepository{clusters: []*entity.Cluster{}}
			cacheRepo := &mockClusterCacheRepository{clusters: nil}
			jobRepo := &mockClusterJobRepository{hasPendingJob: false}
			logger := getTestLogger()

			uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

			input := GetClustersInput{
				Zoom:  tt.zoom,
				SWLat: 35.0,
				SWLng: 139.0,
				NELat: 36.0,
				NELng: 140.0,
			}

			output, err := uc.Execute(context.Background(), input)

			require.NoError(t, err, "Executeでエラーが発生")
			require.NotNil(t, output, "出力がnilです")
		})
	}
}

// TestGetClustersUseCase_Execute_HasPendingJobError はHasPendingOrProcessingJobエラー時に安全側に倒してtrueを返すことをテストする
func TestGetClustersUseCase_Execute_HasPendingJobError(t *testing.T) {
	clusters := []*entity.Cluster{
		{
			ID:           uuid.New(),
			Resolution:   entity.Res7,
			H3Index:      "871f1a4adffffff",
			FieldCount:   10,
			CenterLat:    35.5,
			CenterLng:    139.5,
			CalculatedAt: time.Now(),
		},
	}

	clusterRepo := &mockClusterRepository{}
	cacheRepo := &mockClusterCacheRepository{clusters: clusters}
	jobRepo := &mockClusterJobRepository{hasPendingErr: errors.New("db error")}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err, "HasPendingエラーでも処理は継続するべき")
	require.NotNil(t, output, "出力がnilです")
	// エラー時は安全側に倒してtrueを返す(クライアントに「再計算中かもしれない」と伝える)
	require.True(t, output.IsStale, "エラー時はIsStaleはtrueであるべき(安全側)")
}

// TestGetClustersUseCase_Execute_CacheSetError はキャッシュ保存エラーでも処理が継続することをテストする
func TestGetClustersUseCase_Execute_CacheSetError(t *testing.T) {
	clusters := []*entity.Cluster{
		{
			ID:           uuid.New(),
			Resolution:   entity.Res7,
			H3Index:      "871f1a4adffffff",
			FieldCount:   10,
			CenterLat:    35.5,
			CenterLng:    139.5,
			CalculatedAt: time.Now(),
		},
	}

	clusterRepo := &mockClusterRepository{clusters: clusters}
	cacheRepo := &mockClusterCacheRepository{clusters: nil, setErr: errors.New("cache set error")}
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	uc := NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)

	input := GetClustersInput{
		Zoom:  12.0,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err, "キャッシュ保存エラーでも処理は継続するべき")
	require.NotNil(t, output, "出力がnilです")
	require.Len(t, output.Clusters, 1, "クラスター数が期待値と異なります")
}

// TestGetClustersInput はGetClustersInputの構造体が正しくフィールドを持つことをテストする
func TestGetClustersInput(t *testing.T) {
	input := GetClustersInput{
		Zoom:  12.5,
		SWLat: 35.0,
		SWLng: 139.0,
		NELat: 36.0,
		NELng: 140.0,
	}

	if input.Zoom != 12.5 {
		t.Errorf("Zoom = %f, 期待値 12.5", input.Zoom)
	}
	if input.SWLat != 35.0 {
		t.Errorf("SWLat = %f, 期待値 35.0", input.SWLat)
	}
	if input.SWLng != 139.0 {
		t.Errorf("SWLng = %f, 期待値 139.0", input.SWLng)
	}
	if input.NELat != 36.0 {
		t.Errorf("NELat = %f, 期待値 36.0", input.NELat)
	}
	if input.NELng != 140.0 {
		t.Errorf("NELng = %f, 期待値 140.0", input.NELng)
	}
}

// TestGetClustersOutput はGetClustersOutputの構造体が正しくフィールドを持つことをテストする
func TestGetClustersOutput(t *testing.T) {
	output := GetClustersOutput{
		Clusters: []*entity.ClusterResult{
			{H3Index: "test", Lat: 35.0, Lng: 139.0, Count: 10},
		},
		IsStale: true,
	}

	if len(output.Clusters) != 1 {
		t.Errorf("Clusters長さ = %d, 期待値 1", len(output.Clusters))
	}
	if !output.IsStale {
		t.Error("IsStale = false, 期待値 true")
	}
}
