// Package server はHTTPサーバーのルーティングとDI設定を提供する
package server

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/application/usecase"
	clusterRepo "github.com/mktkhr/field-manager-api/internal/features/cluster/infrastructure/repository"
	clusterHandler "github.com/mktkhr/field-manager-api/internal/features/cluster/presentation"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/cache"
)

// StrictServerHandler はStrictServerInterfaceを実装する
type StrictServerHandler struct {
	clusterHandler *clusterHandler.ClusterHandler
	logger         *slog.Logger
}

// NewStrictServerHandler はStrictServerHandlerを作成する
func NewStrictServerHandler(
	pool *pgxpool.Pool,
	cacheClient *cache.Client,
	logger *slog.Logger,
) *StrictServerHandler {
	// クラスター機能のDI
	clusterRepository := clusterRepo.NewClusterPostgresRepository(pool)
	clusterCacheRepository := clusterRepo.NewClusterCacheRedisRepository(cacheClient)
	clusterJobRepository := clusterRepo.NewClusterJobPostgresRepository(pool)

	getClustersUC := usecase.NewGetClustersUseCase(
		clusterRepository,
		clusterCacheRepository,
		clusterJobRepository,
		logger,
	)

	enqueueJobUC := usecase.NewEnqueueJobUseCase(
		clusterJobRepository,
		logger,
	)

	clusterHdlr := clusterHandler.NewClusterHandler(getClustersUC, enqueueJobUC, logger)

	return &StrictServerHandler{
		clusterHandler: clusterHdlr,
		logger:         logger,
	}
}

// GetClusters はクラスター一覧取得エンドポイント
func (h *StrictServerHandler) GetClusters(ctx context.Context, request openapi.GetClustersRequestObject) (openapi.GetClustersResponseObject, error) {
	return h.clusterHandler.GetClusters(ctx, request)
}

// RecalculateClusters はクラスター再計算リクエストエンドポイント
func (h *StrictServerHandler) RecalculateClusters(ctx context.Context, request openapi.RecalculateClustersRequestObject) (openapi.RecalculateClustersResponseObject, error) {
	return h.clusterHandler.RecalculateClusters(ctx, request)
}

// ListFields は圃場一覧取得エンドポイント(未実装)
func (h *StrictServerHandler) ListFields(_ context.Context, _ openapi.ListFieldsRequestObject) (openapi.ListFieldsResponseObject, error) {
	return openapi.ListFields500JSONResponse{
		Code:    "not_implemented",
		Message: "未実装",
	}, nil
}

// GetField は圃場詳細取得エンドポイント(未実装)
func (h *StrictServerHandler) GetField(_ context.Context, _ openapi.GetFieldRequestObject) (openapi.GetFieldResponseObject, error) {
	return openapi.GetField500JSONResponse{
		Code:    "not_implemented",
		Message: "未実装",
	}, nil
}

// RequestImport はインポートリクエストエンドポイント(未実装)
func (h *StrictServerHandler) RequestImport(_ context.Context, _ openapi.RequestImportRequestObject) (openapi.RequestImportResponseObject, error) {
	return openapi.RequestImport500JSONResponse{
		Code:    "not_implemented",
		Message: "未実装",
	}, nil
}

// GetImportStatus はインポートステータス取得エンドポイント(未実装)
func (h *StrictServerHandler) GetImportStatus(_ context.Context, _ openapi.GetImportStatusRequestObject) (openapi.GetImportStatusResponseObject, error) {
	return openapi.GetImportStatus500JSONResponse{
		Code:    "not_implemented",
		Message: "未実装",
	}, nil
}

// HealthCheck はヘルスチェックエンドポイント
func (h *StrictServerHandler) HealthCheck(_ context.Context, _ openapi.HealthCheckRequestObject) (openapi.HealthCheckResponseObject, error) {
	return openapi.HealthCheck200JSONResponse{
		Status: "ok",
	}, nil
}

// SetupRouter はGinルーターをセットアップする
func SetupRouter(handler openapi.StrictServerInterface) *gin.Engine {
	router := gin.Default()
	strictHandler := openapi.NewStrictHandler(handler, nil)
	openapi.RegisterHandlers(router, strictHandler)
	return router
}
