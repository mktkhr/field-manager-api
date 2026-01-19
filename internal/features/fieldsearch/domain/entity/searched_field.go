package entity

import (
	"github.com/google/uuid"
	"github.com/twpayne/go-geom"
)

// SearchedField は検索結果の圃場エンティティ
type SearchedField struct {
	ID         uuid.UUID
	Name       string
	CityCode   string
	SoilTypeID *uuid.UUID
	AreaSqm    *float64
	Geometry   *geom.Polygon
	Centroid   *geom.Point
}

// NewSearchedField は新しいSearchedFieldを作成する
func NewSearchedField(
	id uuid.UUID,
	name string,
	cityCode string,
	soilTypeID *uuid.UUID,
	areaSqm *float64,
	geometry *geom.Polygon,
	centroid *geom.Point,
) *SearchedField {
	return &SearchedField{
		ID:         id,
		Name:       name,
		CityCode:   cityCode,
		SoilTypeID: soilTypeID,
		AreaSqm:    areaSqm,
		Geometry:   geometry,
		Centroid:   centroid,
	}
}
