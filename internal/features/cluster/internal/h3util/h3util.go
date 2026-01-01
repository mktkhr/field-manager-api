// Package h3util はH3関連のユーティリティを提供する
// このパッケージはcluster機能内に閉じており、他の機能からは使用しない
package h3util

import (
	"fmt"

	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/uber/h3-go/v4"
)

// BoundingBox はバウンディングボックスを表す
type BoundingBox struct {
	SWLat float64 // 南西端の緯度
	SWLng float64 // 南西端の経度
	NELat float64 // 北東端の緯度
	NELng float64 // 北東端の経度
}

// NewBoundingBox は新しいBoundingBoxを作成する
func NewBoundingBox(swLat, swLng, neLat, neLng float64) *BoundingBox {
	return &BoundingBox{
		SWLat: swLat,
		SWLng: swLng,
		NELat: neLat,
		NELng: neLng,
	}
}

// IsValid はBoundingBoxが有効かどうかを判定する
func (bb *BoundingBox) IsValid() bool {
	// 緯度の範囲チェック
	if bb.SWLat < -90 || bb.SWLat > 90 || bb.NELat < -90 || bb.NELat > 90 {
		return false
	}
	// 経度の範囲チェック
	if bb.SWLng < -180 || bb.SWLng > 180 || bb.NELng < -180 || bb.NELng > 180 {
		return false
	}
	// 南西が北東より北にある場合は無効
	if bb.SWLat > bb.NELat {
		return false
	}
	return true
}

// Contains は指定した緯度経度がBoundingBox内にあるかを判定する
func (bb *BoundingBox) Contains(lat, lng float64) bool {
	// 緯度のチェック
	if lat < bb.SWLat || lat > bb.NELat {
		return false
	}

	// 経度のチェック（日付変更線をまたぐケースを考慮）
	if bb.SWLng <= bb.NELng {
		// 通常のケース
		return lng >= bb.SWLng && lng <= bb.NELng
	}
	// 日付変更線をまたぐケース
	return lng >= bb.SWLng || lng <= bb.NELng
}

// ZoomToResolution はズームレベルからH3解像度を決定する
// res3, res5, res7, res9の4解像度を使用
func ZoomToResolution(zoom float64) entity.Resolution {
	switch {
	case zoom < 6:
		return entity.Res3
	case zoom < 10:
		return entity.Res5
	case zoom < 14:
		return entity.Res7
	default:
		return entity.Res9
	}
}

// CellToLatLng はH3セルの中心座標を取得する
func CellToLatLng(h3Index string) (lat, lng float64, err error) {
	cell := h3.CellFromString(h3Index)
	if !cell.IsValid() {
		return 0, 0, fmt.Errorf("無効なH3インデックス: %s", h3Index)
	}

	latLng, err := cell.LatLng()
	if err != nil {
		return 0, 0, fmt.Errorf("H3セルから座標の取得に失敗しました: %w", err)
	}
	return latLng.Lat, latLng.Lng, nil
}

// FilterCellsInBBox はBoundingBox内のセルをフィルタリングする
func FilterCellsInBBox(h3Indexes []string, bb *BoundingBox) ([]string, error) {
	if bb == nil {
		return h3Indexes, nil
	}

	filtered := make([]string, 0, len(h3Indexes))
	for _, h3Index := range h3Indexes {
		lat, lng, err := CellToLatLng(h3Index)
		if err != nil {
			// 無効なH3インデックスはスキップ
			continue
		}

		if bb.Contains(lat, lng) {
			filtered = append(filtered, h3Index)
		}
	}

	return filtered, nil
}

// CalculateCenterFromH3 はH3インデックスから中心座標を計算する
func CalculateCenterFromH3(h3Index string) (lat, lng float64, err error) {
	return CellToLatLng(h3Index)
}

// IsValidH3Index はH3インデックスが有効かどうかを判定する
func IsValidH3Index(h3Index string) bool {
	cell := h3.CellFromString(h3Index)
	return cell.IsValid()
}
