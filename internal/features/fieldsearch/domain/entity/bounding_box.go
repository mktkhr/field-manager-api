package entity

import (
	"fmt"
	"math"
)

const (
	// MaxBBoxAreaKm2 はBBoxの最大面積(km²)
	// 約31km × 31kmの範囲に相当
	MaxBBoxAreaKm2 = 1000.0

	// MaxH3Cells はH3セルの最大数
	// これを超える場合はクエリ効率が低下するためエラーとする
	MaxH3Cells = 100
)

// BoundingBox は検索範囲を表すValue Object
type BoundingBox struct {
	swLat float64
	swLng float64
	neLat float64
	neLng float64
}

// NewBoundingBox は新しいBoundingBoxを作成する
func NewBoundingBox(swLat, swLng, neLat, neLng float64) (*BoundingBox, error) {
	if err := validateLatitude(swLat, "swLat"); err != nil {
		return nil, err
	}
	if err := validateLatitude(neLat, "neLat"); err != nil {
		return nil, err
	}
	if err := validateLongitude(swLng, "swLng"); err != nil {
		return nil, err
	}
	if err := validateLongitude(neLng, "neLng"); err != nil {
		return nil, err
	}

	// 南西端が北東端より北にある場合はエラー
	if swLat > neLat {
		return nil, fmt.Errorf("南西端の緯度(%f)は北東端の緯度(%f)より小さくなければなりません", swLat, neLat)
	}

	bbox := &BoundingBox{
		swLat: swLat,
		swLng: swLng,
		neLat: neLat,
		neLng: neLng,
	}

	// 面積チェック
	area := bbox.AreaKm2()
	if area > MaxBBoxAreaKm2 {
		return nil, fmt.Errorf("BBox面積が上限を超えています: %.2f km²(上限: %.0f km²)", area, MaxBBoxAreaKm2)
	}

	return bbox, nil
}

// SwLat は南西端の緯度を返す
func (b *BoundingBox) SwLat() float64 {
	return b.swLat
}

// SwLng は南西端の経度を返す
func (b *BoundingBox) SwLng() float64 {
	return b.swLng
}

// NeLat は北東端の緯度を返す
func (b *BoundingBox) NeLat() float64 {
	return b.neLat
}

// NeLng は北東端の経度を返す
func (b *BoundingBox) NeLng() float64 {
	return b.neLng
}

// DiagonalDistanceKm は対角線の距離(km)を計算する
// Haversine公式を使用
func (b *BoundingBox) DiagonalDistanceKm() float64 {
	const earthRadiusKm = 6371.0

	lat1 := b.swLat * math.Pi / 180.0
	lat2 := b.neLat * math.Pi / 180.0
	dLat := (b.neLat - b.swLat) * math.Pi / 180.0
	dLng := (b.neLng - b.swLng) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// AreaKm2 はBBoxの面積(km²)を概算する
// 緯度による経度方向の縮小を考慮した平面近似
func (b *BoundingBox) AreaKm2() float64 {
	const kmPerDegree = 111.0 // 緯度1度 ≈ 111km

	latDiff := b.neLat - b.swLat
	lngDiff := b.neLng - b.swLng
	avgLat := (b.neLat + b.swLat) / 2.0

	latKm := latDiff * kmPerDegree
	// 経度方向は緯度によって縮小(cos補正)
	lngKm := lngDiff * kmPerDegree * math.Cos(avgLat*math.Pi/180.0)

	return latKm * lngKm
}

// OptimalH3Resolution はBBoxサイズに基づいて最適なH3解像度を返す
func (b *BoundingBox) OptimalH3Resolution() int {
	distance := b.DiagonalDistanceKm()

	switch {
	case distance > 300:
		return 3
	case distance > 30:
		return 5
	case distance > 3:
		return 7
	default:
		return 9
	}
}

// validateLatitude は緯度の範囲をチェックする
func validateLatitude(lat float64, name string) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("%sは-90から90の範囲でなければなりません(現在値: %f)", name, lat)
	}
	return nil
}

// validateLongitude は経度の範囲をチェックする
func validateLongitude(lng float64, name string) error {
	if lng < -180 || lng > 180 {
		return fmt.Errorf("%sは-180から180の範囲でなければなりません(現在値: %f)", name, lng)
	}
	return nil
}
