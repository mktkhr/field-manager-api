package presentation

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/application/usecase"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/repository"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
	"github.com/stretchr/testify/require"
)

// mockClusterRepository はClusterRepositoryのモック実装
type mockClusterRepository struct {
	clusters   []*entity.Cluster
	aggregated []*repository.AggregatedCluster
	getErr     error
}

func (m *mockClusterRepository) GetClusters(_ context.Context, _ entity.Resolution) ([]*entity.Cluster, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.clusters, nil
}

func (m *mockClusterRepository) SaveClusters(_ context.Context, _ []*entity.Cluster) error {
	return nil
}

func (m *mockClusterRepository) DeleteClustersByResolution(_ context.Context, _ entity.Resolution) error {
	return nil
}

func (m *mockClusterRepository) DeleteAllClusters(_ context.Context) error {
	return nil
}

func (m *mockClusterRepository) AggregateByH3(_ context.Context, _ entity.Resolution) ([]*repository.AggregatedCluster, error) {
	return m.aggregated, nil
}

// mockClusterCacheRepository はClusterCacheRepositoryのモック実装
type mockClusterCacheRepository struct {
	clusters []*entity.Cluster
	getErr   error
}

func (m *mockClusterCacheRepository) GetClusters(_ context.Context, _ entity.Resolution) ([]*entity.Cluster, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.clusters, nil
}

func (m *mockClusterCacheRepository) SetClusters(_ context.Context, _ entity.Resolution, _ []*entity.Cluster) error {
	return nil
}

func (m *mockClusterCacheRepository) DeleteClusters(_ context.Context) error {
	return nil
}

// mockClusterJobRepository はClusterJobRepositoryのモック実装
type mockClusterJobRepository struct {
	hasPendingJob bool
	hasPendingErr error
}

func (m *mockClusterJobRepository) Create(_ context.Context, _ *entity.ClusterJob) error {
	return nil
}

func (m *mockClusterJobRepository) FindByID(_ context.Context, _ uuid.UUID) (*entity.ClusterJob, error) {
	return nil, nil
}

func (m *mockClusterJobRepository) FindPendingJobs(_ context.Context, _ int32) ([]*entity.ClusterJob, error) {
	return nil, nil
}

func (m *mockClusterJobRepository) UpdateToProcessing(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockClusterJobRepository) UpdateToCompleted(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockClusterJobRepository) UpdateToFailed(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockClusterJobRepository) HasPendingOrProcessingJob(_ context.Context) (bool, error) {
	if m.hasPendingErr != nil {
		return false, m.hasPendingErr
	}
	return m.hasPendingJob, nil
}

func (m *mockClusterJobRepository) DeleteOldCompletedJobs(_ context.Context) error {
	return nil
}

func (m *mockClusterJobRepository) DeleteOldFailedJobs(_ context.Context) error {
	return nil
}

// getTestLogger はテスト用のロガーを返す
func getTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// TestNewClusterHandler はNewClusterHandlerが正しくハンドラーを生成することをテストする
func TestNewClusterHandler(t *testing.T) {
	clusterRepo := &mockClusterRepository{}
	cacheRepo := &mockClusterCacheRepository{}
	jobRepo := &mockClusterJobRepository{}
	logger := getTestLogger()

	getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
	handler := NewClusterHandler(getClustersUC, logger)

	require.NotNil(t, handler, "ハンドラーがnilです")
}

// TestClusterHandler_GetClusters_Success は正常にクラスターを取得することをテストする
func TestClusterHandler_GetClusters_Success(t *testing.T) {
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

	clusterRepo := &mockClusterRepository{clusters: clusters}
	cacheRepo := &mockClusterCacheRepository{}
	jobRepo := &mockClusterJobRepository{hasPendingJob: false}
	logger := getTestLogger()

	getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
	handler := NewClusterHandler(getClustersUC, logger)

	request := openapi.GetClustersRequestObject{
		Params: openapi.GetClustersParams{
			Zoom:  12.0,
			SwLat: 35.0,
			SwLng: 139.0,
			NeLat: 36.0,
			NeLng: 140.0,
		},
	}

	response, err := handler.GetClusters(context.Background(), request)

	require.NoError(t, err, "GetClustersでエラーが発生")
	require.NotNil(t, response, "レスポンスがnilです")

	resp200, ok := response.(openapi.GetClusters200JSONResponse)
	require.True(t, ok, "200レスポンスを期待")
	require.Len(t, resp200.Clusters, 1, "クラスター数が期待値と異なります")
	require.False(t, resp200.IsStale, "IsStaleがtrueです")
}

// TestClusterHandler_GetClusters_ValidationError_ZoomOutOfRange はzoomが範囲外の場合にエラーを返すことをテストする
func TestClusterHandler_GetClusters_ValidationError_ZoomOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		zoom float64
	}{
		{"zoom 0", 0.0},
		{"zoom 0.5", 0.5},
		{"zoom 23", 23.0},
		{"zoom 100", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepository{}
			cacheRepo := &mockClusterCacheRepository{}
			jobRepo := &mockClusterJobRepository{}
			logger := getTestLogger()

			getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
			handler := NewClusterHandler(getClustersUC, logger)

			request := openapi.GetClustersRequestObject{
				Params: openapi.GetClustersParams{
					Zoom:  tt.zoom,
					SwLat: 35.0,
					SwLng: 139.0,
					NeLat: 36.0,
					NeLng: 140.0,
				},
			}

			response, err := handler.GetClusters(context.Background(), request)

			require.NoError(t, err, "エラーは返さずにレスポンスで返すべき")
			_, ok := response.(openapi.GetClusters400JSONResponse)
			require.True(t, ok, "400レスポンスを期待")
		})
	}
}

// TestClusterHandler_GetClusters_ValidationError_LatOutOfRange は緯度が範囲外の場合にエラーを返すことをテストする
func TestClusterHandler_GetClusters_ValidationError_LatOutOfRange(t *testing.T) {
	tests := []struct {
		name  string
		swLat float64
		neLat float64
	}{
		{"sw_lat -91", -91.0, 36.0},
		{"sw_lat 91", 91.0, 36.0},
		{"ne_lat -91", 35.0, -91.0},
		{"ne_lat 91", 35.0, 91.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepository{}
			cacheRepo := &mockClusterCacheRepository{}
			jobRepo := &mockClusterJobRepository{}
			logger := getTestLogger()

			getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
			handler := NewClusterHandler(getClustersUC, logger)

			request := openapi.GetClustersRequestObject{
				Params: openapi.GetClustersParams{
					Zoom:  12.0,
					SwLat: tt.swLat,
					SwLng: 139.0,
					NeLat: tt.neLat,
					NeLng: 140.0,
				},
			}

			response, err := handler.GetClusters(context.Background(), request)

			require.NoError(t, err, "エラーは返さずにレスポンスで返すべき")
			_, ok := response.(openapi.GetClusters400JSONResponse)
			require.True(t, ok, "400レスポンスを期待")
		})
	}
}

// TestClusterHandler_GetClusters_ValidationError_LngOutOfRange は経度が範囲外の場合にエラーを返すことをテストする
func TestClusterHandler_GetClusters_ValidationError_LngOutOfRange(t *testing.T) {
	tests := []struct {
		name  string
		swLng float64
		neLng float64
	}{
		{"sw_lng -181", -181.0, 140.0},
		{"sw_lng 181", 181.0, 140.0},
		{"ne_lng -181", 139.0, -181.0},
		{"ne_lng 181", 139.0, 181.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepository{}
			cacheRepo := &mockClusterCacheRepository{}
			jobRepo := &mockClusterJobRepository{}
			logger := getTestLogger()

			getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
			handler := NewClusterHandler(getClustersUC, logger)

			request := openapi.GetClustersRequestObject{
				Params: openapi.GetClustersParams{
					Zoom:  12.0,
					SwLat: 35.0,
					SwLng: tt.swLng,
					NeLat: 36.0,
					NeLng: tt.neLng,
				},
			}

			response, err := handler.GetClusters(context.Background(), request)

			require.NoError(t, err, "エラーは返さずにレスポンスで返すべき")
			_, ok := response.(openapi.GetClusters400JSONResponse)
			require.True(t, ok, "400レスポンスを期待")
		})
	}
}

// TestClusterHandler_GetClusters_ValidationError_SwLatGreaterThanNeLat はsw_latがne_latより大きい場合にエラーを返すことをテストする
func TestClusterHandler_GetClusters_ValidationError_SwLatGreaterThanNeLat(t *testing.T) {
	clusterRepo := &mockClusterRepository{}
	cacheRepo := &mockClusterCacheRepository{}
	jobRepo := &mockClusterJobRepository{}
	logger := getTestLogger()

	getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
	handler := NewClusterHandler(getClustersUC, logger)

	request := openapi.GetClustersRequestObject{
		Params: openapi.GetClustersParams{
			Zoom:  12.0,
			SwLat: 37.0, // ne_latより大きい
			SwLng: 139.0,
			NeLat: 36.0,
			NeLng: 140.0,
		},
	}

	response, err := handler.GetClusters(context.Background(), request)

	require.NoError(t, err, "エラーは返さずにレスポンスで返すべき")
	_, ok := response.(openapi.GetClusters400JSONResponse)
	require.True(t, ok, "400レスポンスを期待")
}

// TestClusterHandler_GetClusters_UseCaseError はUseCaseエラー時に500を返すことをテストする
func TestClusterHandler_GetClusters_UseCaseError(t *testing.T) {
	clusterRepo := &mockClusterRepository{getErr: errors.New("db error")}
	cacheRepo := &mockClusterCacheRepository{}
	jobRepo := &mockClusterJobRepository{}
	logger := getTestLogger()

	getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
	handler := NewClusterHandler(getClustersUC, logger)

	request := openapi.GetClustersRequestObject{
		Params: openapi.GetClustersParams{
			Zoom:  12.0,
			SwLat: 35.0,
			SwLng: 139.0,
			NeLat: 36.0,
			NeLng: 140.0,
		},
	}

	response, err := handler.GetClusters(context.Background(), request)

	require.NoError(t, err, "エラーは返さずにレスポンスで返すべき")
	_, ok := response.(openapi.GetClusters500JSONResponse)
	require.True(t, ok, "500レスポンスを期待")
}

// TestClusterHandler_GetClusters_IsStale はIsStaleフラグが正しく設定されることをテストする
func TestClusterHandler_GetClusters_IsStale(t *testing.T) {
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
	cacheRepo := &mockClusterCacheRepository{}
	jobRepo := &mockClusterJobRepository{hasPendingJob: true} // 再計算中
	logger := getTestLogger()

	getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
	handler := NewClusterHandler(getClustersUC, logger)

	request := openapi.GetClustersRequestObject{
		Params: openapi.GetClustersParams{
			Zoom:  12.0,
			SwLat: 35.0,
			SwLng: 139.0,
			NeLat: 36.0,
			NeLng: 140.0,
		},
	}

	response, err := handler.GetClusters(context.Background(), request)

	require.NoError(t, err, "GetClustersでエラーが発生")

	resp200, ok := response.(openapi.GetClusters200JSONResponse)
	require.True(t, ok, "200レスポンスを期待")
	require.True(t, resp200.IsStale, "再計算中はIsStaleがtrueであるべき")
}

// TestClusterHandler_GetClusters_EmptyClusters は空のクラスターでも正常に動作することをテストする
func TestClusterHandler_GetClusters_EmptyClusters(t *testing.T) {
	clusterRepo := &mockClusterRepository{clusters: []*entity.Cluster{}}
	cacheRepo := &mockClusterCacheRepository{}
	jobRepo := &mockClusterJobRepository{}
	logger := getTestLogger()

	getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
	handler := NewClusterHandler(getClustersUC, logger)

	request := openapi.GetClustersRequestObject{
		Params: openapi.GetClustersParams{
			Zoom:  12.0,
			SwLat: 35.0,
			SwLng: 139.0,
			NeLat: 36.0,
			NeLng: 140.0,
		},
	}

	response, err := handler.GetClusters(context.Background(), request)

	require.NoError(t, err, "GetClustersでエラーが発生")

	resp200, ok := response.(openapi.GetClusters200JSONResponse)
	require.True(t, ok, "200レスポンスを期待")
	require.Len(t, resp200.Clusters, 0, "クラスター数は0であるべき")
}

// TestClusterHandler_GetClusters_BoundaryValues_Zoom はzoomの境界値を正しく処理することをテストする
func TestClusterHandler_GetClusters_BoundaryValues_Zoom(t *testing.T) {
	tests := []struct {
		name string
		zoom float64
	}{
		{"zoom 1.0 (最小値)", 1.0},
		{"zoom 22.0 (最大値)", 22.0},
		{"zoom 1.1", 1.1},
		{"zoom 21.9", 21.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepository{clusters: []*entity.Cluster{}}
			cacheRepo := &mockClusterCacheRepository{}
			jobRepo := &mockClusterJobRepository{}
			logger := getTestLogger()

			getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
			handler := NewClusterHandler(getClustersUC, logger)

			request := openapi.GetClustersRequestObject{
				Params: openapi.GetClustersParams{
					Zoom:  tt.zoom,
					SwLat: 35.0,
					SwLng: 139.0,
					NeLat: 36.0,
					NeLng: 140.0,
				},
			}

			response, err := handler.GetClusters(context.Background(), request)

			require.NoError(t, err, "GetClustersでエラーが発生")
			_, ok := response.(openapi.GetClusters200JSONResponse)
			require.True(t, ok, "200レスポンスを期待")
		})
	}
}

// TestClusterHandler_GetClusters_BoundaryValues_Lat は緯度の境界値を正しく処理することをテストする
func TestClusterHandler_GetClusters_BoundaryValues_Lat(t *testing.T) {
	tests := []struct {
		name  string
		swLat float64
		neLat float64
	}{
		{"緯度 -90 to -89", -90.0, -89.0},
		{"緯度 89 to 90", 89.0, 90.0},
		{"緯度 0 to 1", 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepository{clusters: []*entity.Cluster{}}
			cacheRepo := &mockClusterCacheRepository{}
			jobRepo := &mockClusterJobRepository{}
			logger := getTestLogger()

			getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
			handler := NewClusterHandler(getClustersUC, logger)

			request := openapi.GetClustersRequestObject{
				Params: openapi.GetClustersParams{
					Zoom:  12.0,
					SwLat: tt.swLat,
					SwLng: 139.0,
					NeLat: tt.neLat,
					NeLng: 140.0,
				},
			}

			response, err := handler.GetClusters(context.Background(), request)

			require.NoError(t, err, "GetClustersでエラーが発生")
			_, ok := response.(openapi.GetClusters200JSONResponse)
			require.True(t, ok, "200レスポンスを期待")
		})
	}
}

// TestClusterHandler_GetClusters_BoundaryValues_Lng は経度の境界値を正しく処理することをテストする
func TestClusterHandler_GetClusters_BoundaryValues_Lng(t *testing.T) {
	tests := []struct {
		name  string
		swLng float64
		neLng float64
	}{
		{"経度 -180 to -179", -180.0, -179.0},
		{"経度 179 to 180", 179.0, 180.0},
		{"経度 0 to 1", 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepository{clusters: []*entity.Cluster{}}
			cacheRepo := &mockClusterCacheRepository{}
			jobRepo := &mockClusterJobRepository{}
			logger := getTestLogger()

			getClustersUC := usecase.NewGetClustersUseCase(clusterRepo, cacheRepo, jobRepo, logger)
			handler := NewClusterHandler(getClustersUC, logger)

			request := openapi.GetClustersRequestObject{
				Params: openapi.GetClustersParams{
					Zoom:  12.0,
					SwLat: 35.0,
					SwLng: tt.swLng,
					NeLat: 36.0,
					NeLng: tt.neLng,
				},
			}

			response, err := handler.GetClusters(context.Background(), request)

			require.NoError(t, err, "GetClustersでエラーが発生")
			_, ok := response.(openapi.GetClusters200JSONResponse)
			require.True(t, ok, "200レスポンスを期待")
		})
	}
}

// TestValidationError はValidationErrorのError()メソッドをテストする
func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "zoom",
		Message: "ズームレベルは1.0から22.0の範囲で指定してください",
	}

	expected := "ズームレベルは1.0から22.0の範囲で指定してください"
	if err.Error() != expected {
		t.Errorf("Error() = %q, 期待値 %q", err.Error(), expected)
	}
}

// TestValidationError_Field はValidationErrorのFieldフィールドをテストする
func TestValidationError_Field(t *testing.T) {
	err := &ValidationError{
		Field:   "sw_lat",
		Message: "南西端の緯度は-90から90の範囲で指定してください",
	}

	if err.Field != "sw_lat" {
		t.Errorf("Field = %q, 期待値 sw_lat", err.Field)
	}
}
