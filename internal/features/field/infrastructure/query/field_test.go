package query

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
)

// FieldQueryTestSuite はFieldQueryの単体テストスイート
type FieldQueryTestSuite struct {
	suite.Suite
	query *fieldQuery
}

func (s *FieldQueryTestSuite) SetupTest() {
	// dbはnilでも、parseGeometryやtoEntityのテストは可能
	s.query = &fieldQuery{}
}

func TestFieldQuerySuite(t *testing.T) {
	suite.Run(t, new(FieldQueryTestSuite))
}

// TestFieldQuery_parseGeometry_Success はWKBパースが正しく動作することをテスト
func (s *FieldQueryTestSuite) TestFieldQuery_parseGeometry_Success() {
	// テスト用のPolygon作成
	polygon := geom.NewPolygon(geom.XY)
	coords := [][]geom.Coord{{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}}
	_, err := polygon.SetCoords(coords)
	require.NoError(s.T(), err, "ポリゴン座標設定に失敗")

	// WKBにエンコード
	wkbBytes, err := wkb.Marshal(polygon, wkb.NDR)
	require.NoError(s.T(), err, "WKBエンコードに失敗")

	// parseGeometryでデコード
	result, err := s.query.parseGeometry(wkbBytes)
	require.NoError(s.T(), err, "parseGeometry実行時にエラーが発生")
	require.NotNil(s.T(), result, "結果がnilです")

	// Polygon型であることを確認
	resultPolygon, ok := result.(*geom.Polygon)
	require.True(s.T(), ok, "結果がPolygon型ではありません")
	require.Equal(s.T(), 5, resultPolygon.NumCoords(), "頂点数が期待値と異なります")
}

// TestFieldQuery_parseGeometry_Point はPoint型のWKBパースをテスト
func (s *FieldQueryTestSuite) TestFieldQuery_parseGeometry_Point() {
	// テスト用のPoint作成
	point := geom.NewPoint(geom.XY)
	_, err := point.SetCoords(geom.Coord{139.05, 35.05})
	require.NoError(s.T(), err, "ポイント座標設定に失敗")

	// WKBにエンコード
	wkbBytes, err := wkb.Marshal(point, wkb.NDR)
	require.NoError(s.T(), err, "WKBエンコードに失敗")

	// parseGeometryでデコード
	result, err := s.query.parseGeometry(wkbBytes)
	require.NoError(s.T(), err, "parseGeometry実行時にエラーが発生")
	require.NotNil(s.T(), result, "結果がnilです")

	// Point型であることを確認
	resultPoint, ok := result.(*geom.Point)
	require.True(s.T(), ok, "結果がPoint型ではありません")
	require.InDelta(s.T(), 139.05, resultPoint.X(), 0.001, "X座標が一致しません")
	require.InDelta(s.T(), 35.05, resultPoint.Y(), 0.001, "Y座標が一致しません")
}

// TestFieldQuery_parseGeometry_Nil はnilデータでnilが返ることをテスト
func (s *FieldQueryTestSuite) TestFieldQuery_parseGeometry_Nil() {
	result, err := s.query.parseGeometry(nil)
	require.NoError(s.T(), err, "parseGeometry(nil)実行時にエラーが発生")
	require.Nil(s.T(), result, "結果がnilではありません")
}

// TestFieldQuery_parseGeometry_EmptyBytes は空のバイト列でnilが返ることをテスト
func (s *FieldQueryTestSuite) TestFieldQuery_parseGeometry_EmptyBytes() {
	result, err := s.query.parseGeometry([]byte{})
	require.NoError(s.T(), err, "parseGeometry(空バイト)実行時にエラーが発生")
	require.Nil(s.T(), result, "結果がnilではありません")
}

// TestFieldQuery_parseGeometry_InvalidData は不正なデータでエラーが返ることをテスト
func (s *FieldQueryTestSuite) TestFieldQuery_parseGeometry_InvalidData() {
	_, err := s.query.parseGeometry([]byte{0x01, 0x02, 0x03})
	require.Error(s.T(), err, "不正なWKBデータでエラーが返されませんでした")
}

// TestFieldQuery_parseGeometry_InvalidType は不正な型でエラーが返ることをテスト
func (s *FieldQueryTestSuite) TestFieldQuery_parseGeometry_InvalidType() {
	_, err := s.query.parseGeometry("invalid string")
	require.Error(s.T(), err, "不正な型でエラーが返されませんでした")
}

// TestFieldQuery_toEntity_Success はSQLCモデルからエンティティへの変換が正しく動作することをテスト
func (s *FieldQueryTestSuite) TestFieldQuery_toEntity_Success() {
	fieldID := uuid.New()
	soilTypeID := uuid.New()
	areaSqm := 10000.0

	// テスト用のPolygon作成
	polygon := geom.NewPolygon(geom.XY)
	coords := [][]geom.Coord{{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}}
	_, err := polygon.SetCoords(coords)
	require.NoError(s.T(), err, "ポリゴン座標設定に失敗")

	// テスト用のPoint作成
	point := geom.NewPoint(geom.XY)
	_, err = point.SetCoords(geom.Coord{139.05, 35.05})
	require.NoError(s.T(), err, "ポイント座標設定に失敗")

	// WKBにエンコード
	geometryWKB, err := wkb.Marshal(polygon, wkb.NDR)
	require.NoError(s.T(), err, "ジオメトリWKBエンコードに失敗")
	centroidWKB, err := wkb.Marshal(point, wkb.NDR)
	require.NoError(s.T(), err, "重心WKBエンコードに失敗")

	row := &sqlc.Field{
		ID:       fieldID,
		Geometry: geometryWKB,
		Centroid: centroidWKB,
		AreaSqm:  &areaSqm,
		CityCode: "12345",
		Name:     "テスト圃場",
		SoilTypeID: uuid.NullUUID{
			UUID:  soilTypeID,
			Valid: true,
		},
	}

	entity, err := s.query.toEntity(row)
	require.NoError(s.T(), err, "toEntity実行時にエラーが発生")
	require.NotNil(s.T(), entity, "エンティティがnilです")
	require.Equal(s.T(), fieldID, entity.ID, "IDが一致しません")
	require.Equal(s.T(), "12345", entity.CityCode, "市区町村コードが一致しません")
	require.Equal(s.T(), "テスト圃場", entity.Name, "圃場名が一致しません")
	require.NotNil(s.T(), entity.SoilTypeID, "SoilTypeIDがnilです")
	require.Equal(s.T(), soilTypeID, *entity.SoilTypeID, "SoilTypeIDが一致しません")
	require.NotNil(s.T(), entity.AreaSqm, "AreaSqmがnilです")
	require.Equal(s.T(), areaSqm, *entity.AreaSqm, "AreaSqmが一致しません")
	require.NotNil(s.T(), entity.Geometry, "Geometryがnilです")
	require.NotNil(s.T(), entity.Centroid, "Centroidがnilです")
}

// TestFieldQuery_toEntity_Nil はnilでnilが返ることをテスト
func (s *FieldQueryTestSuite) TestFieldQuery_toEntity_Nil() {
	entity, err := s.query.toEntity(nil)
	require.NoError(s.T(), err, "toEntity(nil)実行時にエラーが発生")
	require.Nil(s.T(), entity, "エンティティがnilではありません")
}

// TestFieldQuery_toEntity_WithNullValues はnull値を含む場合の変換をテスト
func (s *FieldQueryTestSuite) TestFieldQuery_toEntity_WithNullValues() {
	fieldID := uuid.New()

	row := &sqlc.Field{
		ID:         fieldID,
		Geometry:   nil,
		Centroid:   nil,
		AreaSqm:    nil,
		CityCode:   "12345",
		Name:       "テスト圃場",
		SoilTypeID: uuid.NullUUID{Valid: false},
		CreatedAt:  pgtype.Timestamptz{Valid: false},
		UpdatedAt:  pgtype.Timestamptz{Valid: false},
		CreatedBy:  uuid.NullUUID{Valid: false},
		UpdatedBy:  uuid.NullUUID{Valid: false},
	}

	entity, err := s.query.toEntity(row)
	require.NoError(s.T(), err, "toEntity実行時にエラーが発生")
	require.NotNil(s.T(), entity, "エンティティがnilです")
	require.Equal(s.T(), fieldID, entity.ID, "IDが一致しません")
	require.Nil(s.T(), entity.Geometry, "Geometryがnilではありません")
	require.Nil(s.T(), entity.Centroid, "Centroidがnilではありません")
	require.Nil(s.T(), entity.AreaSqm, "AreaSqmがnilではありません")
	require.Nil(s.T(), entity.SoilTypeID, "SoilTypeIDがnilではありません")
}
