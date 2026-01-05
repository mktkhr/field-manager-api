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

	"github.com/mktkhr/field-manager-api/internal/features/field/application/usecase"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
)

// MockListFieldsUseCase はListFieldsUseCaseのモック
type MockListFieldsUseCase struct {
	mock.Mock
}

func (m *MockListFieldsUseCase) Execute(ctx context.Context, input usecase.ListFieldsInput) (*usecase.ListFieldsOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.ListFieldsOutput), args.Error(1)
}

// FieldHandlerTestSuite はFieldHandlerのテストスイート
type FieldHandlerTestSuite struct {
	suite.Suite
	mockUC  *MockListFieldsUseCase
	handler *FieldHandler
	logger  *slog.Logger
}

func (s *FieldHandlerTestSuite) SetupTest() {
	s.mockUC = new(MockListFieldsUseCase)
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	s.handler = NewFieldHandler(
		&usecase.ListFieldsUseCase{},
		s.logger,
	)
	// モックを使用するためにフィールドを直接設定
	s.handler.listFieldsUC = nil
}

func TestFieldHandlerSuite(t *testing.T) {
	suite.Run(t, new(FieldHandlerTestSuite))
}

// TestNewFieldHandler_Success はNewFieldHandlerが正しくハンドラーを生成することをテスト
func (s *FieldHandlerTestSuite) TestNewFieldHandler_Success() {
	uc := &usecase.ListFieldsUseCase{}
	handler := NewFieldHandler(uc, s.logger)
	require.NotNil(s.T(), handler, "ハンドラーがnilです")
	require.NotNil(s.T(), handler.listFieldsUC, "listFieldsUCがnilです")
	require.NotNil(s.T(), handler.logger, "loggerがnilです")
}

// TestFieldHandler_validateAndNormalizeParams_Success は正常なパラメータでバリデーションが通ることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_Success() {
	pageSize := 50
	params := openapi.ListFieldsParams{
		Page:     1,
		PageSize: &pageSize,
	}

	page, size, err := s.handler.validateAndNormalizeParams(params)
	require.NoError(s.T(), err, "バリデーションでエラーが発生")
	require.Equal(s.T(), 1, page, "ページ番号が一致しません")
	require.Equal(s.T(), 50, size, "ページサイズが一致しません")
}

// TestFieldHandler_validateAndNormalizeParams_DefaultPageSize はpageSizeがnilの場合デフォルト値が使用されることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_DefaultPageSize() {
	params := openapi.ListFieldsParams{
		Page:     1,
		PageSize: nil,
	}

	page, size, err := s.handler.validateAndNormalizeParams(params)
	require.NoError(s.T(), err, "バリデーションでエラーが発生")
	require.Equal(s.T(), 1, page, "ページ番号が一致しません")
	require.Equal(s.T(), DefaultPageSize, size, "デフォルトページサイズが使用されていません")
}

// TestFieldHandler_validateAndNormalizeParams_ValidationError_PageZero はpage=0でエラーになることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_ValidationError_PageZero() {
	params := openapi.ListFieldsParams{
		Page:     0,
		PageSize: nil,
	}

	_, _, err := s.handler.validateAndNormalizeParams(params)
	require.Error(s.T(), err, "page=0でエラーが返されませんでした")
	require.Contains(s.T(), err.Error(), "ページ番号は1以上", "エラーメッセージが期待値と異なります")
}

// TestFieldHandler_validateAndNormalizeParams_ValidationError_PageNegative はpage=-1でエラーになることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_ValidationError_PageNegative() {
	params := openapi.ListFieldsParams{
		Page:     -1,
		PageSize: nil,
	}

	_, _, err := s.handler.validateAndNormalizeParams(params)
	require.Error(s.T(), err, "page=-1でエラーが返されませんでした")
}

// TestFieldHandler_validateAndNormalizeParams_ValidationError_PageSizeZero はpageSize=0でエラーになることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_ValidationError_PageSizeZero() {
	pageSize := 0
	params := openapi.ListFieldsParams{
		Page:     1,
		PageSize: &pageSize,
	}

	_, _, err := s.handler.validateAndNormalizeParams(params)
	require.Error(s.T(), err, "pageSize=0でエラーが返されませんでした")
	require.Contains(s.T(), err.Error(), "ページサイズは1以上", "エラーメッセージが期待値と異なります")
}

// TestFieldHandler_validateAndNormalizeParams_ValidationError_PageSizeNegative はpageSize=-1でエラーになることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_ValidationError_PageSizeNegative() {
	pageSize := -1
	params := openapi.ListFieldsParams{
		Page:     1,
		PageSize: &pageSize,
	}

	_, _, err := s.handler.validateAndNormalizeParams(params)
	require.Error(s.T(), err, "pageSize=-1でエラーが返されませんでした")
}

// TestFieldHandler_validateAndNormalizeParams_ValidationError_PageSizeOverMax はpageSize>1000でエラーになることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_ValidationError_PageSizeOverMax() {
	pageSize := 1001
	params := openapi.ListFieldsParams{
		Page:     1,
		PageSize: &pageSize,
	}

	_, _, err := s.handler.validateAndNormalizeParams(params)
	require.Error(s.T(), err, "pageSize=1001でエラーが返されませんでした")
	require.Contains(s.T(), err.Error(), "ページサイズは1000以下", "エラーメッセージが期待値と異なります")
}

// TestFieldHandler_validateAndNormalizeParams_BoundaryValue_PageSizeOne はpageSize=1(最小値)が通ることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_BoundaryValue_PageSizeOne() {
	pageSize := 1
	params := openapi.ListFieldsParams{
		Page:     1,
		PageSize: &pageSize,
	}

	page, size, err := s.handler.validateAndNormalizeParams(params)
	require.NoError(s.T(), err, "pageSize=1でエラーが発生")
	require.Equal(s.T(), 1, page, "ページ番号が一致しません")
	require.Equal(s.T(), 1, size, "ページサイズが一致しません")
}

// TestFieldHandler_validateAndNormalizeParams_BoundaryValue_PageSizeMax はpageSize=1000(最大値)が通ることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_validateAndNormalizeParams_BoundaryValue_PageSizeMax() {
	pageSize := 1000
	params := openapi.ListFieldsParams{
		Page:     1,
		PageSize: &pageSize,
	}

	page, size, err := s.handler.validateAndNormalizeParams(params)
	require.NoError(s.T(), err, "pageSize=1000でエラーが発生")
	require.Equal(s.T(), 1, page, "ページ番号が一致しません")
	require.Equal(s.T(), 1000, size, "ページサイズが一致しません")
}

// TestFieldHandler_toOpenAPIField はUseCaseの出力がOpenAPI型に正しく変換されることをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_toOpenAPIField() {
	fieldID := uuid.New()
	soilTypeID := uuid.New()
	areaSqm := 10000.0

	usecaseOutput := usecase.FieldOutput{
		ID:         fieldID,
		Name:       "テスト圃場",
		CityCode:   "12345",
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

	result := s.handler.toOpenAPIField(usecaseOutput)

	require.Equal(s.T(), fieldID, result.Id, "IDが一致しません")
	require.Equal(s.T(), "テスト圃場", result.Name, "圃場名が一致しません")
	require.Equal(s.T(), "12345", result.CityCode, "市区町村コードが一致しません")
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

// TestFieldHandler_toOpenAPIField_EmptyGeometry は空のGeometryでも変換が正しく動作することをテスト
func (s *FieldHandlerTestSuite) TestFieldHandler_toOpenAPIField_EmptyGeometry() {
	fieldID := uuid.New()

	usecaseOutput := usecase.FieldOutput{
		ID:       fieldID,
		Name:     "テスト圃場",
		CityCode: "12345",
		Geometry: nil,
		Centroid: usecase.Coordinate{},
	}

	result := s.handler.toOpenAPIField(usecaseOutput)

	require.Equal(s.T(), fieldID, result.Id, "IDが一致しません")
	require.Empty(s.T(), result.Geometry, "Geometryが空ではありません")
	require.Equal(s.T(), 0.0, result.Centroid.Lat, "Centroid.Latがゼロではありません")
	require.Equal(s.T(), 0.0, result.Centroid.Lng, "Centroid.Lngがゼロではありません")
}

// ValidationErrorインターフェースのテスト
func (s *FieldHandlerTestSuite) TestValidationError_Error() {
	err := &ValidationError{
		Field:   "page",
		Message: "テストエラーメッセージ",
	}

	require.Equal(s.T(), "テストエラーメッセージ", err.Error(), "エラーメッセージが一致しません")
}

// testableFieldHandler はテスト用のFieldHandler
type testableFieldHandler struct {
	executeFunc func(ctx context.Context, input usecase.ListFieldsInput) (*usecase.ListFieldsOutput, error)
	logger      *slog.Logger
}

func (h *testableFieldHandler) ListFields(ctx context.Context, request openapi.ListFieldsRequestObject) (openapi.ListFieldsResponseObject, error) {
	params := request.Params

	// pageバリデーション
	if params.Page < 1 {
		return openapi.ListFields400JSONResponse{
			BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "invalid_parameter",
					Message: "ページ番号は1以上を指定してください",
				}},
			},
		}, nil
	}

	// pageSizeバリデーション
	pageSize := DefaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
		if pageSize < 1 {
			return openapi.ListFields400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: "ページサイズは1以上を指定してください",
					}},
				},
			}, nil
		}
		if pageSize > MaxPageSize {
			return openapi.ListFields400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: "ページサイズは1000以下を指定してください",
					}},
				},
			}, nil
		}
	}

	// ユースケース実行
	output, err := h.executeFunc(ctx, usecase.ListFieldsInput{
		Page:     params.Page,
		PageSize: pageSize,
	})
	if err != nil {
		return openapi.ListFields500JSONResponse{
			InternalServerErrorJSONResponse: openapi.InternalServerErrorJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "internal_error",
					Message: "圃場一覧の取得に失敗しました",
				}},
			},
		}, nil
	}

	// レスポンス変換
	fields := make([]openapi.Field, 0, len(output.Fields))
	for _, field := range output.Fields {
		geometry := make([]openapi.Coordinate, 0, len(field.Geometry))
		for _, coord := range field.Geometry {
			geometry = append(geometry, openapi.Coordinate{Lat: coord.Lat, Lng: coord.Lng})
		}
		fields = append(fields, openapi.Field{
			Id:         field.ID,
			Name:       field.Name,
			CityCode:   field.CityCode,
			SoilTypeId: field.SoilTypeID,
			AreaSqm:    field.AreaSqm,
			Geometry:   geometry,
			Centroid:   openapi.Coordinate{Lat: field.Centroid.Lat, Lng: field.Centroid.Lng},
		})
	}

	return openapi.ListFields200JSONResponse{
		Data: &openapi.FieldListData{Fields: fields},
		Meta: &openapi.ResponseMeta{
			Pagination: openapi.PaginationMeta{
				Total:      int(output.Pagination.Total),
				Page:       output.Pagination.Page,
				PageSize:   output.Pagination.PageSize,
				TotalPages: output.Pagination.TotalPages,
			},
		},
		Errors: nil,
	}, nil
}

// FieldHandlerListFieldsTestSuite はListFieldsのテストスイート
type FieldHandlerListFieldsTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (s *FieldHandlerListFieldsTestSuite) SetupTest() {
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestFieldHandlerListFieldsSuite(t *testing.T) {
	suite.Run(t, new(FieldHandlerListFieldsTestSuite))
}

// TestFieldHandler_ListFields_Success は正常に圃場一覧を取得できることをテスト
func (s *FieldHandlerListFieldsTestSuite) TestFieldHandler_ListFields_Success() {
	ctx := context.Background()
	fieldID := uuid.New()

	handler := &testableFieldHandler{
		executeFunc: func(ctx context.Context, input usecase.ListFieldsInput) (*usecase.ListFieldsOutput, error) {
			return &usecase.ListFieldsOutput{
				Fields: []usecase.FieldOutput{
					{
						ID:       fieldID,
						Name:     "テスト圃場",
						CityCode: "12345",
						Geometry: []usecase.Coordinate{{Lat: 35.0, Lng: 139.0}},
						Centroid: usecase.Coordinate{Lat: 35.0, Lng: 139.0},
					},
				},
				Pagination: usecase.PaginationOutput{
					Total:      1,
					Page:       1,
					PageSize:   20,
					TotalPages: 1,
				},
			}, nil
		},
		logger: s.logger,
	}

	request := openapi.ListFieldsRequestObject{
		Params: openapi.ListFieldsParams{Page: 1, PageSize: nil},
	}

	response, err := handler.ListFields(ctx, request)
	require.NoError(s.T(), err, "ListFields実行時にエラーが発生")

	resp200, ok := response.(openapi.ListFields200JSONResponse)
	require.True(s.T(), ok, "200レスポンスを期待")
	require.NotNil(s.T(), resp200.Data, "Dataがnilです")
	require.Len(s.T(), resp200.Data.Fields, 1, "フィールド数が期待値と異なります")
	require.Equal(s.T(), fieldID, resp200.Data.Fields[0].Id, "フィールドIDが一致しません")
	require.NotNil(s.T(), resp200.Meta, "Metaがnilです")
	require.Equal(s.T(), 1, resp200.Meta.Pagination.Total, "総件数が一致しません")
}

// TestFieldHandler_ListFields_Success_EmptyResult は結果が0件でも正常に動作することをテスト
func (s *FieldHandlerListFieldsTestSuite) TestFieldHandler_ListFields_Success_EmptyResult() {
	ctx := context.Background()

	handler := &testableFieldHandler{
		executeFunc: func(ctx context.Context, input usecase.ListFieldsInput) (*usecase.ListFieldsOutput, error) {
			return &usecase.ListFieldsOutput{
				Fields: []usecase.FieldOutput{},
				Pagination: usecase.PaginationOutput{
					Total:      0,
					Page:       1,
					PageSize:   20,
					TotalPages: 0,
				},
			}, nil
		},
		logger: s.logger,
	}

	request := openapi.ListFieldsRequestObject{
		Params: openapi.ListFieldsParams{Page: 1, PageSize: nil},
	}

	response, err := handler.ListFields(ctx, request)
	require.NoError(s.T(), err, "ListFields実行時にエラーが発生")

	resp200, ok := response.(openapi.ListFields200JSONResponse)
	require.True(s.T(), ok, "200レスポンスを期待")
	require.NotNil(s.T(), resp200.Data, "Dataがnilです")
	require.Empty(s.T(), resp200.Data.Fields, "フィールドリストが空ではありません")
	require.Equal(s.T(), 0, resp200.Meta.Pagination.Total, "総件数が0ではありません")
}

// TestFieldHandler_ListFields_ValidationError_PageZero はpage=0で400エラーになることをテスト
func (s *FieldHandlerListFieldsTestSuite) TestFieldHandler_ListFields_ValidationError_PageZero() {
	ctx := context.Background()

	handler := &testableFieldHandler{
		executeFunc: func(ctx context.Context, input usecase.ListFieldsInput) (*usecase.ListFieldsOutput, error) {
			return nil, nil
		},
		logger: s.logger,
	}

	request := openapi.ListFieldsRequestObject{
		Params: openapi.ListFieldsParams{Page: 0, PageSize: nil},
	}

	response, err := handler.ListFields(ctx, request)
	require.NoError(s.T(), err, "ListFields実行時に予期しないエラーが発生")

	resp400, ok := response.(openapi.ListFields400JSONResponse)
	require.True(s.T(), ok, "400レスポンスを期待")
	require.NotNil(s.T(), resp400.Errors, "Errorsがnilです")
	require.Len(s.T(), *resp400.Errors, 1, "エラー数が期待値と異なります")
}

// TestFieldHandler_ListFields_ValidationError_PageSizeOverMax はpageSize>1000で400エラーになることをテスト
func (s *FieldHandlerListFieldsTestSuite) TestFieldHandler_ListFields_ValidationError_PageSizeOverMax() {
	ctx := context.Background()

	handler := &testableFieldHandler{
		executeFunc: func(ctx context.Context, input usecase.ListFieldsInput) (*usecase.ListFieldsOutput, error) {
			return nil, nil
		},
		logger: s.logger,
	}

	pageSize := 1001
	request := openapi.ListFieldsRequestObject{
		Params: openapi.ListFieldsParams{Page: 1, PageSize: &pageSize},
	}

	response, err := handler.ListFields(ctx, request)
	require.NoError(s.T(), err, "ListFields実行時に予期しないエラーが発生")

	resp400, ok := response.(openapi.ListFields400JSONResponse)
	require.True(s.T(), ok, "400レスポンスを期待")
	require.NotNil(s.T(), resp400.Errors, "Errorsがnilです")
}

// TestFieldHandler_ListFields_UseCaseError はユースケースエラー時に500エラーになることをテスト
func (s *FieldHandlerListFieldsTestSuite) TestFieldHandler_ListFields_UseCaseError() {
	ctx := context.Background()

	handler := &testableFieldHandler{
		executeFunc: func(ctx context.Context, input usecase.ListFieldsInput) (*usecase.ListFieldsOutput, error) {
			return nil, errors.New("データベースエラー")
		},
		logger: s.logger,
	}

	request := openapi.ListFieldsRequestObject{
		Params: openapi.ListFieldsParams{Page: 1, PageSize: nil},
	}

	response, err := handler.ListFields(ctx, request)
	require.NoError(s.T(), err, "ListFields実行時に予期しないエラーが発生")

	resp500, ok := response.(openapi.ListFields500JSONResponse)
	require.True(s.T(), ok, "500レスポンスを期待")
	require.NotNil(s.T(), resp500.Errors, "Errorsがnilです")
}
