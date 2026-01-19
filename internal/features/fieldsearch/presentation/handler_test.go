package presentation

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

	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/application/usecase"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
)

// MockSearchFieldsUseCase はSearchFieldsUseCaseのモック
type MockSearchFieldsUseCase struct {
	mock.Mock
}

func (m *MockSearchFieldsUseCase) Execute(ctx context.Context, input usecase.SearchFieldsInput) (*usecase.SearchFieldsOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.SearchFieldsOutput), args.Error(1)
}

// FieldSearchHandlerTestSuite はFieldSearchHandlerのテストスイート
type FieldSearchHandlerTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (s *FieldSearchHandlerTestSuite) SetupTest() {
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestFieldSearchHandlerSuite(t *testing.T) {
	suite.Run(t, new(FieldSearchHandlerTestSuite))
}

// TestNewFieldSearchHandler_Success はNewFieldSearchHandlerが正しくハンドラーを生成することをテスト
func (s *FieldSearchHandlerTestSuite) TestNewFieldSearchHandler_Success() {
	uc := &usecase.SearchFieldsUseCase{}
	handler := NewFieldSearchHandler(uc, s.logger)
	require.NotNil(s.T(), handler, "ハンドラーがnilです")
	require.NotNil(s.T(), handler.searchFieldsUC, "searchFieldsUCがnilです")
	require.NotNil(s.T(), handler.logger, "loggerがnilです")
}

// TestFieldSearchHandler_toOpenAPISearchedField はUseCaseの出力がOpenAPI型に正しく変換されることをテスト
func (s *FieldSearchHandlerTestSuite) TestFieldSearchHandler_toOpenAPISearchedField() {
	handler := &FieldSearchHandler{
		searchFieldsUC: nil,
		logger:         s.logger,
	}

	fieldID := uuid.New()
	soilTypeID := uuid.New()
	areaSqm := 1500.5

	usecaseOutput := usecase.SearchedFieldOutput{
		ID:         fieldID,
		Name:       "田んぼA",
		CityCode:   "13101",
		SoilTypeID: &soilTypeID,
		AreaSqm:    &areaSqm,
		Geometry: []usecase.Coordinate{
			{Lat: 35.0, Lng: 139.0},
			{Lat: 35.0, Lng: 139.1},
			{Lat: 35.1, Lng: 139.1},
			{Lat: 35.1, Lng: 139.0},
			{Lat: 35.0, Lng: 139.0},
		},
		Centroid: usecase.Coordinate{Lat: 35.05, Lng: 139.05},
	}

	result := handler.toOpenAPISearchedField(usecaseOutput)

	require.Equal(s.T(), fieldID, result.Id, "IDが一致しません")
	require.Equal(s.T(), "田んぼA", result.Name, "圃場名が一致しません")
	require.Equal(s.T(), "13101", result.CityCode, "市区町村コードが一致しません")
	require.NotNil(s.T(), result.SoilTypeId, "SoilTypeIdがnilです")
	require.Equal(s.T(), soilTypeID, *result.SoilTypeId, "SoilTypeIdが一致しません")
	require.NotNil(s.T(), result.AreaSqm, "AreaSqmがnilです")
	require.Equal(s.T(), areaSqm, *result.AreaSqm, "AreaSqmが一致しません")
	require.Len(s.T(), result.Geometry, 5, "Geometry頂点数が一致しません")
	require.Equal(s.T(), 35.0, result.Geometry[0].Lat, "Geometry[0].Latが一致しません")
	require.Equal(s.T(), 139.0, result.Geometry[0].Lng, "Geometry[0].Lngが一致しません")
	require.Equal(s.T(), 35.05, result.Centroid.Lat, "Centroid.Latが一致しません")
	require.Equal(s.T(), 139.05, result.Centroid.Lng, "Centroid.Lngが一致しません")
}

// TestFieldSearchHandler_toOpenAPISearchedField_NullableFields はnullableフィールドがnilの場合も変換が正しく動作することをテスト
func (s *FieldSearchHandlerTestSuite) TestFieldSearchHandler_toOpenAPISearchedField_NullableFields() {
	handler := &FieldSearchHandler{
		searchFieldsUC: nil,
		logger:         s.logger,
	}

	fieldID := uuid.New()

	usecaseOutput := usecase.SearchedFieldOutput{
		ID:         fieldID,
		Name:       "田んぼB",
		CityCode:   "13102",
		SoilTypeID: nil,
		AreaSqm:    nil,
		Geometry:   nil,
		Centroid:   usecase.Coordinate{},
	}

	result := handler.toOpenAPISearchedField(usecaseOutput)

	require.Equal(s.T(), fieldID, result.Id, "IDが一致しません")
	require.Nil(s.T(), result.SoilTypeId, "SoilTypeIdがnilではありません")
	require.Nil(s.T(), result.AreaSqm, "AreaSqmがnilではありません")
	require.Empty(s.T(), result.Geometry, "Geometryが空ではありません")
	require.Equal(s.T(), 0.0, result.Centroid.Lat, "Centroid.Latがゼロではありません")
	require.Equal(s.T(), 0.0, result.Centroid.Lng, "Centroid.Lngがゼロではありません")
}

// TestIsValidationError_True はバリデーションエラーの場合trueを返すことをテスト
func (s *FieldSearchHandlerTestSuite) TestIsValidationError_True() {
	testCases := []struct {
		name string
		err  error
	}{
		{"南西端緯度エラー", errors.New("南西端の緯度は-90から90の範囲でなければなりません")},
		{"swLatキーワード", errors.New("swLatは-90から90の範囲でなければなりません")},
		{"北東端エラー", errors.New("北東端の経度が不正です")},
		{"neLngキーワード", errors.New("neLngは-180から180の範囲でなければなりません")},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := isValidationError(tc.err)
			require.True(s.T(), result, "バリデーションエラーと判定されるべき: %s", tc.err.Error())
		})
	}
}

// TestIsValidationError_False はビジネスロジックエラーの場合falseを返すことをテスト
func (s *FieldSearchHandlerTestSuite) TestIsValidationError_False() {
	testCases := []struct {
		name string
		err  error
	}{
		{"データベースエラー", errors.New("データベース接続に失敗しました")},
		{"H3計算エラー", errors.New("H3セルの計算に失敗しました")},
		{"一般的なエラー", errors.New("予期しないエラーが発生しました")},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := isValidationError(tc.err)
			require.False(s.T(), result, "バリデーションエラーではないと判定されるべき: %s", tc.err.Error())
		})
	}
}

// testableSearchFieldsHandler はテスト用のSearchFieldsハンドラー
type testableSearchFieldsHandler struct {
	executeFunc func(ctx context.Context, input usecase.SearchFieldsInput) (*usecase.SearchFieldsOutput, error)
	logger      *slog.Logger
}

func (h *testableSearchFieldsHandler) SearchFields(ctx context.Context, request openapi.SearchFieldsRequestObject) (openapi.SearchFieldsResponseObject, error) {
	params := request.Params

	output, err := h.executeFunc(ctx, usecase.SearchFieldsInput{
		SwLat: params.SwLat,
		SwLng: params.SwLng,
		NeLat: params.NeLat,
		NeLng: params.NeLng,
	})
	if err != nil {
		if isValidationError(err) {
			return openapi.SearchFields400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: err.Error(),
					}},
				},
			}, nil
		}

		return openapi.SearchFields500JSONResponse{
			InternalServerErrorJSONResponse: openapi.InternalServerErrorJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "internal_error",
					Message: "圃場検索に失敗しました",
				}},
			},
		}, nil
	}

	fields := make([]openapi.SearchedField, 0, len(output.Fields))
	for _, field := range output.Fields {
		geometry := make([]openapi.SearchedCoordinate, 0, len(field.Geometry))
		for _, coord := range field.Geometry {
			geometry = append(geometry, openapi.SearchedCoordinate{Lat: coord.Lat, Lng: coord.Lng})
		}
		fields = append(fields, openapi.SearchedField{
			Id:         field.ID,
			Name:       field.Name,
			CityCode:   field.CityCode,
			SoilTypeId: field.SoilTypeID,
			AreaSqm:    field.AreaSqm,
			Geometry:   geometry,
			Centroid:   openapi.SearchedCoordinate{Lat: field.Centroid.Lat, Lng: field.Centroid.Lng},
		})
	}

	return openapi.SearchFields200JSONResponse{
		Data:   &openapi.FieldSearchData{Fields: fields},
		Errors: nil,
	}, nil
}

// FieldSearchHandlerSearchFieldsTestSuite はSearchFieldsのテストスイート
type FieldSearchHandlerSearchFieldsTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (s *FieldSearchHandlerSearchFieldsTestSuite) SetupTest() {
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestFieldSearchHandlerSearchFieldsSuite(t *testing.T) {
	suite.Run(t, new(FieldSearchHandlerSearchFieldsTestSuite))
}

// TestFieldSearchHandler_SearchFields_Success は正常に圃場検索を実行できることをテスト
func (s *FieldSearchHandlerSearchFieldsTestSuite) TestFieldSearchHandler_SearchFields_Success() {
	ctx := context.Background()
	fieldID := uuid.New()

	handler := &testableSearchFieldsHandler{
		executeFunc: func(_ context.Context, _ usecase.SearchFieldsInput) (*usecase.SearchFieldsOutput, error) {
			return &usecase.SearchFieldsOutput{
				Fields: []usecase.SearchedFieldOutput{
					{
						ID:       fieldID,
						Name:     "田んぼA",
						CityCode: "13101",
						Geometry: []usecase.Coordinate{{Lat: 35.0, Lng: 139.0}},
						Centroid: usecase.Coordinate{Lat: 35.0, Lng: 139.0},
					},
				},
			}, nil
		},
		logger: s.logger,
	}

	request := openapi.SearchFieldsRequestObject{
		Params: openapi.SearchFieldsParams{
			SwLat: 35.6800,
			SwLng: 139.7000,
			NeLat: 35.6900,
			NeLng: 139.7100,
		},
	}

	response, err := handler.SearchFields(ctx, request)
	require.NoError(s.T(), err, "SearchFields実行時にエラーが発生")

	resp200, ok := response.(openapi.SearchFields200JSONResponse)
	require.True(s.T(), ok, "200レスポンスを期待")
	require.NotNil(s.T(), resp200.Data, "Dataがnilです")
	require.Len(s.T(), resp200.Data.Fields, 1, "フィールド数が期待値と異なります")
	require.Equal(s.T(), fieldID, resp200.Data.Fields[0].Id, "フィールドIDが一致しません")
}

// TestFieldSearchHandler_SearchFields_Success_EmptyResult は結果が0件でも正常に動作することをテスト
func (s *FieldSearchHandlerSearchFieldsTestSuite) TestFieldSearchHandler_SearchFields_Success_EmptyResult() {
	ctx := context.Background()

	handler := &testableSearchFieldsHandler{
		executeFunc: func(_ context.Context, _ usecase.SearchFieldsInput) (*usecase.SearchFieldsOutput, error) {
			return &usecase.SearchFieldsOutput{
				Fields: []usecase.SearchedFieldOutput{},
			}, nil
		},
		logger: s.logger,
	}

	request := openapi.SearchFieldsRequestObject{
		Params: openapi.SearchFieldsParams{
			SwLat: 35.6800,
			SwLng: 139.7000,
			NeLat: 35.6900,
			NeLng: 139.7100,
		},
	}

	response, err := handler.SearchFields(ctx, request)
	require.NoError(s.T(), err, "SearchFields実行時にエラーが発生")

	resp200, ok := response.(openapi.SearchFields200JSONResponse)
	require.True(s.T(), ok, "200レスポンスを期待")
	require.NotNil(s.T(), resp200.Data, "Dataがnilです")
	require.Empty(s.T(), resp200.Data.Fields, "フィールドリストが空ではありません")
}

// TestFieldSearchHandler_SearchFields_ValidationError_InvalidCoordinates は無効な座標で400エラーになることをテスト
func (s *FieldSearchHandlerSearchFieldsTestSuite) TestFieldSearchHandler_SearchFields_ValidationError_InvalidCoordinates() {
	ctx := context.Background()

	handler := &testableSearchFieldsHandler{
		executeFunc: func(_ context.Context, _ usecase.SearchFieldsInput) (*usecase.SearchFieldsOutput, error) {
			return nil, errors.New("南西端の緯度は北東端の緯度より小さくなければなりません")
		},
		logger: s.logger,
	}

	request := openapi.SearchFieldsRequestObject{
		Params: openapi.SearchFieldsParams{
			SwLat: 36.0, // 南西端が北東端より北
			SwLng: 139.0,
			NeLat: 35.0,
			NeLng: 140.0,
		},
	}

	response, err := handler.SearchFields(ctx, request)
	require.NoError(s.T(), err, "SearchFields実行時に予期しないエラーが発生")

	resp400, ok := response.(openapi.SearchFields400JSONResponse)
	require.True(s.T(), ok, "400レスポンスを期待")
	require.NotNil(s.T(), resp400.Errors, "Errorsがnilです")
	require.Len(s.T(), *resp400.Errors, 1, "エラー数が期待値と異なります")
}

// TestFieldSearchHandler_SearchFields_UseCaseError はユースケースエラー時に500エラーになることをテスト
func (s *FieldSearchHandlerSearchFieldsTestSuite) TestFieldSearchHandler_SearchFields_UseCaseError() {
	ctx := context.Background()

	handler := &testableSearchFieldsHandler{
		executeFunc: func(_ context.Context, _ usecase.SearchFieldsInput) (*usecase.SearchFieldsOutput, error) {
			return nil, errors.New("データベースエラー")
		},
		logger: s.logger,
	}

	request := openapi.SearchFieldsRequestObject{
		Params: openapi.SearchFieldsParams{
			SwLat: 35.6800,
			SwLng: 139.7000,
			NeLat: 35.6900,
			NeLng: 139.7100,
		},
	}

	response, err := handler.SearchFields(ctx, request)
	require.NoError(s.T(), err, "SearchFields実行時に予期しないエラーが発生")

	resp500, ok := response.(openapi.SearchFields500JSONResponse)
	require.True(s.T(), ok, "500レスポンスを期待")
	require.NotNil(s.T(), resp500.Errors, "Errorsがnilです")
}
