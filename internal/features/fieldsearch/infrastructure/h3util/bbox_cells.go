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

	// Cell -> string変換
	cellStrings := make([]string, len(cells))
	for i, cell := range cells {
		cellStrings[i] = cell.String()
	}

	return cellStrings, nil
}
