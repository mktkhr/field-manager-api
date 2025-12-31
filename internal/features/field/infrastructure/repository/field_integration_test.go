//go:build integration

package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/shared/types"
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

func cleanupTestData(t *testing.T, ctx context.Context) {
	t.Helper()
	// 依存関係の順序で削除
	_, _ = testDB.Exec(ctx, "DELETE FROM field_land_registries")
	_, _ = testDB.Exec(ctx, "DELETE FROM fields")
}

func TestNewFieldRepository_Integration(t *testing.T) {
	repo := NewFieldRepository(testDB)

	if repo == nil {
		t.Error("NewFieldRepository() returned nil")
	}
}

func TestFieldRepository_Create_Integration(t *testing.T) {
	ctx := context.Background()
	repo := NewFieldRepository(testDB)

	field := entity.NewField(uuid.New(), "163210")

	// Create is a stub that returns nil
	err := repo.Create(ctx, field)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
}

func TestFieldRepository_Update_Integration(t *testing.T) {
	ctx := context.Background()
	repo := NewFieldRepository(testDB)

	field := entity.NewField(uuid.New(), "163210")

	// Update is a stub that returns nil
	err := repo.Update(ctx, field)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
}

func TestFieldRepository_Upsert_Integration(t *testing.T) {
	ctx := context.Background()
	repo := NewFieldRepository(testDB)

	field := entity.NewField(uuid.New(), "163210")

	// Upsert is a stub that returns nil
	err := repo.Upsert(ctx, field)
	if err != nil {
		t.Errorf("Upsert() error = %v", err)
	}
}

func TestFieldRepository_Delete_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	// 存在しないIDでの削除（エラーにはならない）
	err := repo.Delete(ctx, uuid.New())
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestFieldRepository_FindByID_NotFound_Integration(t *testing.T) {
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	// 存在しないIDで検索
	_, err := repo.FindByID(ctx, uuid.New())
	if err == nil {
		t.Error("FindByID() expected error for non-existent ID, got nil")
	}
}

func TestFieldRepository_UpsertBatch_EmptyFeatures_Integration(t *testing.T) {
	// 空の入力リストでUpsertBatchが正常に動作することを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	// 空の入力リストでUpsertBatch
	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{})
	if err != nil {
		t.Errorf("UpsertBatch() with empty features error = %v", err)
	}
}

func TestFieldRepository_UpsertBatch_SingleFeature_Integration(t *testing.T) {
	// 単一の入力でUpsertBatchが正常に動作することを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	fieldID := uuid.New()
	input := types.FieldBatchInput{
		ID:       fieldID.String(),
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{139.6917, 35.6895},
					{139.6920, 35.6895},
					{139.6920, 35.6898},
					{139.6917, 35.6898},
					{139.6917, 35.6895},
				},
			},
		},
	}

	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err != nil {
		t.Errorf("UpsertBatch() error = %v", err)
	}

	// 作成されたフィールドを確認
	found, err := repo.FindByID(ctx, fieldID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.CityCode != "163210" {
		t.Errorf("CityCode = %q, want %q", found.CityCode, "163210")
	}
}

func TestFieldRepository_UpsertBatch_WithSoilType_Integration(t *testing.T) {
	// 土壌タイプ情報を含む入力でUpsertBatchが正常に動作することを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	fieldID := uuid.New()
	input := types.FieldBatchInput{
		ID:       fieldID.String(),
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{139.6917, 35.6895},
					{139.6920, 35.6895},
					{139.6920, 35.6898},
					{139.6917, 35.6898},
					{139.6917, 35.6895},
				},
			},
		},
		SoilType: &types.FieldBatchSoilType{
			LargeCode:  "A",
			MiddleCode: "A1",
			SmallCode:  "A1a",
			SmallName:  "テスト土壌",
		},
	}

	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err != nil {
		t.Errorf("UpsertBatch() with soil type error = %v", err)
	}

	found, err := repo.FindByID(ctx, fieldID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.SoilTypeID == nil {
		t.Error("SoilTypeID should not be nil")
	}
}

func TestFieldRepository_UpsertBatch_WithPinInfo_Integration(t *testing.T) {
	// 農地台帳情報を含む入力でUpsertBatchが正常に動作することを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	fieldID := uuid.New()
	descriptiveStudyDataRaw := "2024-01-15"
	input := types.FieldBatchInput{
		ID:       fieldID.String(),
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{139.6917, 35.6895},
					{139.6920, 35.6895},
					{139.6920, 35.6898},
					{139.6917, 35.6898},
					{139.6917, 35.6895},
				},
			},
		},
		PinInfoList: []types.FieldBatchPinInfo{
			{
				FarmerNumber:            "F001",
				Address:                 "東京都渋谷区1-1-1",
				Area:                    1234,
				LandCategoryCode:        "01",
				LandCategory:            "田",
				IdleLandStatusCode:      "1",
				IdleLandStatus:          "遊休農地ではない",
				DescriptiveStudyDataRaw: &descriptiveStudyDataRaw,
			},
		},
	}

	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err != nil {
		t.Errorf("UpsertBatch() with PinInfo error = %v", err)
	}

	found, err := repo.FindByID(ctx, fieldID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.ID != fieldID {
		t.Errorf("ID = %v, want %v", found.ID, fieldID)
	}
}

func TestFieldRepository_UpsertBatch_InvalidFieldID_Integration(t *testing.T) {
	// 不正なフィールドIDでUpsertBatchがエラーを返すことを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	input := types.FieldBatchInput{
		ID:       "invalid-uuid",
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{139.6917, 35.6895},
					{139.6920, 35.6895},
					{139.6920, 35.6898},
					{139.6917, 35.6898},
					{139.6917, 35.6895},
				},
			},
		},
	}

	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err == nil {
		t.Error("UpsertBatch() with invalid field ID should return error")
	}
}

func TestFieldRepository_UpsertBatch_InvalidGeometry_Integration(t *testing.T) {
	// 不正なジオメトリでUpsertBatchがエラーを返すことを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	fieldID := uuid.New()
	input := types.FieldBatchInput{
		ID:       fieldID.String(),
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					// 2点だけ - 不正なポリゴン
					{139.6917, 35.6895},
					{139.6920, 35.6895},
				},
			},
		},
	}

	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err == nil {
		t.Error("UpsertBatch() with invalid geometry should return error")
	}
}

func TestFieldRepository_UpsertBatch_MultipleFeatures_Integration(t *testing.T) {
	// 複数の入力でUpsertBatchが正常に動作することを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	fieldID1 := uuid.New()
	fieldID2 := uuid.New()

	inputs := []types.FieldBatchInput{
		{
			ID:       fieldID1.String(),
			CityCode: "163210",
			Geometry: types.FieldBatchGeometry{
				Type: "Polygon",
				Coordinates: [][][]float64{
					{
						{139.6917, 35.6895},
						{139.6920, 35.6895},
						{139.6920, 35.6898},
						{139.6917, 35.6898},
						{139.6917, 35.6895},
					},
				},
			},
		},
		{
			ID:       fieldID2.String(),
			CityCode: "131016",
			Geometry: types.FieldBatchGeometry{
				Type: "Polygon",
				Coordinates: [][][]float64{
					{
						{139.7000, 35.7000},
						{139.7010, 35.7000},
						{139.7010, 35.7010},
						{139.7000, 35.7010},
						{139.7000, 35.7000},
					},
				},
			},
		},
	}

	err := repo.UpsertBatch(ctx, inputs)
	if err != nil {
		t.Errorf("UpsertBatch() with multiple features error = %v", err)
	}

	// 両方のフィールドを確認
	found1, err := repo.FindByID(ctx, fieldID1)
	if err != nil {
		t.Fatalf("FindByID() for field1 error = %v", err)
	}
	if found1.CityCode != "163210" {
		t.Errorf("field1.CityCode = %q, want %q", found1.CityCode, "163210")
	}

	found2, err := repo.FindByID(ctx, fieldID2)
	if err != nil {
		t.Fatalf("FindByID() for field2 error = %v", err)
	}
	if found2.CityCode != "131016" {
		t.Errorf("field2.CityCode = %q, want %q", found2.CityCode, "131016")
	}
}

func TestFieldRepository_UpsertBatch_UpdateExisting_Integration(t *testing.T) {
	// 既存のレコードをUpsertBatchで更新できることを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	fieldID := uuid.New()
	input := types.FieldBatchInput{
		ID:       fieldID.String(),
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{139.6917, 35.6895},
					{139.6920, 35.6895},
					{139.6920, 35.6898},
					{139.6917, 35.6898},
					{139.6917, 35.6895},
				},
			},
		},
	}

	// 最初のUpsert
	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err != nil {
		t.Fatalf("First UpsertBatch() error = %v", err)
	}

	// 2回目のUpsert(同じIDで更新)
	err = repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err != nil {
		t.Errorf("Second UpsertBatch() error = %v", err)
	}

	found, err := repo.FindByID(ctx, fieldID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.ID != fieldID {
		t.Errorf("ID = %v, want %v", found.ID, fieldID)
	}
}

func TestFieldRepository_FindByID_Error_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewFieldRepository(testDB)

	_, err := repo.FindByID(ctx, uuid.New())
	if err == nil {
		t.Error("FindByID() with cancelled context should return error")
	}
}

func TestFieldRepository_Delete_Error_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewFieldRepository(testDB)

	err := repo.Delete(ctx, uuid.New())
	if err == nil {
		t.Error("Delete() with cancelled context should return error")
	}
}

func TestFieldRepository_UpsertBatch_TransactionError_Integration(t *testing.T) {
	// キャンセルされたコンテキストでUpsertBatchがエラーを返すことを確認する
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewFieldRepository(testDB)

	input := types.FieldBatchInput{
		ID:       uuid.New().String(),
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type:        "Polygon",
			Coordinates: [][][]float64{{{139.6917, 35.6895}, {139.6920, 35.6895}, {139.6920, 35.6898}, {139.6917, 35.6898}, {139.6917, 35.6895}}},
		},
	}

	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err == nil {
		t.Error("UpsertBatch() with cancelled context should return error")
	}
}

func TestFieldRepository_UpsertBatch_EmptyGeometry_Integration(t *testing.T) {
	// 空のジオメトリでUpsertBatchがエラーを返すことを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	fieldID := uuid.New()
	input := types.FieldBatchInput{
		ID:       fieldID.String(),
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type:        "Polygon",
			Coordinates: [][][]float64{}, // 空のコーディネート
		},
	}

	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err == nil {
		t.Error("UpsertBatch() with empty geometry should return error")
	}
}

func TestFieldRepository_UpsertBatch_TwoPointsGeometry_Integration(t *testing.T) {
	// 2点のみのポリゴンでUpsertBatchがエラーを返すことを確認する
	ctx := context.Background()
	cleanupTestData(t, ctx)

	repo := NewFieldRepository(testDB)

	fieldID := uuid.New()
	// 2点だけのポリゴン - geom.SetCoordsでエラーになるはず
	input := types.FieldBatchInput{
		ID:       fieldID.String(),
		CityCode: "163210",
		Geometry: types.FieldBatchGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{139.6917, 35.6895},
					{139.6920, 35.6895},
				},
			},
		},
	}

	err := repo.UpsertBatch(ctx, []types.FieldBatchInput{input})
	if err == nil {
		t.Error("UpsertBatch() with two points polygon should return error")
	}
}
