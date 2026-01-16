// Package presentation は圃場機能のHTTPハンドラーを提供する
package presentation

import (
	"context"
	"log/slog"

	"github.com/mktkhr/field-manager-api/internal/features/field/application/usecase"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
)

const (
	// DefaultPageSize はデフォルトのページサイズ
	DefaultPageSize = 20
	// MaxPageSize は最大ページサイズ
	MaxPageSize = 1000
)

// FieldHandler は圃場APIのハンドラー
type FieldHandler struct {
	listFieldsUC *usecase.ListFieldsUseCase
	logger       *slog.Logger
}

// NewFieldHandler はFieldHandlerを作成する
func NewFieldHandler(
	listFieldsUC *usecase.ListFieldsUseCase,
	logger *slog.Logger,
) *FieldHandler {
	return &FieldHandler{
		listFieldsUC: listFieldsUC,
		logger:       logger,
	}
}

// ListFields は圃場一覧を取得する
func (h *FieldHandler) ListFields(ctx context.Context, request openapi.ListFieldsRequestObject) (openapi.ListFieldsResponseObject, error) {
	params := request.Params

	// パラメータバリデーション
	cursor, pageSize, err := h.validateAndNormalizeParams(params)
	if err != nil {
		return openapi.ListFields400JSONResponse{
			BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "invalid_parameter",
					Message: err.Error(),
				}},
			},
		}, nil
	}

	// ユースケース実行
	output, err := h.listFieldsUC.Execute(ctx, usecase.ListFieldsInput{
		Cursor:   cursor,
		PageSize: pageSize,
	})
	if err != nil {
		h.logger.Error("圃場一覧取得に失敗しました",
			slog.String("error", err.Error()))
		return openapi.ListFields500JSONResponse{
			InternalServerErrorJSONResponse: openapi.InternalServerErrorJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "internal_error",
					Message: "圃場一覧の取得に失敗しました",
				}},
			},
		}, nil
	}

	// レスポンス変換
	fields := make([]openapi.Field, 0, len(output.Fields))
	for _, field := range output.Fields {
		fields = append(fields, h.toOpenAPIField(field))
	}

	return openapi.ListFields200JSONResponse{
		Data: &openapi.FieldListData{
			Fields: fields,
		},
		Meta: &openapi.ResponseMeta{
			Pagination: openapi.CursorPaginationMeta{
				NextCursor: output.Pagination.NextCursor,
				HasMore:    output.Pagination.HasMore,
				PageSize:   output.Pagination.PageSize,
			},
		},
		Errors: nil,
	}, nil
}

// validateAndNormalizeParams はパラメータをバリデーションして正規化する
func (h *FieldHandler) validateAndNormalizeParams(params openapi.ListFieldsParams) (cursor *entity.FieldCursor, pageSize int, err error) {
	// cursorバリデーション(オプション)
	if params.Cursor != nil && *params.Cursor != "" {
		cursor, err = entity.DecodeFieldCursor(*params.Cursor)
		if err != nil {
			return nil, 0, &ValidationError{
				Field:   "cursor",
				Message: "カーソルの形式が不正です",
			}
		}
	}

	// pageSizeバリデーション(オプション、デフォルト20、最大1000)
	if params.PageSize == nil {
		pageSize = DefaultPageSize
	} else {
		ps := *params.PageSize
		if ps < 1 {
			return nil, 0, &ValidationError{
				Field:   "pageSize",
				Message: "ページサイズは1以上を指定してください",
			}
		}
		if ps > MaxPageSize {
			return nil, 0, &ValidationError{
				Field:   "pageSize",
				Message: "ページサイズは1000以下を指定してください",
			}
		}
		pageSize = ps
	}

	return cursor, pageSize, nil
}

// toOpenAPIField はUseCaseの出力をOpenAPIの型に変換する
func (h *FieldHandler) toOpenAPIField(field usecase.FieldOutput) openapi.Field {
	// Geometry変換
	geometry := make([]openapi.Coordinate, 0, len(field.Geometry))
	for _, coord := range field.Geometry {
		geometry = append(geometry, openapi.Coordinate{
			Lat: coord.Lat,
			Lng: coord.Lng,
		})
	}

	return openapi.Field{
		Id:         field.ID,
		Name:       field.Name,
		CityCode:   field.CityCode,
		SoilTypeId: field.SoilTypeID,
		AreaSqm:    field.AreaSqm,
		Geometry:   geometry,
		Centroid: openapi.Coordinate{
			Lat: field.Centroid.Lat,
			Lng: field.Centroid.Lng,
		},
	}
}

// ValidationError はバリデーションエラー
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
