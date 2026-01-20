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
	fieldUsecase "github.com/mktkhr/field-manager-api/internal/features/field/application/usecase"
	fieldQuery "github.com/mktkhr/field-manager-api/internal/features/field/infrastructure/query"
	fieldHandler "github.com/mktkhr/field-manager-api/internal/features/field/presentation"
	fieldmanagerUsecase "github.com/mktkhr/field-manager-api/internal/features/fieldmanager/application/usecase"
	fieldmanagerRepo "github.com/mktkhr/field-manager-api/internal/features/fieldmanager/infrastructure/repository"
	fieldmanagerHandler "github.com/mktkhr/field-manager-api/internal/features/fieldmanager/presentation"
	fieldsearchUsecase "github.com/mktkhr/field-manager-api/internal/features/fieldsearch/application/usecase"
	fieldsearchH3util "github.com/mktkhr/field-manager-api/internal/features/fieldsearch/infrastructure/h3util"
	fieldsearchQuery "github.com/mktkhr/field-manager-api/internal/features/fieldsearch/infrastructure/query"
	fieldsearchHandler "github.com/mktkhr/field-manager-api/internal/features/fieldsearch/presentation"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
	"github.com/mktkhr/field-manager-api/internal/infrastructure/cache"
)

// StrictServerHandler はStrictServerInterfaceを実装する
type StrictServerHandler struct {
	clusterHandler      *clusterHandler.ClusterHandler
	fieldHandler        *fieldHandler.FieldHandler
	fieldmanagerHandler *fieldmanagerHandler.FieldManagerHandler
	fieldsearchHandler  *fieldsearchHandler.FieldSearchHandler
	logger              *slog.Logger
}

// NewStrictServerHandler はStrictServerHandlerを作成する
func NewStrictServerHandler(
	pool *pgxpool.Pool,
	cacheClient *cache.Client,
	logger *slog.Logger,
) *StrictServerHandler {
	// クラスター機能のDI
	clusterRepository := clusterRepo.NewClusterPostgresRepository(pool, logger)
	clusterCacheRepository := clusterRepo.NewClusterCacheRedisRepository(cacheClient, logger)
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

	// 圃場機能のDI
	fieldQueryImpl := fieldQuery.NewFieldQuery(pool)
	listFieldsUC := fieldUsecase.NewListFieldsUseCase(fieldQueryImpl, logger)
	fieldHdlr := fieldHandler.NewFieldHandler(listFieldsUC, logger)

	// 圃場検索機能のDI
	fieldsearchQueryImpl := fieldsearchQuery.NewFieldSearchQuery(pool)
	h3Calculator := fieldsearchH3util.NewBBoxCellCalculator()
	searchFieldsUC := fieldsearchUsecase.NewSearchFieldsUseCase(fieldsearchQueryImpl, h3Calculator, logger)
	fieldsearchHdlr := fieldsearchHandler.NewFieldSearchHandler(searchFieldsUC, logger)

	// 圃場管理者機能のDI
	fieldmanagerRepository := fieldmanagerRepo.NewFieldManagerRepository(pool, logger)
	fieldExistsChecker := fieldmanagerRepo.NewFieldExistsChecker(pool)
	assignManagerUC := fieldmanagerUsecase.NewAssignManagerUseCase(fieldmanagerRepository, fieldExistsChecker, logger)
	unassignManagerUC := fieldmanagerUsecase.NewUnassignManagerUseCase(fieldmanagerRepository, logger)
	listFieldsByManagerUC := fieldmanagerUsecase.NewListFieldsByManagerUseCase(fieldmanagerRepository, logger)
	listManagersByFieldUC := fieldmanagerUsecase.NewListManagersByFieldUseCase(fieldmanagerRepository, fieldExistsChecker, logger)
	fieldmanagerHdlr := fieldmanagerHandler.NewFieldManagerHandler(
		assignManagerUC,
		unassignManagerUC,
		listFieldsByManagerUC,
		listManagersByFieldUC,
		logger,
	)

	return &StrictServerHandler{
		clusterHandler:      clusterHdlr,
		fieldHandler:        fieldHdlr,
		fieldmanagerHandler: fieldmanagerHdlr,
		fieldsearchHandler:  fieldsearchHdlr,
		logger:              logger,
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

// ListFields は圃場一覧取得エンドポイント
func (h *StrictServerHandler) ListFields(ctx context.Context, request openapi.ListFieldsRequestObject) (openapi.ListFieldsResponseObject, error) {
	return h.fieldHandler.ListFields(ctx, request)
}

// SearchFields は緯度経度範囲による圃場検索エンドポイント
func (h *StrictServerHandler) SearchFields(ctx context.Context, request openapi.SearchFieldsRequestObject) (openapi.SearchFieldsResponseObject, error) {
	return h.fieldsearchHandler.SearchFields(ctx, request)
}

// ListManagersByField は圃場の管理者一覧取得エンドポイント
func (h *StrictServerHandler) ListManagersByField(ctx context.Context, request openapi.ListManagersByFieldRequestObject) (openapi.ListManagersByFieldResponseObject, error) {
	return h.fieldmanagerHandler.ListManagersByField(ctx, request)
}

// AssignManager は管理者割り当てエンドポイント
func (h *StrictServerHandler) AssignManager(ctx context.Context, request openapi.AssignManagerRequestObject) (openapi.AssignManagerResponseObject, error) {
	return h.fieldmanagerHandler.AssignManager(ctx, request)
}

// UnassignManager は管理者解除エンドポイント
func (h *StrictServerHandler) UnassignManager(ctx context.Context, request openapi.UnassignManagerRequestObject) (openapi.UnassignManagerResponseObject, error) {
	return h.fieldmanagerHandler.UnassignManager(ctx, request)
}

// ListFieldsByManager は管理配下の圃場一覧取得エンドポイント
func (h *StrictServerHandler) ListFieldsByManager(ctx context.Context, request openapi.ListFieldsByManagerRequestObject) (openapi.ListFieldsByManagerResponseObject, error) {
	return h.fieldmanagerHandler.ListFieldsByManager(ctx, request)
}

// GetField は圃場詳細取得エンドポイント(未実装)
func (h *StrictServerHandler) GetField(_ context.Context, _ openapi.GetFieldRequestObject) (openapi.GetFieldResponseObject, error) {
	return openapi.GetField501JSONResponse{
		NotImplementedJSONResponse: openapi.NotImplementedJSONResponse{
			Data: nil,
			Errors: &[]openapi.Error{{
				Code:    "not_implemented",
				Message: "このエンドポイントは未実装です",
			}},
		},
	}, nil
}

// RequestImport はインポートリクエストエンドポイント(未実装)
func (h *StrictServerHandler) RequestImport(_ context.Context, _ openapi.RequestImportRequestObject) (openapi.RequestImportResponseObject, error) {
	return openapi.RequestImport501JSONResponse{
		NotImplementedJSONResponse: openapi.NotImplementedJSONResponse{
			Data: nil,
			Errors: &[]openapi.Error{{
				Code:    "not_implemented",
				Message: "このエンドポイントは未実装です",
			}},
		},
	}, nil
}

// GetImportStatus はインポートステータス取得エンドポイント(未実装)
func (h *StrictServerHandler) GetImportStatus(_ context.Context, _ openapi.GetImportStatusRequestObject) (openapi.GetImportStatusResponseObject, error) {
	return openapi.GetImportStatus501JSONResponse{
		NotImplementedJSONResponse: openapi.NotImplementedJSONResponse{
			Data: nil,
			Errors: &[]openapi.Error{{
				Code:    "not_implemented",
				Message: "このエンドポイントは未実装です",
			}},
		},
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
