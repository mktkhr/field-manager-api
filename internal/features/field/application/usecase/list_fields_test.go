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

	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
)

// MockFieldQuery はFieldQueryのモック
type MockFieldQuery struct {
	mock.Mock
}

func (m *MockFieldQuery) List(ctx context.Context, limit, offset int32) ([]*entity.Field, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Field), args.Error(1)
}

func (m *MockFieldQuery) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// ListFieldsUseCaseTestSuite はListFieldsUseCaseのテストスイート
type ListFieldsUseCaseTestSuite struct {
	suite.Suite
	mockQuery *MockFieldQuery
	useCase   *ListFieldsUseCase
	logger    *slog.Logger
}

func (s *ListFieldsUseCaseTestSuite) SetupTest() {
	s.mockQuery = new(MockFieldQuery)
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	s.useCase = NewListFieldsUseCase(s.mockQuery, s.logger)
}

func TestListFieldsUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ListFieldsUseCaseTestSuite))
}

// TestNewListFieldsUseCase_Success はNewListFieldsUseCaseが正しくUseCaseを生成することをテスト
func (s *ListFieldsUseCaseTestSuite) TestNewListFieldsUseCase_Success() {
	uc := NewListFieldsUseCase(s.mockQuery, s.logger)
	require.NotNil(s.T(), uc, "UseCaseがnilです")
	require.NotNil(s.T(), uc.fieldQuery, "fieldQueryがnilです")
	require.NotNil(s.T(), uc.logger, "loggerがnilです")
}

// TestListFieldsUseCase_Execute_Success は正常に圃場一覧を取得できることをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_Success() {
	ctx := context.Background()
	fieldID := uuid.New()

	// テスト用のPolygon作成
	polygon := geom.NewPolygon(geom.XY)
	coords := [][]geom.Coord{{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}}
	_, err := polygon.SetCoords(coords)
	require.NoError(s.T(), err, "ポリゴン座標設定に失敗")

	// テスト用のPoint作成
	centroid := geom.NewPoint(geom.XY)
	_, err = centroid.SetCoords(geom.Coord{139.05, 35.05})
	require.NoError(s.T(), err, "重心座標設定に失敗")

	area := 10000.0
	fields := []*entity.Field{
		{
			ID:       fieldID,
			Name:     "テスト圃場",
			CityCode: "12345",
			AreaSqm:  &area,
			Geometry: polygon,
			Centroid: centroid,
		},
	}

	s.mockQuery.On("Count", ctx).Return(int64(1), nil)
	s.mockQuery.On("List", ctx, int32(20), int32(0)).Return(fields, nil)

	input := ListFieldsInput{Page: 1, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 1, "圃場数が期待値と異なります")
	require.Equal(s.T(), fieldID, output.Fields[0].ID, "圃場IDが一致しません")
	require.Equal(s.T(), "テスト圃場", output.Fields[0].Name, "圃場名が一致しません")
	require.Equal(s.T(), "12345", output.Fields[0].CityCode, "市区町村コードが一致しません")
	require.Equal(s.T(), int64(1), output.Pagination.Total, "総件数が一致しません")
	require.Equal(s.T(), 1, output.Pagination.Page, "ページ番号が一致しません")
	require.Equal(s.T(), 20, output.Pagination.PageSize, "ページサイズが一致しません")
	require.Equal(s.T(), 1, output.Pagination.TotalPages, "総ページ数が一致しません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_Success_EmptyResult は結果が0件でも正常に動作することをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_Success_EmptyResult() {
	ctx := context.Background()

	s.mockQuery.On("Count", ctx).Return(int64(0), nil)
	s.mockQuery.On("List", ctx, int32(20), int32(0)).Return([]*entity.Field{}, nil)

	input := ListFieldsInput{Page: 1, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Empty(s.T(), output.Fields, "圃場リストが空ではありません")
	require.Equal(s.T(), int64(0), output.Pagination.Total, "総件数が0ではありません")
	require.Equal(s.T(), 0, output.Pagination.TotalPages, "総ページ数が0ではありません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_Success_WithGeometry はGeometry変換が正しく行われることをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_Success_WithGeometry() {
	ctx := context.Background()
	fieldID := uuid.New()

	// テスト用のPolygon作成(経度139.0-139.1, 緯度35.0-35.1の四角形)
	polygon := geom.NewPolygon(geom.XY)
	coords := [][]geom.Coord{{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}}
	_, err := polygon.SetCoords(coords)
	require.NoError(s.T(), err, "ポリゴン座標設定に失敗")

	// テスト用のPoint作成
	centroid := geom.NewPoint(geom.XY)
	_, err = centroid.SetCoords(geom.Coord{139.05, 35.05})
	require.NoError(s.T(), err, "重心座標設定に失敗")

	fields := []*entity.Field{
		{
			ID:       fieldID,
			Name:     "テスト圃場",
			CityCode: "12345",
			Geometry: polygon,
			Centroid: centroid,
		},
	}

	s.mockQuery.On("Count", ctx).Return(int64(1), nil)
	s.mockQuery.On("List", ctx, int32(20), int32(0)).Return(fields, nil)

	input := ListFieldsInput{Page: 1, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 1, "圃場数が期待値と異なります")

	// Geometry変換の検証(座標順序: lat=Y, lng=X)
	require.Len(s.T(), output.Fields[0].Geometry, 5, "ポリゴン頂点数が期待値と異なります")
	require.Equal(s.T(), 35.0, output.Fields[0].Geometry[0].Lat, "1番目の頂点の緯度が一致しません")
	require.Equal(s.T(), 139.0, output.Fields[0].Geometry[0].Lng, "1番目の頂点の経度が一致しません")

	// Centroid変換の検証
	require.Equal(s.T(), 35.05, output.Fields[0].Centroid.Lat, "重心の緯度が一致しません")
	require.Equal(s.T(), 139.05, output.Fields[0].Centroid.Lng, "重心の経度が一致しません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_Success_WithoutGeometry はGeometryがnilの場合も正常に動作することをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_Success_WithoutGeometry() {
	ctx := context.Background()
	fieldID := uuid.New()

	fields := []*entity.Field{
		{
			ID:       fieldID,
			Name:     "テスト圃場",
			CityCode: "12345",
			Geometry: nil,
			Centroid: nil,
		},
	}

	s.mockQuery.On("Count", ctx).Return(int64(1), nil)
	s.mockQuery.On("List", ctx, int32(20), int32(0)).Return(fields, nil)

	input := ListFieldsInput{Page: 1, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 1, "圃場数が期待値と異なります")
	require.Nil(s.T(), output.Fields[0].Geometry, "Geometryがnilではありません")
	require.Equal(s.T(), Coordinate{}, output.Fields[0].Centroid, "Centroidがゼロ値ではありません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_CountError はCount失敗時にエラーを返すことをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_CountError() {
	ctx := context.Background()
	expectedErr := errors.New("データベースエラー")

	s.mockQuery.On("Count", ctx).Return(int64(0), expectedErr)

	input := ListFieldsInput{Page: 1, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.Error(s.T(), err, "エラーが返されませんでした")
	require.Nil(s.T(), output, "出力がnilではありません")
	require.Equal(s.T(), expectedErr, err, "エラーが期待値と一致しません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_ListError はList失敗時にエラーを返すことをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_ListError() {
	ctx := context.Background()
	expectedErr := errors.New("データベースエラー")

	s.mockQuery.On("Count", ctx).Return(int64(10), nil)
	s.mockQuery.On("List", ctx, int32(20), int32(0)).Return(nil, expectedErr)

	input := ListFieldsInput{Page: 1, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.Error(s.T(), err, "エラーが返されませんでした")
	require.Nil(s.T(), output, "出力がnilではありません")
	require.Equal(s.T(), expectedErr, err, "エラーが期待値と一致しません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_PaginationCalculation はページネーション計算が正しく行われることをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_PaginationCalculation() {
	ctx := context.Background()

	s.mockQuery.On("Count", ctx).Return(int64(45), nil)
	s.mockQuery.On("List", ctx, int32(20), int32(20)).Return([]*entity.Field{}, nil)

	// 2ページ目をリクエスト
	input := ListFieldsInput{Page: 2, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Equal(s.T(), int64(45), output.Pagination.Total, "総件数が一致しません")
	require.Equal(s.T(), 2, output.Pagination.Page, "ページ番号が一致しません")
	require.Equal(s.T(), 20, output.Pagination.PageSize, "ページサイズが一致しません")
	require.Equal(s.T(), 3, output.Pagination.TotalPages, "総ページ数が一致しません(45件/20件=3ページ)")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_TotalPagesCalculation_ExactDivision は総ページ数計算(割り切れる場合)をテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_TotalPagesCalculation_ExactDivision() {
	ctx := context.Background()

	s.mockQuery.On("Count", ctx).Return(int64(40), nil)
	s.mockQuery.On("List", ctx, int32(20), int32(0)).Return([]*entity.Field{}, nil)

	input := ListFieldsInput{Page: 1, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.Equal(s.T(), 2, output.Pagination.TotalPages, "総ページ数が一致しません(40件/20件=2ページ)")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_TotalPagesCalculation_WithRemainder は総ページ数計算(端数がある場合)をテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_TotalPagesCalculation_WithRemainder() {
	ctx := context.Background()

	s.mockQuery.On("Count", ctx).Return(int64(41), nil)
	s.mockQuery.On("List", ctx, int32(20), int32(0)).Return([]*entity.Field{}, nil)

	input := ListFieldsInput{Page: 1, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.Equal(s.T(), 3, output.Pagination.TotalPages, "総ページ数が一致しません(41件/20件=3ページ)")

	s.mockQuery.AssertExpectations(s.T())
}
