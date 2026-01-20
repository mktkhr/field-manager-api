// Package presentation は圃場検索機能のHTTPハンドラーを提供する
package presentation

import (
	"context"
	"log/slog"
	"strings"

	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/application/usecase"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
)

// FieldSearchHandler は圃場検索APIのハンドラー
type FieldSearchHandler struct {
	searchFieldsUC *usecase.SearchFieldsUseCase
	logger         *slog.Logger
}

// NewFieldSearchHandler はFieldSearchHandlerを作成する
func NewFieldSearchHandler(
	searchFieldsUC *usecase.SearchFieldsUseCase,
	logger *slog.Logger,
) *FieldSearchHandler {
	return &FieldSearchHandler{
		searchFieldsUC: searchFieldsUC,
		logger:         logger,
	}
}

// SearchFields は緯度経度範囲による圃場検索を実行する
func (h *FieldSearchHandler) SearchFields(ctx context.Context, request openapi.SearchFieldsRequestObject) (openapi.SearchFieldsResponseObject, error) {
	params := request.Params

	// ユースケース実行
	output, err := h.searchFieldsUC.Execute(ctx, usecase.SearchFieldsInput{
		SwLat: params.SwLat,
		SwLng: params.SwLng,
		NeLat: params.NeLat,
		NeLng: params.NeLng,
	})
	if err != nil {
		// バリデーションエラーかビジネスロジックエラーかを判別
		h.logger.Warn("圃場検索に失敗しました",
			slog.String("error", err.Error()),
			slog.Float64("swLat", float64(params.SwLat)),
			slog.Float64("swLng", float64(params.SwLng)),
			slog.Float64("neLat", float64(params.NeLat)),
			slog.Float64("neLng", float64(params.NeLng)))

		// 座標の検証エラーは400を返す
		if isValidationError(err) {
			return openapi.SearchFields400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: err.Error(),
					}},
				},
			}, nil
		}

		// その他のエラーは500を返す
		return openapi.SearchFields500JSONResponse{
			InternalServerErrorJSONResponse: openapi.InternalServerErrorJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "internal_error",
					Message: "圃場検索に失敗しました",
				}},
			},
		}, nil
	}

	// レスポンス変換
	fields := make([]openapi.SearchedField, 0, len(output.Fields))
	for _, field := range output.Fields {
		fields = append(fields, h.toOpenAPISearchedField(field))
	}

	return openapi.SearchFields200JSONResponse{
		Data: &openapi.FieldSearchData{
			Fields: fields,
		},
		Errors: nil,
	}, nil
}

// toOpenAPISearchedField はUseCaseの出力をOpenAPIの型に変換する
func (h *FieldSearchHandler) toOpenAPISearchedField(field usecase.SearchedFieldOutput) openapi.SearchedField {
	// Geometry変換
	geometry := make([]openapi.SearchedCoordinate, 0, len(field.Geometry))
	for _, coord := range field.Geometry {
		geometry = append(geometry, openapi.SearchedCoordinate{
			Lat: coord.Lat,
			Lng: coord.Lng,
		})
	}

	return openapi.SearchedField{
		Id:         field.ID,
		Name:       field.Name,
		CityCode:   field.CityCode,
		SoilTypeId: field.SoilTypeID,
		AreaSqm:    field.AreaSqm,
		Geometry:   geometry,
		Centroid: openapi.SearchedCoordinate{
			Lat: field.Centroid.Lat,
			Lng: field.Centroid.Lng,
		},
	}
}

// isValidationError はエラーがバリデーションエラーかどうかを判定する
func isValidationError(err error) bool {
	// BoundingBox作成時の検証エラーは特定の文字列を含む
	errMsg := err.Error()
	validationKeywords := []string{
		"swLat",
		"swLng",
		"neLat",
		"neLng",
		"南西端",
		"北東端",
		"-90",
		"90",
		"-180",
		"180",
		"面積",
		"H3セル数",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	return false
}
