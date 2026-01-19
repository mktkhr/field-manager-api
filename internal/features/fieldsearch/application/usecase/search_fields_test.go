package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/twpayne/go-geom"

	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/domain/entity"
)

// MockFieldSearchQuery はFieldSearchQueryのモック
type MockFieldSearchQuery struct {
	mock.Mock
}

func (m *MockFieldSearchQuery) SearchByBBox(ctx context.Context, bbox *entity.BoundingBox, h3Cells []string, resolution int) ([]*entity.SearchedField, error) {
	args := m.Called(ctx, bbox, h3Cells, resolution)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.SearchedField), args.Error(1)
}

// MockH3CellCalculator はH3CellCalculatorのモック
type MockH3CellCalculator struct {
	mock.Mock
}

func (m *MockH3CellCalculator) CalculateCells(bbox *entity.BoundingBox, resolution int) ([]string, error) {
	args := m.Called(bbox, resolution)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// SearchFieldsUseCaseTestSuite はSearchFieldsUseCaseのテストスイート
type SearchFieldsUseCaseTestSuite struct {
	suite.Suite
	mockQuery      *MockFieldSearchQuery
	mockCalculator *MockH3CellCalculator
	useCase        *SearchFieldsUseCase
	logger         *slog.Logger
}

func (s *SearchFieldsUseCaseTestSuite) SetupTest() {
	s.mockQuery = new(MockFieldSearchQuery)
	s.mockCalculator = new(MockH3CellCalculator)
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	s.useCase = NewSearchFieldsUseCase(s.mockQuery, s.mockCalculator, s.logger)
}

func TestSearchFieldsUseCaseSuite(t *testing.T) {
	suite.Run(t, new(SearchFieldsUseCaseTestSuite))
}

// TestNewSearchFieldsUseCase_Success はNewSearchFieldsUseCaseが正しくUseCaseを生成することをテスト
func (s *SearchFieldsUseCaseTestSuite) TestNewSearchFieldsUseCase_Success() {
	uc := NewSearchFieldsUseCase(s.mockQuery, s.mockCalculator, s.logger)
	require.NotNil(s.T(), uc, "UseCaseがnilです")
	require.NotNil(s.T(), uc.searchQuery, "searchQueryがnilです")
	require.NotNil(s.T(), uc.h3Calculator, "h3Calculatorがnilです")
	require.NotNil(s.T(), uc.logger, "loggerがnilです")
}

// TestSearchFieldsUseCase_Execute_Success_WithResults は検索結果がある場合に正常に動作することをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_Success_WithResults() {
	ctx := context.Background()
	fieldID := uuid.New()
	soilTypeID := uuid.New()
	areaSqm := 1500.5

	// テスト用のPolygon作成
	polygon := geom.NewPolygon(geom.XY)
	coords := [][]geom.Coord{{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}}
	_, err := polygon.SetCoords(coords)
	require.NoError(s.T(), err, "ポリゴン座標設定に失敗")

	// テスト用のPoint作成
	centroid := geom.NewPoint(geom.XY)
	_, err = centroid.SetCoords(geom.Coord{139.05, 35.05})
	require.NoError(s.T(), err, "重心座標設定に失敗")

	h3Cells := []string{"872f1a6ddffffff", "872f1a6ddffffff"}

	fields := []*entity.SearchedField{
		{
			ID:         fieldID,
			Name:       "田んぼA",
			CityCode:   "13101",
			SoilTypeID: &soilTypeID,
			AreaSqm:    &areaSqm,
			Geometry:   polygon,
			Centroid:   centroid,
		},
	}

	// 東京周辺の小さな範囲(resolution=9が選択される)
	s.mockCalculator.On("CalculateCells", mock.Anything, 9).Return(h3Cells, nil)
	s.mockQuery.On("SearchByBBox", ctx, mock.Anything, h3Cells, 9).Return(fields, nil)

	input := SearchFieldsInput{
		SwLat: 35.6800,
		SwLng: 139.7000,
		NeLat: 35.6900,
		NeLng: 139.7100,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 1, "圃場数が期待値と異なります")
	require.Equal(s.T(), fieldID, output.Fields[0].ID, "圃場IDが一致しません")
	require.Equal(s.T(), "田んぼA", output.Fields[0].Name, "圃場名が一致しません")
	require.Equal(s.T(), "13101", output.Fields[0].CityCode, "市区町村コードが一致しません")
	require.NotNil(s.T(), output.Fields[0].SoilTypeID, "SoilTypeIDがnilです")
	require.Equal(s.T(), soilTypeID, *output.Fields[0].SoilTypeID, "SoilTypeIDが一致しません")
	require.NotNil(s.T(), output.Fields[0].AreaSqm, "AreaSqmがnilです")
	require.Equal(s.T(), areaSqm, *output.Fields[0].AreaSqm, "AreaSqmが一致しません")

	// Geometry変換の検証
	require.Len(s.T(), output.Fields[0].Geometry, 5, "ポリゴン頂点数が期待値と異なります")
	require.Equal(s.T(), 35.0, output.Fields[0].Geometry[0].Lat, "1番目の頂点の緯度が一致しません")
	require.Equal(s.T(), 139.0, output.Fields[0].Geometry[0].Lng, "1番目の頂点の経度が一致しません")

	// Centroid変換の検証
	require.Equal(s.T(), 35.05, output.Fields[0].Centroid.Lat, "重心の緯度が一致しません")
	require.Equal(s.T(), 139.05, output.Fields[0].Centroid.Lng, "重心の経度が一致しません")

	s.mockCalculator.AssertExpectations(s.T())
	s.mockQuery.AssertExpectations(s.T())
}

// TestSearchFieldsUseCase_Execute_Success_EmptyResults は検索結果が0件の場合も正常に動作することをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_Success_EmptyResults() {
	ctx := context.Background()
	h3Cells := []string{"872f1a6ddffffff"}

	s.mockCalculator.On("CalculateCells", mock.Anything, 9).Return(h3Cells, nil)
	s.mockQuery.On("SearchByBBox", ctx, mock.Anything, h3Cells, 9).Return([]*entity.SearchedField{}, nil)

	input := SearchFieldsInput{
		SwLat: 35.6800,
		SwLng: 139.7000,
		NeLat: 35.6900,
		NeLng: 139.7100,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Empty(s.T(), output.Fields, "圃場リストが空ではありません")

	s.mockCalculator.AssertExpectations(s.T())
	s.mockQuery.AssertExpectations(s.T())
}

// TestSearchFieldsUseCase_Execute_Success_LargeArea は大きな検索範囲でresolution=5が選択されることをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_Success_LargeArea() {
	ctx := context.Background()
	h3Cells := []string{"852f1a6ddffffff", "852f1a6ddffffff"}

	// 対角線約100km(resolution=5が選択される)
	s.mockCalculator.On("CalculateCells", mock.Anything, 5).Return(h3Cells, nil)
	s.mockQuery.On("SearchByBBox", ctx, mock.Anything, h3Cells, 5).Return([]*entity.SearchedField{}, nil)

	input := SearchFieldsInput{
		SwLat: 35.0,
		SwLng: 139.0,
		NeLat: 36.0,
		NeLng: 140.0,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")

	s.mockCalculator.AssertExpectations(s.T())
	s.mockQuery.AssertExpectations(s.T())
}

// TestSearchFieldsUseCase_Execute_Success_NullableFields はnullableフィールドがnilの場合も正常に動作することをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_Success_NullableFields() {
	ctx := context.Background()
	fieldID := uuid.New()
	h3Cells := []string{"872f1a6ddffffff"}

	fields := []*entity.SearchedField{
		{
			ID:         fieldID,
			Name:       "田んぼB",
			CityCode:   "13102",
			SoilTypeID: nil,
			AreaSqm:    nil,
			Geometry:   nil,
			Centroid:   nil,
		},
	}

	s.mockCalculator.On("CalculateCells", mock.Anything, 9).Return(h3Cells, nil)
	s.mockQuery.On("SearchByBBox", ctx, mock.Anything, h3Cells, 9).Return(fields, nil)

	input := SearchFieldsInput{
		SwLat: 35.6800,
		SwLng: 139.7000,
		NeLat: 35.6900,
		NeLng: 139.7100,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 1, "圃場数が期待値と異なります")
	require.Equal(s.T(), fieldID, output.Fields[0].ID, "圃場IDが一致しません")
	require.Nil(s.T(), output.Fields[0].SoilTypeID, "SoilTypeIDがnilではありません")
	require.Nil(s.T(), output.Fields[0].AreaSqm, "AreaSqmがnilではありません")
	require.Nil(s.T(), output.Fields[0].Geometry, "Geometryがnilではありません")
	require.Equal(s.T(), Coordinate{}, output.Fields[0].Centroid, "Centroidがゼロ値ではありません")

	s.mockCalculator.AssertExpectations(s.T())
	s.mockQuery.AssertExpectations(s.T())
}

// TestSearchFieldsUseCase_Execute_ValidationError_InvalidBBox は無効なBBox座標でエラーを返すことをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_ValidationError_InvalidBBox() {
	ctx := context.Background()

	// 南西端緯度が北東端緯度より大きい
	input := SearchFieldsInput{
		SwLat: 36.0,
		SwLng: 139.0,
		NeLat: 35.0,
		NeLng: 140.0,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.Error(s.T(), err, "BBox検証エラーが発生すべき")
	require.Nil(s.T(), output, "出力がnilではありません")
	require.Contains(s.T(), err.Error(), "南西端の緯度", "エラーメッセージが期待と異なります")
}

// TestSearchFieldsUseCase_Execute_ValidationError_LatitudeOutOfRange は緯度範囲外でエラーを返すことをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_ValidationError_LatitudeOutOfRange() {
	ctx := context.Background()

	input := SearchFieldsInput{
		SwLat: -91.0,
		SwLng: 139.0,
		NeLat: 35.0,
		NeLng: 140.0,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.Error(s.T(), err, "緯度範囲外エラーが発生すべき")
	require.Nil(s.T(), output, "出力がnilではありません")
}

// TestSearchFieldsUseCase_Execute_ValidationError_LongitudeOutOfRange は経度範囲外でエラーを返すことをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_ValidationError_LongitudeOutOfRange() {
	ctx := context.Background()

	input := SearchFieldsInput{
		SwLat: 35.0,
		SwLng: 181.0,
		NeLat: 36.0,
		NeLng: 140.0,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.Error(s.T(), err, "経度範囲外エラーが発生すべき")
	require.Nil(s.T(), output, "出力がnilではありません")
}

// TestSearchFieldsUseCase_Execute_BusinessLogicError_H3CalculationFailed はH3セル計算失敗時にエラーを返すことをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_BusinessLogicError_H3CalculationFailed() {
	ctx := context.Background()
	expectedErr := errors.New("H3セル計算エラー")

	s.mockCalculator.On("CalculateCells", mock.Anything, 9).Return(nil, expectedErr)

	input := SearchFieldsInput{
		SwLat: 35.6800,
		SwLng: 139.7000,
		NeLat: 35.6900,
		NeLng: 139.7100,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.Error(s.T(), err, "エラーが返されませんでした")
	require.Nil(s.T(), output, "出力がnilではありません")
	require.Equal(s.T(), expectedErr, err, "エラーが期待値と一致しません")

	s.mockCalculator.AssertExpectations(s.T())
}

// TestSearchFieldsUseCase_Execute_BusinessLogicError_SearchFailed は検索失敗時にエラーを返すことをテスト
func (s *SearchFieldsUseCaseTestSuite) TestSearchFieldsUseCase_Execute_BusinessLogicError_SearchFailed() {
	ctx := context.Background()
	expectedErr := errors.New("データベースエラー")
	h3Cells := []string{"872f1a6ddffffff"}

	s.mockCalculator.On("CalculateCells", mock.Anything, 9).Return(h3Cells, nil)
	s.mockQuery.On("SearchByBBox", ctx, mock.Anything, h3Cells, 9).Return(nil, expectedErr)

	input := SearchFieldsInput{
		SwLat: 35.6800,
		SwLng: 139.7000,
		NeLat: 35.6900,
		NeLng: 139.7100,
	}
	output, err := s.useCase.Execute(ctx, input)

	require.Error(s.T(), err, "エラーが返されませんでした")
	require.Nil(s.T(), output, "出力がnilではありません")
	require.Equal(s.T(), expectedErr, err, "エラーが期待値と一致しません")

	s.mockCalculator.AssertExpectations(s.T())
	s.mockQuery.AssertExpectations(s.T())
}
