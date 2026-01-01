// Package presentation はクラスタリング機能のHTTPハンドラーを提供する
package presentation

import (
	"context"
	"log/slog"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/application/usecase"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
)

// ClusterHandler はクラスターAPIのハンドラー
type ClusterHandler struct {
	getClustersUC *usecase.GetClustersUseCase
	logger        *slog.Logger
}

// NewClusterHandler はClusterHandlerを作成する
func NewClusterHandler(
	getClustersUC *usecase.GetClustersUseCase,
	logger *slog.Logger,
) *ClusterHandler {
	return &ClusterHandler{
		getClustersUC: getClustersUC,
		logger:        logger,
	}
}

// GetClusters はクラスター一覧を取得する
func (h *ClusterHandler) GetClusters(ctx context.Context, request openapi.GetClustersRequestObject) (openapi.GetClustersResponseObject, error) {
	params := request.Params

	// パラメータのバリデーション
	if err := h.validateGetClustersParams(params); err != nil {
		return openapi.GetClusters400JSONResponse{
			Code:    "invalid_parameter",
			Message: err.Error(),
		}, nil
	}

	// ユースケース実行
	output, err := h.getClustersUC.Execute(ctx, usecase.GetClustersInput{
		Zoom:  params.Zoom,
		SWLat: params.SwLat,
		SWLng: params.SwLng,
		NELat: params.NeLat,
		NELng: params.NeLng,
	})
	if err != nil {
		h.logger.Error("クラスター取得に失敗しました",
			slog.String("error", err.Error()))
		return openapi.GetClusters500JSONResponse{
			Code:    "internal_error",
			Message: "クラスターの取得に失敗しました",
		}, nil
	}

	// レスポンス変換
	clusters := make([]openapi.Cluster, 0, len(output.Clusters))
	for _, cluster := range output.Clusters {
		clusters = append(clusters, openapi.Cluster{
			H3Index: cluster.H3Index,
			Lat:     cluster.Lat,
			Lng:     cluster.Lng,
			Count:   int(cluster.Count),
		})
	}

	return openapi.GetClusters200JSONResponse{
		Clusters: clusters,
		IsStale:  output.IsStale,
	}, nil
}

// validateGetClustersParams はリクエストパラメータをバリデーションする
func (h *ClusterHandler) validateGetClustersParams(params openapi.GetClustersParams) error {
	// zoomのバリデーション
	if params.Zoom < 1.0 || params.Zoom > 22.0 {
		return &ValidationError{Field: "zoom", Message: "ズームレベルは1.0から22.0の範囲で指定してください"}
	}

	// 緯度のバリデーション
	if params.SwLat < -90 || params.SwLat > 90 {
		return &ValidationError{Field: "sw_lat", Message: "南西端の緯度は-90から90の範囲で指定してください"}
	}
	if params.NeLat < -90 || params.NeLat > 90 {
		return &ValidationError{Field: "ne_lat", Message: "北東端の緯度は-90から90の範囲で指定してください"}
	}

	// 経度のバリデーション
	if params.SwLng < -180 || params.SwLng > 180 {
		return &ValidationError{Field: "sw_lng", Message: "南西端の経度は-180から180の範囲で指定してください"}
	}
	if params.NeLng < -180 || params.NeLng > 180 {
		return &ValidationError{Field: "ne_lng", Message: "北東端の経度は-180から180の範囲で指定してください"}
	}

	// 南西端が北東端より北にある場合はエラー
	if params.SwLat > params.NeLat {
		return &ValidationError{Field: "sw_lat", Message: "南西端の緯度は北東端の緯度より小さくしてください"}
	}

	return nil
}

// ValidationError はバリデーションエラー
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
