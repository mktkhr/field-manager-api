package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

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

func (m *MockFieldQuery) ListByCursor(ctx context.Context, cursor *entity.FieldCursor, limit int32) ([]*entity.Field, error) {
	args := m.Called(ctx, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Field), args.Error(1)
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

// TestListFieldsUseCase_Execute_Success_FirstPage は初回ページ取得が正常に動作することをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_Success_FirstPage() {
	ctx := context.Background()
	fieldID := uuid.New()
	createdAt := time.Now()

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
	// pageSize + 1 件を返す(次ページがあることを示す)
	fields := []*entity.Field{
		{
			ID:        fieldID,
			Name:      "テスト圃場",
			CityCode:  "12345",
			AreaSqm:   &area,
			Geometry:  polygon,
			Centroid:  centroid,
			CreatedAt: createdAt,
		},
	}

	// cursor=nil, limit=pageSize+1(21)で呼ばれる
	s.mockQuery.On("ListByCursor", ctx, (*entity.FieldCursor)(nil), int32(21)).Return(fields, nil)

	input := ListFieldsInput{Cursor: nil, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 1, "圃場数が期待値と異なります")
	require.Equal(s.T(), fieldID, output.Fields[0].ID, "圃場IDが一致しません")
	require.Equal(s.T(), "テスト圃場", output.Fields[0].Name, "圃場名が一致しません")
	require.Equal(s.T(), "12345", output.Fields[0].CityCode, "市区町村コードが一致しません")
	require.False(s.T(), output.Pagination.HasMore, "HasMoreがfalseではありません")
	require.Nil(s.T(), output.Pagination.NextCursor, "NextCursorがnilではありません")
	require.Equal(s.T(), 20, output.Pagination.PageSize, "PageSizeが一致しません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_Success_HasMore は次ページがある場合のテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_Success_HasMore() {
	ctx := context.Background()
	now := time.Now()

	// pageSize + 1 件(21件)を返す(次ページがあることを示す)
	fields := make([]*entity.Field, 21)
	for i := 0; i < 21; i++ {
		fields[i] = &entity.Field{
			ID:        uuid.New(),
			Name:      "テスト圃場",
			CityCode:  "12345",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	s.mockQuery.On("ListByCursor", ctx, (*entity.FieldCursor)(nil), int32(21)).Return(fields, nil)

	input := ListFieldsInput{Cursor: nil, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 20, "圃場数が期待値と異なります(21件中20件を返す)")
	require.True(s.T(), output.Pagination.HasMore, "HasMoreがtrueではありません")
	require.NotNil(s.T(), output.Pagination.NextCursor, "NextCursorがnilです")
	require.Equal(s.T(), 20, output.Pagination.PageSize, "PageSizeが一致しません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_Success_EmptyResult は結果が0件でも正常に動作することをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_Success_EmptyResult() {
	ctx := context.Background()

	s.mockQuery.On("ListByCursor", ctx, (*entity.FieldCursor)(nil), int32(21)).Return([]*entity.Field{}, nil)

	input := ListFieldsInput{Cursor: nil, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Empty(s.T(), output.Fields, "圃場リストが空ではありません")
	require.False(s.T(), output.Pagination.HasMore, "HasMoreがfalseではありません")
	require.Nil(s.T(), output.Pagination.NextCursor, "NextCursorがnilではありません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_Success_WithCursor はカーソル指定時の取得をテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_Success_WithCursor() {
	ctx := context.Background()
	cursorTime := time.Now().Add(-time.Hour)
	cursorID := uuid.New()
	cursor := entity.NewFieldCursor(cursorTime, cursorID)

	fields := []*entity.Field{
		{
			ID:        uuid.New(),
			Name:      "テスト圃場2",
			CityCode:  "12345",
			CreatedAt: cursorTime.Add(-time.Hour),
		},
	}

	s.mockQuery.On("ListByCursor", ctx, cursor, int32(21)).Return(fields, nil)

	input := ListFieldsInput{Cursor: cursor, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 1, "圃場数が期待値と異なります")
	require.False(s.T(), output.Pagination.HasMore, "HasMoreがfalseではありません")

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
			ID:        fieldID,
			Name:      "テスト圃場",
			CityCode:  "12345",
			Geometry:  polygon,
			Centroid:  centroid,
			CreatedAt: time.Now(),
		},
	}

	s.mockQuery.On("ListByCursor", ctx, (*entity.FieldCursor)(nil), int32(21)).Return(fields, nil)

	input := ListFieldsInput{Cursor: nil, PageSize: 20}
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
			ID:        fieldID,
			Name:      "テスト圃場",
			CityCode:  "12345",
			Geometry:  nil,
			Centroid:  nil,
			CreatedAt: time.Now(),
		},
	}

	s.mockQuery.On("ListByCursor", ctx, (*entity.FieldCursor)(nil), int32(21)).Return(fields, nil)

	input := ListFieldsInput{Cursor: nil, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.NoError(s.T(), err, "Execute実行時にエラーが発生")
	require.NotNil(s.T(), output, "出力がnilです")
	require.Len(s.T(), output.Fields, 1, "圃場数が期待値と異なります")
	require.Nil(s.T(), output.Fields[0].Geometry, "Geometryがnilではありません")
	require.Equal(s.T(), Coordinate{}, output.Fields[0].Centroid, "Centroidがゼロ値ではありません")

	s.mockQuery.AssertExpectations(s.T())
}

// TestListFieldsUseCase_Execute_ListError はList失敗時にエラーを返すことをテスト
func (s *ListFieldsUseCaseTestSuite) TestListFieldsUseCase_Execute_ListError() {
	ctx := context.Background()
	expectedErr := errors.New("データベースエラー")

	s.mockQuery.On("ListByCursor", ctx, (*entity.FieldCursor)(nil), int32(21)).Return(nil, expectedErr)

	input := ListFieldsInput{Cursor: nil, PageSize: 20}
	output, err := s.useCase.Execute(ctx, input)

	require.Error(s.T(), err, "エラーが返されませんでした")
	require.Nil(s.T(), output, "出力がnilではありません")
	require.Equal(s.T(), expectedErr, err, "エラーが期待値と一致しません")

	s.mockQuery.AssertExpectations(s.T())
}
