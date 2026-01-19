package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/twpayne/go-geom"
)

type SearchedFieldTestSuite struct {
	suite.Suite
}

func TestSearchedFieldTestSuite(t *testing.T) {
	suite.Run(t, new(SearchedFieldTestSuite))
}

// TestNewSearchedField_Success_AllFieldsSet は全フィールドが設定されたSearchedFieldが正常に作成されることをテストする
func (s *SearchedFieldTestSuite) TestNewSearchedField_Success_AllFieldsSet() {
	id := uuid.New()
	name := "田んぼA"
	cityCode := "13101"
	soilTypeID := uuid.New()
	areaSqm := 1500.5

	polygon := geom.NewPolygon(geom.XY)
	coords := [][]geom.Coord{{{139.0, 35.0}, {139.1, 35.0}, {139.05, 35.1}, {139.0, 35.0}}}
	_, err := polygon.SetCoords(coords)
	require.NoError(s.T(), err, "ポリゴン座標の設定でエラーが発生してはならない")

	centroid := geom.NewPoint(geom.XY)
	_, err = centroid.SetCoords(geom.Coord{139.05, 35.033})
	require.NoError(s.T(), err, "重心座標の設定でエラーが発生してはならない")

	field := NewSearchedField(id, name, cityCode, &soilTypeID, &areaSqm, polygon, centroid)

	require.NotNil(s.T(), field, "SearchedFieldがnilであってはならない")
	require.Equal(s.T(), id, field.ID)
	require.Equal(s.T(), name, field.Name)
	require.Equal(s.T(), cityCode, field.CityCode)
	require.NotNil(s.T(), field.SoilTypeID)
	require.Equal(s.T(), soilTypeID, *field.SoilTypeID)
	require.NotNil(s.T(), field.AreaSqm)
	require.Equal(s.T(), areaSqm, *field.AreaSqm)
	require.NotNil(s.T(), field.Geometry)
	require.NotNil(s.T(), field.Centroid)
}

// TestNewSearchedField_Success_NullableFieldsNil はnullableフィールドがnilでも正常に作成されることをテストする
func (s *SearchedFieldTestSuite) TestNewSearchedField_Success_NullableFieldsNil() {
	id := uuid.New()
	name := "田んぼB"
	cityCode := "13102"

	field := NewSearchedField(id, name, cityCode, nil, nil, nil, nil)

	require.NotNil(s.T(), field, "SearchedFieldがnilであってはならない")
	require.Equal(s.T(), id, field.ID)
	require.Equal(s.T(), name, field.Name)
	require.Equal(s.T(), cityCode, field.CityCode)
	require.Nil(s.T(), field.SoilTypeID, "SoilTypeIDはnilであるべき")
	require.Nil(s.T(), field.AreaSqm, "AreaSqmはnilであるべき")
	require.Nil(s.T(), field.Geometry, "Geometryはnilであるべき")
	require.Nil(s.T(), field.Centroid, "Centroidはnilであるべき")
}
