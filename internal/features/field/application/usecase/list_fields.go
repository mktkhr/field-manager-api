// Package usecase は圃場機能のユースケースを提供する
package usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/field/application/query"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
)

// ListFieldsInput は圃場一覧取得ユースケースの入力
type ListFieldsInput struct {
	Page     int // ページ番号(1始まり)
	PageSize int // 1ページあたりの件数
}

// Coordinate は座標を表す
type Coordinate struct {
	Lat float64
	Lng float64
}

// FieldOutput は圃場出力
type FieldOutput struct {
	ID         uuid.UUID
	Name       string
	CityCode   string
	SoilTypeID *uuid.UUID
	AreaSqm    *float64
	Geometry   []Coordinate // ポリゴンの頂点座標
	Centroid   Coordinate   // 重心座標
}

// PaginationOutput はページネーション出力
type PaginationOutput struct {
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// ListFieldsOutput は圃場一覧取得ユースケースの出力
type ListFieldsOutput struct {
	Fields     []FieldOutput
	Pagination PaginationOutput
}

// ListFieldsUseCase は圃場一覧取得ユースケース
type ListFieldsUseCase struct {
	fieldQuery query.FieldQuery
	logger     *slog.Logger
}

// NewListFieldsUseCase はListFieldsUseCaseを作成する
func NewListFieldsUseCase(
	fieldQuery query.FieldQuery,
	logger *slog.Logger,
) *ListFieldsUseCase {
	return &ListFieldsUseCase{
		fieldQuery: fieldQuery,
		logger:     logger,
	}
}

// Execute は圃場一覧取得を実行する
func (u *ListFieldsUseCase) Execute(ctx context.Context, input ListFieldsInput) (*ListFieldsOutput, error) {
	// ページネーション計算
	limit := int32(input.PageSize)
	offset := int32((input.Page - 1) * input.PageSize)

	// 総件数取得
	total, err := u.fieldQuery.Count(ctx)
	if err != nil {
		u.logger.Error("圃場の総数取得に失敗しました",
			slog.String("error", err.Error()))
		return nil, err
	}

	// 圃場一覧取得
	fields, err := u.fieldQuery.List(ctx, limit, offset)
	if err != nil {
		u.logger.Error("圃場一覧の取得に失敗しました",
			slog.String("error", err.Error()))
		return nil, err
	}

	// 出力変換
	outputs := make([]FieldOutput, 0, len(fields))
	for _, field := range fields {
		output := u.toFieldOutput(field)
		outputs = append(outputs, output)
	}

	// 総ページ数計算
	totalPages := int(total) / input.PageSize
	if int(total)%input.PageSize > 0 {
		totalPages++
	}
	// 0件の場合は総ページ数を0にする
	if total == 0 {
		totalPages = 0
	}

	return &ListFieldsOutput{
		Fields: outputs,
		Pagination: PaginationOutput{
			Total:      total,
			Page:       input.Page,
			PageSize:   input.PageSize,
			TotalPages: totalPages,
		},
	}, nil
}

// toFieldOutput はエンティティを出力に変換する
func (u *ListFieldsUseCase) toFieldOutput(field *entity.Field) FieldOutput {
	output := FieldOutput{
		ID:         field.ID,
		Name:       field.Name,
		CityCode:   field.CityCode,
		SoilTypeID: field.SoilTypeID,
		AreaSqm:    field.AreaSqm,
	}

	// Geometry変換(Polygon -> []Coordinate)
	if field.Geometry != nil {
		coords := field.Geometry.FlatCoords()
		stride := field.Geometry.Stride()
		numPoints := len(coords) / stride

		geometry := make([]Coordinate, 0, numPoints)
		for i := 0; i < numPoints; i++ {
			lng := coords[i*stride]   // X = 経度
			lat := coords[i*stride+1] // Y = 緯度
			geometry = append(geometry, Coordinate{Lat: lat, Lng: lng})
		}
		output.Geometry = geometry
	}

	// Centroid変換(Point -> Coordinate)
	if field.Centroid != nil {
		output.Centroid = Coordinate{
			Lat: field.Centroid.Y(), // Y = 緯度
			Lng: field.Centroid.X(), // X = 経度
		}
	}

	return output
}
