//go:build integration

package query

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
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

// createTestFieldWithTimestamp はテスト用の圃場を作成する(タイムスタンプ指定可能)
func (s *FieldQueryIntegrationTestSuite) createTestFieldWithTimestamp(fieldID uuid.UUID, cityCode string, coords [][]float64, createdAt time.Time) {
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

	// created_atを指定値に更新
	_, err = testDB.Exec(ctx, "UPDATE fields SET created_at = $1 WHERE id = $2", createdAt, fieldID)
	require.NoError(s.T(), err, "created_at更新に失敗")
}

// createTestField はテスト用の圃場を作成する
func (s *FieldQueryIntegrationTestSuite) createTestField(fieldID uuid.UUID, cityCode string, coords [][]float64) {
	s.createTestFieldWithTimestamp(fieldID, cityCode, coords, time.Now())
}

// TestFieldQuery_ListByCursor_Integration は実際のDBから圃場一覧を取得できることをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_ListByCursor_Integration() {
	ctx := context.Background()

	// テストデータ作成
	fieldID := uuid.New()
	coords := [][]float64{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}
	s.createTestField(fieldID, "12345", coords)

	// カーソルなしで一覧取得
	fields, err := s.query.ListByCursor(ctx, nil, 10)
	require.NoError(s.T(), err, "ListByCursor実行時にエラーが発生")
	require.Len(s.T(), fields, 1, "取得件数が期待値と異なります")
	require.Equal(s.T(), fieldID, fields[0].ID, "圃場IDが一致しません")
	require.Equal(s.T(), "12345", fields[0].CityCode, "市区町村コードが一致しません")
}

// TestFieldQuery_ListByCursor_Integration_GeometryConversion はポリゴン座標が正しい順序で変換されることをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_ListByCursor_Integration_GeometryConversion() {
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
	fields, err := s.query.ListByCursor(ctx, nil, 10)
	require.NoError(s.T(), err, "ListByCursor実行時にエラーが発生")
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

// TestFieldQuery_ListByCursor_Integration_CursorPagination はカーソルベースページネーションが正しく動作することをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_ListByCursor_Integration_CursorPagination() {
	ctx := context.Background()

	// テストデータ作成(5件、created_atを明示的に設定)
	coords := [][]float64{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}
	now := time.Now()
	fieldIDs := make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		fieldIDs[i] = uuid.New()
		// created_atを1秒ずつずらして明確な順序を作る(新しい順: 4, 3, 2, 1, 0)
		createdAt := now.Add(-time.Duration(i) * time.Second)
		s.createTestFieldWithTimestamp(fieldIDs[i], fmt.Sprintf("1234%d", i), coords, createdAt)
	}

	// 1ページ目(limit=2, cursor=nil)
	fields1, err := s.query.ListByCursor(ctx, nil, 2)
	require.NoError(s.T(), err, "ListByCursor(limit=2, cursor=nil)実行時にエラーが発生")
	require.Len(s.T(), fields1, 2, "1ページ目の取得件数が期待値と異なります")
	// 新しい順(created_at DESC)なので、fieldIDs[0]が最初に来る
	require.Equal(s.T(), fieldIDs[0], fields1[0].ID, "1ページ目1件目のIDが期待値と異なります")
	require.Equal(s.T(), fieldIDs[1], fields1[1].ID, "1ページ目2件目のIDが期待値と異なります")

	// 2ページ目(カーソルを使用)
	cursor1 := entity.NewFieldCursor(fields1[1].CreatedAt, fields1[1].ID)
	fields2, err := s.query.ListByCursor(ctx, cursor1, 2)
	require.NoError(s.T(), err, "ListByCursor(2ページ目)実行時にエラーが発生")
	require.Len(s.T(), fields2, 2, "2ページ目の取得件数が期待値と異なります")
	require.Equal(s.T(), fieldIDs[2], fields2[0].ID, "2ページ目1件目のIDが期待値と異なります")
	require.Equal(s.T(), fieldIDs[3], fields2[1].ID, "2ページ目2件目のIDが期待値と異なります")

	// 3ページ目(カーソルを使用)
	cursor2 := entity.NewFieldCursor(fields2[1].CreatedAt, fields2[1].ID)
	fields3, err := s.query.ListByCursor(ctx, cursor2, 2)
	require.NoError(s.T(), err, "ListByCursor(3ページ目)実行時にエラーが発生")
	require.Len(s.T(), fields3, 1, "3ページ目の取得件数が期待値と異なります")
	require.Equal(s.T(), fieldIDs[4], fields3[0].ID, "3ページ目1件目のIDが期待値と異なります")

	// 4ページ目(データなし)
	cursor3 := entity.NewFieldCursor(fields3[0].CreatedAt, fields3[0].ID)
	fields4, err := s.query.ListByCursor(ctx, cursor3, 2)
	require.NoError(s.T(), err, "ListByCursor(4ページ目)実行時にエラーが発生")
	require.Len(s.T(), fields4, 0, "4ページ目の取得件数が0ではありません")

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

// TestFieldQuery_ListByCursor_Integration_SameCreatedAt は同一created_atでもIDで正しくソートされることをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_ListByCursor_Integration_SameCreatedAt() {
	ctx := context.Background()

	// テストデータ作成(3件、同一created_at)
	coords := [][]float64{{139.0, 35.0}, {139.1, 35.0}, {139.1, 35.1}, {139.0, 35.1}, {139.0, 35.0}}
	sameTime := time.Now()
	fieldIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		fieldIDs[i] = uuid.New()
		s.createTestFieldWithTimestamp(fieldIDs[i], fmt.Sprintf("1234%d", i), coords, sameTime)
	}

	// 全件取得して順序を確認
	fields, err := s.query.ListByCursor(ctx, nil, 10)
	require.NoError(s.T(), err, "ListByCursor実行時にエラーが発生")
	require.Len(s.T(), fields, 3, "取得件数が期待値と異なります")

	// IDの降順でソートされていることを確認
	for i := 0; i < len(fields)-1; i++ {
		// UUIDの文字列比較(降順)
		require.Greater(s.T(), fields[i].ID.String(), fields[i+1].ID.String(),
			"ID順序が降順ではありません(index %d, %d)", i, i+1)
	}

	// カーソルを使用して次ページを取得
	cursor := entity.NewFieldCursor(fields[0].CreatedAt, fields[0].ID)
	fields2, err := s.query.ListByCursor(ctx, cursor, 10)
	require.NoError(s.T(), err, "ListByCursor(カーソル指定)実行時にエラーが発生")
	require.Len(s.T(), fields2, 2, "カーソル以降の取得件数が期待値と異なります")
	require.Equal(s.T(), fields[1].ID, fields2[0].ID, "カーソル後の1件目が期待値と異なります")
	require.Equal(s.T(), fields[2].ID, fields2[1].ID, "カーソル後の2件目が期待値と異なります")
}

// TestFieldQuery_ListByCursor_Integration_Empty は圃場が0件の場合も正常に動作することをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_ListByCursor_Integration_Empty() {
	ctx := context.Background()

	// データなしで一覧取得
	fields, err := s.query.ListByCursor(ctx, nil, 10)
	require.NoError(s.T(), err, "ListByCursor実行時にエラーが発生")
	require.Empty(s.T(), fields, "圃場リストが空ではありません")
}

// TestFieldQuery_ListByCursor_Integration_WithSoilType は土壌タイプIDを持つ圃場が正しく取得できることをテスト
func (s *FieldQueryIntegrationTestSuite) TestFieldQuery_ListByCursor_Integration_WithSoilType() {
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
	fields, err := s.query.ListByCursor(ctx, nil, 10)
	require.NoError(s.T(), err, "ListByCursor実行時にエラーが発生")
	require.Len(s.T(), fields, 1, "取得件数が期待値と異なります")
	require.NotNil(s.T(), fields[0].SoilTypeID, "SoilTypeIDがnilです")
	require.Equal(s.T(), soilType.ID, *fields[0].SoilTypeID, "SoilTypeIDが一致しません")
}

// importdtoパッケージのダミー使用(インポート維持のため)
var _ = importdto.FieldBatchInput{}
