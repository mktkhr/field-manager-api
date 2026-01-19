// Package h3util は圃場検索機能のH3ユーティリティを提供する
package h3util

import (
	"fmt"

	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/domain/entity"
	"github.com/uber/h3-go/v4"
)

// BBoxCellCalculator はBBoxからH3セルを計算する
type BBoxCellCalculator struct{}

// NewBBoxCellCalculator は新しいBBoxCellCalculatorを作成する
func NewBBoxCellCalculator() *BBoxCellCalculator {
	return &BBoxCellCalculator{}
}

// CalculateCells はBBox内のH3セルを計算する
func (c *BBoxCellCalculator) CalculateCells(bbox *entity.BoundingBox, resolution int) ([]string, error) {
	if bbox == nil {
		return nil, fmt.Errorf("BoundingBoxがnilです")
	}

	// 解像度の範囲チェック
	if resolution < 0 || resolution > 15 {
		return nil, fmt.Errorf("H3解像度は0-15の範囲で指定してください: %d", resolution)
	}

	// BBoxの4頂点からポリゴンを作成(反時計回り)
	polygon := h3.GeoPolygon{
		GeoLoop: []h3.LatLng{
			h3.NewLatLng(bbox.SwLat(), bbox.SwLng()), // 南西
			h3.NewLatLng(bbox.SwLat(), bbox.NeLng()), // 南東
			h3.NewLatLng(bbox.NeLat(), bbox.NeLng()), // 北東
			h3.NewLatLng(bbox.NeLat(), bbox.SwLng()), // 北西
			h3.NewLatLng(bbox.SwLat(), bbox.SwLng()), // 閉じる(南西に戻る)
		},
	}

	// PolygonToCellsExperimentalでBBox内のH3セルを取得
	// ContainmentMode: ContainmentCenterで中心がBBox内にあるセルを取得
	cells, err := h3.PolygonToCellsExperimental(polygon, resolution, h3.ContainmentCenter)
	if err != nil {
		return nil, fmt.Errorf("H3セルの計算に失敗: %w", err)
	}

	// 重複排除用のマップ
	cellSet := make(map[string]struct{})
	for _, cell := range cells {
		cellSet[cell.String()] = struct{}{}
	}

	// BBoxが小さくてセルが取得できない場合、4隅と中心のセルを追加
	// これにより、BBoxがH3セルより小さい場合でも検索が機能する
	if len(cellSet) == 0 {
		cornerCells := c.calculateCornerCells(bbox, resolution)
		for _, cellStr := range cornerCells {
			cellSet[cellStr] = struct{}{}
		}
	}

	// map -> slice変換
	cellStrings := make([]string, 0, len(cellSet))
	for cellStr := range cellSet {
		cellStrings = append(cellStrings, cellStr)
	}

	return cellStrings, nil
}

// calculateCornerCells はBBoxの4隅と中心からH3セルを計算する
func (c *BBoxCellCalculator) calculateCornerCells(bbox *entity.BoundingBox, resolution int) []string {
	// 4隅 + 中心の5点
	points := []h3.LatLng{
		h3.NewLatLng(bbox.SwLat(), bbox.SwLng()),                                   // 南西
		h3.NewLatLng(bbox.SwLat(), bbox.NeLng()),                                   // 南東
		h3.NewLatLng(bbox.NeLat(), bbox.NeLng()),                                   // 北東
		h3.NewLatLng(bbox.NeLat(), bbox.SwLng()),                                   // 北西
		h3.NewLatLng((bbox.SwLat()+bbox.NeLat())/2, (bbox.SwLng()+bbox.NeLng())/2), // 中心
	}

	cellSet := make(map[string]struct{})
	for _, point := range points {
		cell, err := h3.LatLngToCell(point, resolution)
		if err != nil {
			continue
		}
		cellSet[cell.String()] = struct{}{}
	}

	cells := make([]string, 0, len(cellSet))
	for cellStr := range cellSet {
		cells = append(cells, cellStr)
	}
	return cells
}
