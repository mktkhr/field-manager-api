package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BoundingBoxTestSuite struct {
	suite.Suite
}

func TestBoundingBoxTestSuite(t *testing.T) {
	suite.Run(t, new(BoundingBoxTestSuite))
}

// TestNewBoundingBox_Success_ValidCoordinates は有効な座標でBoundingBoxが正常に作成されることをテストする
func (s *BoundingBoxTestSuite) TestNewBoundingBox_Success_ValidCoordinates() {
	bbox, err := NewBoundingBox(35.6762, 139.6503, 35.7295, 139.7673)

	require.NoError(s.T(), err, "有効な座標でエラーが発生してはならない")
	require.NotNil(s.T(), bbox, "BoundingBoxがnilであってはならない")
	require.Equal(s.T(), 35.6762, bbox.SwLat())
	require.Equal(s.T(), 139.6503, bbox.SwLng())
	require.Equal(s.T(), 35.7295, bbox.NeLat())
	require.Equal(s.T(), 139.7673, bbox.NeLng())
}

// TestNewBoundingBox_ValidationError_SwLatOutOfRange は南西端緯度が範囲外の場合にエラーを返すことをテストする
func (s *BoundingBoxTestSuite) TestNewBoundingBox_ValidationError_SwLatOutOfRange() {
	// 緯度が-90未満
	_, err := NewBoundingBox(-91, 139.6503, 35.7295, 139.7673)
	require.Error(s.T(), err, "南西端緯度が-90未満の場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "swLat")

	// 緯度が90超過
	_, err = NewBoundingBox(91, 139.6503, 35.7295, 139.7673)
	require.Error(s.T(), err, "南西端緯度が90超過の場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "swLat")
}

// TestNewBoundingBox_ValidationError_NeLatOutOfRange は北東端緯度が範囲外の場合にエラーを返すことをテストする
func (s *BoundingBoxTestSuite) TestNewBoundingBox_ValidationError_NeLatOutOfRange() {
	// 緯度が-90未満
	_, err := NewBoundingBox(35.6762, 139.6503, -91, 139.7673)
	require.Error(s.T(), err, "北東端緯度が-90未満の場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "neLat")

	// 緯度が90超過
	_, err = NewBoundingBox(35.6762, 139.6503, 91, 139.7673)
	require.Error(s.T(), err, "北東端緯度が90超過の場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "neLat")
}

// TestNewBoundingBox_ValidationError_SwLngOutOfRange は南西端経度が範囲外の場合にエラーを返すことをテストする
func (s *BoundingBoxTestSuite) TestNewBoundingBox_ValidationError_SwLngOutOfRange() {
	// 経度が-180未満
	_, err := NewBoundingBox(35.6762, -181, 35.7295, 139.7673)
	require.Error(s.T(), err, "南西端経度が-180未満の場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "swLng")

	// 経度が180超過
	_, err = NewBoundingBox(35.6762, 181, 35.7295, 139.7673)
	require.Error(s.T(), err, "南西端経度が180超過の場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "swLng")
}

// TestNewBoundingBox_ValidationError_NeLngOutOfRange は北東端経度が範囲外の場合にエラーを返すことをテストする
func (s *BoundingBoxTestSuite) TestNewBoundingBox_ValidationError_NeLngOutOfRange() {
	// 経度が-180未満
	_, err := NewBoundingBox(35.6762, 139.6503, 35.7295, -181)
	require.Error(s.T(), err, "北東端経度が-180未満の場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "neLng")

	// 経度が180超過
	_, err = NewBoundingBox(35.6762, 139.6503, 35.7295, 181)
	require.Error(s.T(), err, "北東端経度が180超過の場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "neLng")
}

// TestNewBoundingBox_ValidationError_SwLatGreaterThanNeLat は南西端緯度が北東端緯度より大きい場合にエラーを返すことをテストする
func (s *BoundingBoxTestSuite) TestNewBoundingBox_ValidationError_SwLatGreaterThanNeLat() {
	_, err := NewBoundingBox(36.0, 139.6503, 35.0, 139.7673)
	require.Error(s.T(), err, "南西端緯度が北東端緯度より大きい場合はエラーが必要")
	require.Contains(s.T(), err.Error(), "南西端の緯度")
}

// TestNewBoundingBox_BoundaryValue_ExactLatitudeBoundaries は緯度の境界値(-90, 90)で正常に作成されることをテストする
func (s *BoundingBoxTestSuite) TestNewBoundingBox_BoundaryValue_ExactLatitudeBoundaries() {
	bbox, err := NewBoundingBox(-90, 0, 90, 0)
	require.NoError(s.T(), err, "境界値(-90, 90)でエラーが発生してはならない")
	require.NotNil(s.T(), bbox)
}

// TestNewBoundingBox_BoundaryValue_ExactLongitudeBoundaries は経度の境界値(-180, 180)で正常に作成されることをテストする
func (s *BoundingBoxTestSuite) TestNewBoundingBox_BoundaryValue_ExactLongitudeBoundaries() {
	bbox, err := NewBoundingBox(0, -180, 0, 180)
	require.NoError(s.T(), err, "境界値(-180, 180)でエラーが発生してはならない")
	require.NotNil(s.T(), bbox)
}

// TestDiagonalDistanceKm_Success_TokyoArea は東京周辺の範囲で対角線距離が正しく計算されることをテストする
func (s *BoundingBoxTestSuite) TestDiagonalDistanceKm_Success_TokyoArea() {
	// 東京周辺の範囲(約13km x 6km程度)
	bbox, err := NewBoundingBox(35.6762, 139.6503, 35.7295, 139.7673)
	require.NoError(s.T(), err)

	distance := bbox.DiagonalDistanceKm()
	// 対角線距離は約12-15km程度になるはず
	require.Greater(s.T(), distance, 10.0, "対角線距離は10kmより大きいはず")
	require.Less(s.T(), distance, 20.0, "対角線距離は20kmより小さいはず")
}

// TestDiagonalDistanceKm_Success_SmallArea は小さな範囲で対角線距離が正しく計算されることをテストする
func (s *BoundingBoxTestSuite) TestDiagonalDistanceKm_Success_SmallArea() {
	// 小さな範囲(約1km四方)
	bbox, err := NewBoundingBox(35.6800, 139.7000, 35.6900, 139.7100)
	require.NoError(s.T(), err)

	distance := bbox.DiagonalDistanceKm()
	// 対角線距離は約1.4km程度になるはず
	require.Greater(s.T(), distance, 1.0, "対角線距離は1kmより大きいはず")
	require.Less(s.T(), distance, 3.0, "対角線距離は3kmより小さいはず")
}

// TestDiagonalDistanceKm_Success_LargeArea は大きな範囲で対角線距離が正しく計算されることをテストする
func (s *BoundingBoxTestSuite) TestDiagonalDistanceKm_Success_LargeArea() {
	// 日本全国規模(約1000km四方)
	bbox, err := NewBoundingBox(31.0, 130.0, 45.0, 145.0)
	require.NoError(s.T(), err)

	distance := bbox.DiagonalDistanceKm()
	// 対角線距離は1000km以上になるはず
	require.Greater(s.T(), distance, 1000.0, "対角線距離は1000kmより大きいはず")
}

// TestOptimalH3Resolution_Success_VeryLargeArea は対角線300km超の範囲で解像度3が返されることをテストする
func (s *BoundingBoxTestSuite) TestOptimalH3Resolution_Success_VeryLargeArea() {
	// 対角線 > 300km
	bbox, err := NewBoundingBox(33.0, 130.0, 40.0, 140.0)
	require.NoError(s.T(), err)

	resolution := bbox.OptimalH3Resolution()
	require.Equal(s.T(), 3, resolution, "対角線300km超の場合は解像度3")
}

// TestOptimalH3Resolution_Success_LargeArea は対角線30-300kmの範囲で解像度5が返されることをテストする
func (s *BoundingBoxTestSuite) TestOptimalH3Resolution_Success_LargeArea() {
	// 対角線 30-300km(約100km程度)
	bbox, err := NewBoundingBox(35.0, 139.0, 36.0, 140.0)
	require.NoError(s.T(), err)

	resolution := bbox.OptimalH3Resolution()
	require.Equal(s.T(), 5, resolution, "対角線30-300kmの場合は解像度5")
}

// TestOptimalH3Resolution_Success_MediumArea は対角線3-30kmの範囲で解像度7が返されることをテストする
func (s *BoundingBoxTestSuite) TestOptimalH3Resolution_Success_MediumArea() {
	// 対角線 3-30km(約10km程度)
	bbox, err := NewBoundingBox(35.6500, 139.6500, 35.7500, 139.7500)
	require.NoError(s.T(), err)

	resolution := bbox.OptimalH3Resolution()
	require.Equal(s.T(), 7, resolution, "対角線3-30kmの場合は解像度7")
}

// TestOptimalH3Resolution_Success_SmallArea は対角線3km未満の範囲で解像度9が返されることをテストする
func (s *BoundingBoxTestSuite) TestOptimalH3Resolution_Success_SmallArea() {
	// 対角線 < 3km(約1km程度)
	bbox, err := NewBoundingBox(35.6800, 139.7000, 35.6900, 139.7100)
	require.NoError(s.T(), err)

	resolution := bbox.OptimalH3Resolution()
	require.Equal(s.T(), 9, resolution, "対角線3km未満の場合は解像度9")
}
