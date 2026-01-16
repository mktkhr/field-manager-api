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
	Cursor   *entity.FieldCursor // カーソル(nilの場合は先頭から)
	PageSize int                 // 1ページあたりの件数
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

// CursorPaginationOutput はカーソルベースページネーション出力
type CursorPaginationOutput struct {
	NextCursor *string // 次ページ用カーソル(最終ページの場合はnil)
	HasMore    bool    // 次ページが存在するか
	PageSize   int     // ページサイズ
}

// ListFieldsOutput は圃場一覧取得ユースケースの出力
type ListFieldsOutput struct {
	Fields     []FieldOutput
	Pagination CursorPaginationOutput
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
	// limit + 1 を取得して次ページの存在を確認する
	limit := input.PageSize + 1

	// 圃場一覧取得
	fields, err := u.fieldQuery.ListByCursor(ctx, input.Cursor, limit)
	if err != nil {
		u.logger.Error("圃場一覧の取得に失敗しました",
			slog.String("error", err.Error()))
		return nil, err
	}

	// 次ページの存在確認
	hasMore := len(fields) > input.PageSize
	if hasMore {
		// 余分に取得した1件を除外
		fields = fields[:input.PageSize]
	}

	// 出力変換
	outputs := make([]FieldOutput, 0, len(fields))
	for _, field := range fields {
		output := u.toFieldOutput(field)
		outputs = append(outputs, output)
	}

	// 次ページ用カーソルの生成
	var nextCursor *string
	if hasMore && len(fields) > 0 {
		lastField := fields[len(fields)-1]
		cursor := entity.NewFieldCursor(lastField.CreatedAt, lastField.ID)
		encoded, err := cursor.Encode()
		if err != nil {
			u.logger.Error("カーソルのエンコードに失敗しました",
				slog.String("error", err.Error()))
			return nil, err
		}
		nextCursor = &encoded
	}

	return &ListFieldsOutput{
		Fields: outputs,
		Pagination: CursorPaginationOutput{
			NextCursor: nextCursor,
			HasMore:    hasMore,
			PageSize:   input.PageSize,
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

		// strideが0または座標が空の場合は空配列を設定
		if stride == 0 || len(coords) == 0 {
			output.Geometry = []Coordinate{}
		} else {
			numPoints := len(coords) / stride
			geometry := make([]Coordinate, 0, numPoints)
			for i := 0; i < numPoints; i++ {
				lng := coords[i*stride]   // X = 経度
				lat := coords[i*stride+1] // Y = 緯度
				geometry = append(geometry, Coordinate{Lat: lat, Lng: lng})
			}
			output.Geometry = geometry
		}
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
