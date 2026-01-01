package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/twpayne/go-geom"
	"github.com/uber/h3-go/v4"
)

// Field は圃場エンティティ
type Field struct {
	ID          uuid.UUID
	Geometry    *geom.Polygon
	Centroid    *geom.Point
	AreaSqm     *float64
	H3IndexRes3 *string
	H3IndexRes5 *string
	H3IndexRes7 *string
	H3IndexRes9 *string
	CityCode    string
	Name        string
	SoilTypeID  *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
}

// NewField は新しいFieldを作成する
func NewField(id uuid.UUID, cityCode string) *Field {
	now := time.Now()
	return &Field{
		ID:        id,
		CityCode:  cityCode,
		Name:      "名称不明",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetGeometry はジオメトリを設定し、関連する値も計算する
func (f *Field) SetGeometry(polygon *geom.Polygon) error {
	f.Geometry = polygon

	// Centroidを計算
	if polygon != nil {
		centroid := CalculateCentroid(polygon)
		f.Centroid = centroid

		// H3インデックスを計算
		if centroid != nil {
			lat := centroid.Y()
			lng := centroid.X()
			if err := f.CalculateH3Indexes(lat, lng); err != nil {
				return err
			}
		}
	}
	return nil
}

// CalculateH3Indexes はH3インデックスを計算する
func (f *Field) CalculateH3Indexes(lat, lng float64) error {
	latLng := h3.NewLatLng(lat, lng)

	cell3, err := h3.LatLngToCell(latLng, 3)
	if err != nil {
		return fmt.Errorf("H3インデックス(res3)の計算に失敗: %w", err)
	}
	s3 := cell3.String()
	f.H3IndexRes3 = &s3

	cell5, err := h3.LatLngToCell(latLng, 5)
	if err != nil {
		return fmt.Errorf("H3インデックス(res5)の計算に失敗: %w", err)
	}
	s5 := cell5.String()
	f.H3IndexRes5 = &s5

	cell7, err := h3.LatLngToCell(latLng, 7)
	if err != nil {
		return fmt.Errorf("H3インデックス(res7)の計算に失敗: %w", err)
	}
	s7 := cell7.String()
	f.H3IndexRes7 = &s7

	cell9, err := h3.LatLngToCell(latLng, 9)
	if err != nil {
		return fmt.Errorf("H3インデックス(res9)の計算に失敗: %w", err)
	}
	s9 := cell9.String()
	f.H3IndexRes9 = &s9

	return nil
}

// SetSoilType は土壌タイプIDを設定する
func (f *Field) SetSoilType(soilTypeID uuid.UUID) {
	f.SoilTypeID = &soilTypeID
}

// CalculateCentroid はポリゴンの重心を計算する
func CalculateCentroid(polygon *geom.Polygon) *geom.Point {
	if polygon == nil || polygon.NumCoords() == 0 {
		return nil
	}

	coords := polygon.FlatCoords()
	stride := polygon.Stride()
	numPoints := len(coords) / stride

	if numPoints == 0 {
		return nil
	}

	var sumX, sumY float64
	for i := 0; i < numPoints; i++ {
		sumX += coords[i*stride]
		sumY += coords[i*stride+1]
	}

	centroid := geom.NewPoint(geom.XY)
	if _, err := centroid.SetCoords(geom.Coord{sumX / float64(numPoints), sumY / float64(numPoints)}); err != nil {
		return nil
	}
	return centroid
}

// ConvertLinearPolygonToPolygon はLinearPolygon(wagri形式)をPolygonに変換する
func ConvertLinearPolygonToPolygon(coordinates [][]float64) (*geom.Polygon, error) {
	if len(coordinates) == 0 {
		return nil, nil
	}

	coords := make([]geom.Coord, len(coordinates))
	for i, coord := range coordinates {
		if len(coord) < 2 {
			continue
		}
		coords[i] = geom.Coord{coord[0], coord[1]}
	}

	// 座標が閉じていない場合、最初の点を末尾に追加
	if len(coords) > 0 && !coordsEqual(coords[0], coords[len(coords)-1]) {
		coords = append(coords, coords[0])
	}

	polygon := geom.NewPolygon(geom.XY)
	if _, err := polygon.SetCoords([][]geom.Coord{coords}); err != nil {
		return nil, err
	}

	return polygon, nil
}

// coordsEqual は2つの座標が等しいかどうかを判定する
func coordsEqual(a, b geom.Coord) bool {
	if len(a) < 2 || len(b) < 2 {
		return false
	}
	return a[0] == b[0] && a[1] == b[1]
}
