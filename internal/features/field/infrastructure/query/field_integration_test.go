//go:build integration

package query

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	importdto "github.com/mktkhr/field-manager-api/internal/features/import/domain/dto"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5433")
	user := getEnvOrDefault("TEST_DB_USER", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "postgres")
	dbname := getEnvOrDefault("TEST_DB_NAME", "field_manager_db_test")

	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	testDB, err = pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("テスト用DB接続に失敗: %v", err)
	}
	defer testDB.Close()

	if err := testDB.Ping(ctx); err != nil {
		log.Fatalf("テスト用DBへのPingに失敗: %v", err)
	}

	os.Exit(m.Run())
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// FieldQueryIntegrationTestSuite は圃場クエリの統合テストスイート
type FieldQueryIntegrationTestSuite struct {
	suite.Suite
	query   *fieldQuery
	queries *sqlc.Queries
	logger  *slog.Logger
}

func (s *FieldQueryIntegrationTestSuite) SetupSuite() {
	s.logger = slog.Default()
	s.query = NewFieldQuery(testDB).(*fieldQuery)
	s.queries = sqlc.New(testDB)
}

func (s *FieldQueryIntegrationTestSuite) SetupTest() {
	s.cleanupTestData()
}

func (s *FieldQueryIntegrationTestSuite) TearDownTest() {
	s.cleanupTestData()
}

func (s *FieldQueryIntegrationTestSuite) cleanupTestData() {
	ctx := context.Background()
	_, _ = testDB.Exec(ctx, "DELETE FROM field_land_registries")
	_, _ = testDB.Exec(ctx, "DELETE FROM fields")
}

func TestFieldQueryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(FieldQueryIntegrationTestSuite))
}

// createTestField はテスト用の圃場を作成する
func (s *FieldQueryIntegrationTestSuite) createTestField(fieldID uuid.UUID, cityCode string, coords [][]float64) {
	ctx := context.Background()

	// ポリゴン作成
	polygon := geom.NewPolygon(geom.XY)
	geomCoords := make([]geom.Coord, len(coords))
	for i, c := range coords {
		geomCoords[i] = geom.Coord{c[0], c[1]} // lng, lat
	}
	_, err := polygon.SetCoords([][]geom.Coord{geomCoords})
	require.NoError(s.T(), err, "ポリゴン座標設定に失敗")

	// 重心計算
	var sumX, sumY float64
	for _, c := range coords {
		sumX += c[0]
		sumY += c[1]
	}
	numPoints := float64(len(coords))
	centroid := geom.NewPoint(geom.XY)
	_, err = centroid.SetCoords(geom.Coord{sumX / numPoints, sumY / numPoints})
	require.NoError(s.T(), err, "重心座標設定に失敗")

	// WKB変換
	geometryWKB, err := wkb.Marshal(polygon, wkb.NDR)
	require.NoError(s.T(), err, "ジオメトリWKB変換に失敗")
	centroidWKB, err := wkb.Marshal(centroid, wkb.NDR)
	require.NoError(s.T(), err, "重心WKB変換に失敗")

	// DB挿入
	_, err = s.queries.UpsertField(ctx, &sqlc.UpsertFieldParams{
		ID:          fieldID,
		GeometryWkb: geometryWKB,
		CentroidWkb: centroidWKB,
		CityCode:    cityCode,
	})
	require.NoError(s.T(), err, "圃場挿入に失敗")
}

// TestFieldQuery_List_Integration は実際のDBから圃場一覧を取得できることをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_List_Integration() {
	ctx := context.Background()

	// テストデータ作成
	fieldID := uuid.New()
	coords := [][]float64{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}
	s.createTestField(fieldID, "12345", coords)

	// 一覧取得
	fields, err := s.query.List(ctx, 10, 0)
	require.NoError(s.T(), err, "List実行時にエラーが発生")
	require.Len(s.T(), fields, 1, "取得件数が期待値と異なります")
	require.Equal(s.T(), fieldID, fields[0].ID, "圃場IDが一致しません")
	require.Equal(s.T(), "12345", fields[0].CityCode, "市区町村コードが一致しません")
}

// TestFieldQuery_List_Integration_GeometryConversion はポリゴン座標が正しい順序で変換されることをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_List_Integration_GeometryConversion() {
	ctx := context.Background()

	// テストデータ作成(経度139.0-139.1、緯度35.0-35.1の四角形)
	fieldID := uuid.New()
	coords := [][]float64{
		{139.0, 35.0}, // 南西
		{139.1, 35.0}, // 南東
		{139.1, 35.1}, // 北東
		{139.0, 35.1}, // 北西
		{139.0, 35.0}, // 閉じる
	}
	s.createTestField(fieldID, "12345", coords)

	// 一覧取得
	fields, err := s.query.List(ctx, 10, 0)
	require.NoError(s.T(), err, "List実行時にエラーが発生")
	require.Len(s.T(), fields, 1, "取得件数が期待値と異なります")

	// Geometry変換の検証
	require.NotNil(s.T(), fields[0].Geometry, "Geometryがnilです")
	flatCoords := fields[0].Geometry.FlatCoords()
	stride := fields[0].Geometry.Stride()
	numPoints := len(flatCoords) / stride

	require.Equal(s.T(), 5, numPoints, "ポリゴン頂点数が期待値と異なります")

	// 最初の頂点を検証(go-geomではX=経度、Y=緯度)
	require.InDelta(s.T(), 139.0, flatCoords[0], 0.001, "1番目の頂点の経度(X)が一致しません")
	require.InDelta(s.T(), 35.0, flatCoords[1], 0.001, "1番目の頂点の緯度(Y)が一致しません")

	// Centroid変換の検証
	require.NotNil(s.T(), fields[0].Centroid, "Centroidがnilです")
	// 重心は(139.04, 35.04)程度になるはず((139.0+139.1+139.1+139.0+139.0)/5, (35.0+35.0+35.1+35.1+35.0)/5)
	require.InDelta(s.T(), 139.04, fields[0].Centroid.X(), 0.01, "重心の経度(X)が一致しません")
	require.InDelta(s.T(), 35.04, fields[0].Centroid.Y(), 0.01, "重心の緯度(Y)が一致しません")
}

// TestFieldQuery_Count_Integration は実際のDBから圃場総数を取得できることをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_Count_Integration() {
	ctx := context.Background()

	// テストデータ作成
	coords := [][]float64{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}
	s.createTestField(uuid.New(), "12345", coords)
	s.createTestField(uuid.New(), "12346", coords)
	s.createTestField(uuid.New(), "12347", coords)

	// 総数取得
	count, err := s.query.Count(ctx)
	require.NoError(s.T(), err, "Count実行時にエラーが発生")
	require.Equal(s.T(), int64(3), count, "総件数が期待値と異なります")
}

// TestFieldQuery_List_Integration_Pagination はページネーションが正しく動作することをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_List_Integration_Pagination() {
	ctx := context.Background()

	// テストデータ作成(5件)
	coords := [][]float64{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}
	for i := 0; i < 5; i++ {
		s.createTestField(uuid.New(), fmt.Sprintf("1234%d", i), coords)
	}

	// 1ページ目(limit=2, offset=0)
	fields1, err := s.query.List(ctx, 2, 0)
	require.NoError(s.T(), err, "List(limit=2, offset=0)実行時にエラーが発生")
	require.Len(s.T(), fields1, 2, "1ページ目の取得件数が期待値と異なります")

	// 2ページ目(limit=2, offset=2)
	fields2, err := s.query.List(ctx, 2, 2)
	require.NoError(s.T(), err, "List(limit=2, offset=2)実行時にエラーが発生")
	require.Len(s.T(), fields2, 2, "2ページ目の取得件数が期待値と異なります")

	// 3ページ目(limit=2, offset=4)
	fields3, err := s.query.List(ctx, 2, 4)
	require.NoError(s.T(), err, "List(limit=2, offset=4)実行時にエラーが発生")
	require.Len(s.T(), fields3, 1, "3ページ目の取得件数が期待値と異なります")

	// 範囲外(limit=2, offset=10)
	fields4, err := s.query.List(ctx, 2, 10)
	require.NoError(s.T(), err, "List(limit=2, offset=10)実行時にエラーが発生")
	require.Len(s.T(), fields4, 0, "範囲外の取得件数が0ではありません")

	// 各ページのIDが重複していないことを確認
	allIDs := make(map[uuid.UUID]bool)
	for _, f := range fields1 {
		allIDs[f.ID] = true
	}
	for _, f := range fields2 {
		require.False(s.T(), allIDs[f.ID], "2ページ目のIDが1ページ目と重複しています")
		allIDs[f.ID] = true
	}
	for _, f := range fields3 {
		require.False(s.T(), allIDs[f.ID], "3ページ目のIDが前のページと重複しています")
	}
}

// TestFieldQuery_List_Integration_Empty は圃場が0件の場合も正常に動作することをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_List_Integration_Empty() {
	ctx := context.Background()

	// データなしで一覧取得
	fields, err := s.query.List(ctx, 10, 0)
	require.NoError(s.T(), err, "List実行時にエラーが発生")
	require.Empty(s.T(), fields, "圃場リストが空ではありません")
}

// TestFieldQuery_Count_Integration_Empty は圃場が0件の場合のカウントをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_Count_Integration_Empty() {
	ctx := context.Background()

	// データなしで総数取得
	count, err := s.query.Count(ctx)
	require.NoError(s.T(), err, "Count実行時にエラーが発生")
	require.Equal(s.T(), int64(0), count, "総件数が0ではありません")
}

// TestFieldQuery_List_Integration_WithSoilType は土壌タイプIDを持つ圃場が正しく取得できることをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_List_Integration_WithSoilType() {
	ctx := context.Background()

	// 土壌タイプを作成
	soilType, err := s.queries.UpsertSoilType(ctx, &sqlc.UpsertSoilTypeParams{
		LargeCode:  "A",
		MiddleCode: "A1",
		SmallCode:  "A1a",
		SmallName:  "テスト土壌",
	})
	require.NoError(s.T(), err, "土壌タイプ作成に失敗")

	// テストデータ作成(土壌タイプ付き)
	fieldID := uuid.New()
	coords := [][]float64{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}

	// ポリゴン作成
	polygon := geom.NewPolygon(geom.XY)
	geomCoords := make([]geom.Coord, len(coords))
	for i, c := range coords {
		geomCoords[i] = geom.Coord{c[0], c[1]}
	}
	_, err = polygon.SetCoords([][]geom.Coord{geomCoords})
	require.NoError(s.T(), err, "ポリゴン座標設定に失敗")

	// 重心計算
	centroid := geom.NewPoint(geom.XY)
	_, err = centroid.SetCoords(geom.Coord{139.05, 35.05})
	require.NoError(s.T(), err, "重心座標設定に失敗")

	// WKB変換
	geometryWKB, err := wkb.Marshal(polygon, wkb.NDR)
	require.NoError(s.T(), err, "ジオメトリWKB変換に失敗")
	centroidWKB, err := wkb.Marshal(centroid, wkb.NDR)
	require.NoError(s.T(), err, "重心WKB変換に失敗")

	// DB挿入(土壌タイプ付き)
	_, err = s.queries.UpsertField(ctx, &sqlc.UpsertFieldParams{
		ID:          fieldID,
		GeometryWkb: geometryWKB,
		CentroidWkb: centroidWKB,
		CityCode:    "12345",
		SoilTypeID:  uuid.NullUUID{UUID: soilType.ID, Valid: true},
	})
	require.NoError(s.T(), err, "圃場挿入に失敗")

	// 一覧取得
	fields, err := s.query.List(ctx, 10, 0)
	require.NoError(s.T(), err, "List実行時にエラーが発生")
	require.Len(s.T(), fields, 1, "取得件数が期待値と異なります")
	require.NotNil(s.T(), fields[0].SoilTypeID, "SoilTypeIDがnilです")
	require.Equal(s.T(), soilType.ID, *fields[0].SoilTypeID, "SoilTypeIDが一致しません")
}

// importdtoパッケージのダミー使用(インポート維持のため)
var _ = importdto.FieldBatchInput{}
