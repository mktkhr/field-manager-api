// Package usecase は圃場検索機能のユースケースを提供する
package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/application/query"
	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/domain/entity"
)

// H3CellCalculator はBBoxからH3セルを計算するインターフェース
type H3CellCalculator interface {
	// CalculateCells はBBox内のH3セルを計算する
	CalculateCells(bbox *entity.BoundingBox, resolution int) ([]string, error)
}

// SearchFieldsInput は圃場検索ユースケースの入力
type SearchFieldsInput struct {
	SwLat float64
	SwLng float64
	NeLat float64
	NeLng float64
}

// Coordinate は座標を表す
type Coordinate struct {
	Lat float64
	Lng float64
}

// SearchedFieldOutput は検索結果の圃場出力
type SearchedFieldOutput struct {
	ID         uuid.UUID
	Name       string
	CityCode   string
	SoilTypeID *uuid.UUID
	AreaSqm    *float64
	Geometry   []Coordinate // ポリゴンの頂点座標
	Centroid   Coordinate   // 重心座標
}

// SearchFieldsOutput は圃場検索ユースケースの出力
type SearchFieldsOutput struct {
	Fields []SearchedFieldOutput
}

// SearchFieldsUseCase は圃場検索ユースケース
type SearchFieldsUseCase struct {
	searchQuery  query.FieldSearchQuery
	h3Calculator H3CellCalculator
	logger       *slog.Logger
}

// NewSearchFieldsUseCase はSearchFieldsUseCaseを作成する
func NewSearchFieldsUseCase(
	searchQuery query.FieldSearchQuery,
	h3Calculator H3CellCalculator,
	logger *slog.Logger,
) *SearchFieldsUseCase {
	return &SearchFieldsUseCase{
		searchQuery:  searchQuery,
		h3Calculator: h3Calculator,
		logger:       logger,
	}
}

// Execute は圃場検索を実行する
func (u *SearchFieldsUseCase) Execute(ctx context.Context, input SearchFieldsInput) (*SearchFieldsOutput, error) {
	// BoundingBoxの作成
	bbox, err := entity.NewBoundingBox(input.SwLat, input.SwLng, input.NeLat, input.NeLng)
	if err != nil {
		u.logger.Warn("BoundingBoxの作成に失敗しました",
			slog.String("error", err.Error()),
			slog.Float64("swLat", input.SwLat),
			slog.Float64("swLng", input.SwLng),
			slog.Float64("neLat", input.NeLat),
			slog.Float64("neLng", input.NeLng))
		return nil, err
	}

	// 最適なH3解像度を決定
	resolution := bbox.OptimalH3Resolution()

	// H3セルを計算
	h3Cells, err := u.h3Calculator.CalculateCells(bbox, resolution)
	if err != nil {
		u.logger.Error("H3セルの計算に失敗しました",
			slog.String("error", err.Error()),
			slog.Int("resolution", resolution))
		return nil, err
	}

	// H3セル数上限チェック
	if len(h3Cells) > entity.MaxH3Cells {
		u.logger.Warn("H3セル数が上限を超えています",
			slog.Int("cellCount", len(h3Cells)),
			slog.Int("maxCells", entity.MaxH3Cells))
		return nil, fmt.Errorf("H3セル数が上限を超えています: %d(上限: %d)", len(h3Cells), entity.MaxH3Cells)
	}

	u.logger.Debug("H3セル計算完了",
		slog.Int("resolution", resolution),
		slog.Int("cellCount", len(h3Cells)))

	// 検索実行
	fields, err := u.searchQuery.SearchByBBox(ctx, bbox, h3Cells, resolution)
	if err != nil {
		u.logger.Error("圃場検索に失敗しました",
			slog.String("error", err.Error()))
		return nil, err
	}

	// 出力変換
	outputs := make([]SearchedFieldOutput, 0, len(fields))
	for _, field := range fields {
		output := u.toOutput(field)
		outputs = append(outputs, output)
	}

	u.logger.Debug("圃場検索完了",
		slog.Int("resultCount", len(outputs)))

	return &SearchFieldsOutput{
		Fields: outputs,
	}, nil
}

// toOutput はエンティティを出力に変換する
func (u *SearchFieldsUseCase) toOutput(field *entity.SearchedField) SearchedFieldOutput {
	output := SearchedFieldOutput{
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
