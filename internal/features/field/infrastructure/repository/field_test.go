package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
	"github.com/twpayne/go-geom"
)

// TestFieldRepository_ToEntity はtoEntityメソッドがsqlc.FieldをEntity.Fieldに正しく変換することをテストする
func TestFieldRepository_ToEntity(t *testing.T) {
	now := time.Now()
	h3Res3 := "831f8dfffffffff"
	h3Res5 := "851f8d3ffffffff"
	h3Res7 := "871f8d3a7ffffff"
	h3Res9 := "891f8d3a4bfffff"
	areaSqm := 1234.56
	soilTypeID := uuid.New()
	createdBy := uuid.New()
	updatedBy := uuid.New()

	tests := []struct {
		name string
		row  *sqlc.Field
		want *entity.Field
	}{
		{
			name: "nil row",
			row:  nil,
			want: nil,
		},
		{
			name: "minimal field",
			row: &sqlc.Field{
				ID:        uuid.New(),
				CityCode:  "163210",
				Name:      "テスト圃場",
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.Field{
				CityCode: "163210",
				Name:     "テスト圃場",
			},
		},
		{
			name: "field with all H3 indexes",
			row: &sqlc.Field{
				ID:          uuid.New(),
				CityCode:    "163210",
				Name:        "テスト圃場",
				H3IndexRes3: &h3Res3,
				H3IndexRes5: &h3Res5,
				H3IndexRes7: &h3Res7,
				H3IndexRes9: &h3Res9,
				CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.Field{
				CityCode:    "163210",
				Name:        "テスト圃場",
				H3IndexRes3: &h3Res3,
				H3IndexRes5: &h3Res5,
				H3IndexRes7: &h3Res7,
				H3IndexRes9: &h3Res9,
			},
		},
		{
			name: "field with area and soil type",
			row: &sqlc.Field{
				ID:         uuid.New(),
				CityCode:   "163210",
				Name:       "テスト圃場",
				AreaSqm:    &areaSqm,
				SoilTypeID: uuid.NullUUID{UUID: soilTypeID, Valid: true},
				CreatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.Field{
				CityCode:   "163210",
				Name:       "テスト圃場",
				AreaSqm:    &areaSqm,
				SoilTypeID: &soilTypeID,
			},
		},
		{
			name: "field with audit fields",
			row: &sqlc.Field{
				ID:        uuid.New(),
				CityCode:  "163210",
				Name:      "テスト圃場",
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
				CreatedBy: uuid.NullUUID{UUID: createdBy, Valid: true},
				UpdatedBy: uuid.NullUUID{UUID: updatedBy, Valid: true},
			},
			want: &entity.Field{
				CityCode:  "163210",
				Name:      "テスト圃場",
				CreatedBy: &createdBy,
				UpdatedBy: &updatedBy,
			},
		},
		{
			name: "full field with all fields",
			row: &sqlc.Field{
				ID:          uuid.New(),
				CityCode:    "163210",
				Name:        "テスト圃場",
				AreaSqm:     &areaSqm,
				H3IndexRes3: &h3Res3,
				H3IndexRes5: &h3Res5,
				H3IndexRes7: &h3Res7,
				H3IndexRes9: &h3Res9,
				SoilTypeID:  uuid.NullUUID{UUID: soilTypeID, Valid: true},
				CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
				CreatedBy:   uuid.NullUUID{UUID: createdBy, Valid: true},
				UpdatedBy:   uuid.NullUUID{UUID: updatedBy, Valid: true},
			},
			want: &entity.Field{
				CityCode:    "163210",
				Name:        "テスト圃場",
				AreaSqm:     &areaSqm,
				H3IndexRes3: &h3Res3,
				H3IndexRes5: &h3Res5,
				H3IndexRes7: &h3Res7,
				H3IndexRes9: &h3Res9,
				SoilTypeID:  &soilTypeID,
				CreatedBy:   &createdBy,
				UpdatedBy:   &updatedBy,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &fieldRepository{}
			result := r.toEntity(tt.row)

			if tt.want == nil {
				if result != nil {
					t.Errorf("toEntity() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Error("toEntity() = nil, want non-nil")
				return
			}

			if result.CityCode != tt.want.CityCode {
				t.Errorf("CityCode = %q, want %q", result.CityCode, tt.want.CityCode)
			}
			if result.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", result.Name, tt.want.Name)
			}
		})
	}
}

// TestFieldRepository_ToEntity_FieldMapping はtoEntityメソッドが全フィールドを正しくマッピングすることをテストする
func TestFieldRepository_ToEntity_FieldMapping(t *testing.T) {
	now := time.Now()
	h3Res3 := "831f8dfffffffff"
	h3Res5 := "851f8d3ffffffff"
	h3Res7 := "871f8d3a7ffffff"
	h3Res9 := "891f8d3a4bfffff"
	areaSqm := 1234.56
	soilTypeID := uuid.New()
	createdBy := uuid.New()
	updatedBy := uuid.New()
	id := uuid.New()

	row := &sqlc.Field{
		ID:          id,
		CityCode:    "163210",
		Name:        "テスト圃場",
		AreaSqm:     &areaSqm,
		H3IndexRes3: &h3Res3,
		H3IndexRes5: &h3Res5,
		H3IndexRes7: &h3Res7,
		H3IndexRes9: &h3Res9,
		SoilTypeID:  uuid.NullUUID{UUID: soilTypeID, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
		CreatedBy:   uuid.NullUUID{UUID: createdBy, Valid: true},
		UpdatedBy:   uuid.NullUUID{UUID: updatedBy, Valid: true},
	}

	r := &fieldRepository{}
	result := r.toEntity(row)

	// ID
	if result.ID != id {
		t.Errorf("ID = %v, want %v", result.ID, id)
	}

	// CityCode
	if result.CityCode != "163210" {
		t.Errorf("CityCode = %q, want %q", result.CityCode, "163210")
	}

	// Name
	if result.Name != "テスト圃場" {
		t.Errorf("Name = %q, want %q", result.Name, "テスト圃場")
	}

	// AreaSqm
	if result.AreaSqm == nil || *result.AreaSqm != areaSqm {
		t.Errorf("AreaSqm = %v, want %v", result.AreaSqm, &areaSqm)
	}

	// H3IndexRes3
	if result.H3IndexRes3 == nil || *result.H3IndexRes3 != h3Res3 {
		t.Errorf("H3IndexRes3 = %v, want %v", result.H3IndexRes3, &h3Res3)
	}

	// H3IndexRes5
	if result.H3IndexRes5 == nil || *result.H3IndexRes5 != h3Res5 {
		t.Errorf("H3IndexRes5 = %v, want %v", result.H3IndexRes5, &h3Res5)
	}

	// H3IndexRes7
	if result.H3IndexRes7 == nil || *result.H3IndexRes7 != h3Res7 {
		t.Errorf("H3IndexRes7 = %v, want %v", result.H3IndexRes7, &h3Res7)
	}

	// H3IndexRes9
	if result.H3IndexRes9 == nil || *result.H3IndexRes9 != h3Res9 {
		t.Errorf("H3IndexRes9 = %v, want %v", result.H3IndexRes9, &h3Res9)
	}

	// SoilTypeID
	if result.SoilTypeID == nil || *result.SoilTypeID != soilTypeID {
		t.Errorf("SoilTypeID = %v, want %v", result.SoilTypeID, &soilTypeID)
	}

	// CreatedAt
	if result.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// UpdatedAt
	if result.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	// CreatedBy
	if result.CreatedBy == nil || *result.CreatedBy != createdBy {
		t.Errorf("CreatedBy = %v, want %v", result.CreatedBy, &createdBy)
	}

	// UpdatedBy
	if result.UpdatedBy == nil || *result.UpdatedBy != updatedBy {
		t.Errorf("UpdatedBy = %v, want %v", result.UpdatedBy, &updatedBy)
	}
}

// TestFieldRepository_ToEntity_NullableFields はtoEntityメソッドがNULL値を正しくnilに変換することをテストする
func TestFieldRepository_ToEntity_NullableFields(t *testing.T) {
	now := time.Now()

	row := &sqlc.Field{
		ID:          uuid.New(),
		CityCode:    "163210",
		Name:        "テスト圃場",
		AreaSqm:     nil,
		H3IndexRes3: nil,
		H3IndexRes5: nil,
		H3IndexRes7: nil,
		H3IndexRes9: nil,
		SoilTypeID:  uuid.NullUUID{Valid: false},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		CreatedBy:   uuid.NullUUID{Valid: false},
		UpdatedBy:   uuid.NullUUID{Valid: false},
	}

	r := &fieldRepository{}
	result := r.toEntity(row)

	if result.AreaSqm != nil {
		t.Errorf("AreaSqm = %v, want nil", result.AreaSqm)
	}
	if result.H3IndexRes3 != nil {
		t.Errorf("H3IndexRes3 = %v, want nil", result.H3IndexRes3)
	}
	if result.H3IndexRes5 != nil {
		t.Errorf("H3IndexRes5 = %v, want nil", result.H3IndexRes5)
	}
	if result.H3IndexRes7 != nil {
		t.Errorf("H3IndexRes7 = %v, want nil", result.H3IndexRes7)
	}
	if result.H3IndexRes9 != nil {
		t.Errorf("H3IndexRes9 = %v, want nil", result.H3IndexRes9)
	}
	if result.SoilTypeID != nil {
		t.Errorf("SoilTypeID = %v, want nil", result.SoilTypeID)
	}
	if result.CreatedBy != nil {
		t.Errorf("CreatedBy = %v, want nil", result.CreatedBy)
	}
	if result.UpdatedBy != nil {
		t.Errorf("UpdatedBy = %v, want nil", result.UpdatedBy)
	}
}

// TestUuidToNullUUID はuuidToNullUUIDがnilと有効なUUIDを正しく変換することをテストする
func TestUuidToNullUUID(t *testing.T) {
	tests := []struct {
		name    string
		id      *uuid.UUID
		isValid bool
	}{
		{
			name:    "nil uuid",
			id:      nil,
			isValid: false,
		},
		{
			name:    "valid uuid",
			id:      func() *uuid.UUID { id := uuid.New(); return &id }(),
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uuidToNullUUID(tt.id)

			if result.Valid != tt.isValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.isValid)
			}

			if tt.isValid && result.UUID != *tt.id {
				t.Errorf("UUID = %v, want %v", result.UUID, *tt.id)
			}
		})
	}
}

// TestGeometryToWKB_Nil はgeometryToWKBがnilを受け取った場合にnilを返すことをテストする
func TestGeometryToWKB_Nil(t *testing.T) {
	result, err := geometryToWKB(nil)
	if err != nil {
		t.Errorf("geometryToWKB(nil) should not return error, got %v", err)
	}
	if result != nil {
		t.Error("geometryToWKB(nil) should return nil")
	}
}

// TestGeometryToWKB_NilPoint はgeometryToWKBがnil *geom.Pointを受け取った場合にnilを返すことをテストする
func TestGeometryToWKB_NilPoint(t *testing.T) {
	var p *geom.Point
	result, err := geometryToWKB(p)
	if err != nil {
		t.Errorf("geometryToWKB(nil *Point) should not return error, got %v", err)
	}
	if result != nil {
		t.Error("geometryToWKB(nil *Point) should return nil")
	}
}

// TestGeometryToWKB_NilPolygon はgeometryToWKBがnil *geom.Polygonを受け取った場合にnilを返すことをテストする
func TestGeometryToWKB_NilPolygon(t *testing.T) {
	var p *geom.Polygon
	result, err := geometryToWKB(p)
	if err != nil {
		t.Errorf("geometryToWKB(nil *Polygon) should not return error, got %v", err)
	}
	if result != nil {
		t.Error("geometryToWKB(nil *Polygon) should return nil")
	}
}

// TestGeometryToWKB_NilLineString はgeometryToWKBがnil *geom.LineStringを受け取った場合にnilを返すことをテストする
func TestGeometryToWKB_NilLineString(t *testing.T) {
	var l *geom.LineString
	result, err := geometryToWKB(l)
	if err != nil {
		t.Errorf("geometryToWKB(nil *LineString) should not return error, got %v", err)
	}
	if result != nil {
		t.Error("geometryToWKB(nil *LineString) should return nil")
	}
}

// TestGeometryToWKB_NilMultiPoint はgeometryToWKBがnil *geom.MultiPointを受け取った場合にnilを返すことをテストする
func TestGeometryToWKB_NilMultiPoint(t *testing.T) {
	var m *geom.MultiPoint
	result, err := geometryToWKB(m)
	if err != nil {
		t.Errorf("geometryToWKB(nil *MultiPoint) should not return error, got %v", err)
	}
	if result != nil {
		t.Error("geometryToWKB(nil *MultiPoint) should return nil")
	}
}

// TestGeometryToWKB_NilMultiPolygon はgeometryToWKBがnil *geom.MultiPolygonを受け取った場合にnilを返すことをテストする
func TestGeometryToWKB_NilMultiPolygon(t *testing.T) {
	var m *geom.MultiPolygon
	result, err := geometryToWKB(m)
	if err != nil {
		t.Errorf("geometryToWKB(nil *MultiPolygon) should not return error, got %v", err)
	}
	if result != nil {
		t.Error("geometryToWKB(nil *MultiPolygon) should return nil")
	}
}

// TestGeometryToWKB_NilMultiLineString はgeometryToWKBがnil *geom.MultiLineStringを受け取った場合にnilを返すことをテストする
func TestGeometryToWKB_NilMultiLineString(t *testing.T) {
	var m *geom.MultiLineString
	result, err := geometryToWKB(m)
	if err != nil {
		t.Errorf("geometryToWKB(nil *MultiLineString) should not return error, got %v", err)
	}
	if result != nil {
		t.Error("geometryToWKB(nil *MultiLineString) should return nil")
	}
}

// TestGeometryToWKB_ValidPolygon はgeometryToWKBが有効なPolygonをWKBバイト列に変換することをテストする
func TestGeometryToWKB_ValidPolygon(t *testing.T) {
	polygon := geom.NewPolygon(geom.XY)
	coords := [][]geom.Coord{
		{
			{139.6917, 35.6895},
			{139.6920, 35.6895},
			{139.6920, 35.6898},
			{139.6917, 35.6898},
			{139.6917, 35.6895},
		},
	}
	if _, err := polygon.SetCoords(coords); err != nil {
		t.Fatalf("Failed to set coords: %v", err)
	}

	result, err := geometryToWKB(polygon)
	if err != nil {
		t.Errorf("geometryToWKB(polygon) should not return error, got %v", err)
	}
	if result == nil {
		t.Error("geometryToWKB(polygon) should return non-nil bytes")
	}
	if len(result) == 0 {
		t.Error("geometryToWKB(polygon) should return non-empty bytes")
	}
}

// TestGeometryToWKB_ValidPoint はgeometryToWKBが有効なPointをWKBバイト列に変換することをテストする
func TestGeometryToWKB_ValidPoint(t *testing.T) {
	point := geom.NewPoint(geom.XY)
	if _, err := point.SetCoords(geom.Coord{139.6917, 35.6895}); err != nil {
		t.Fatalf("Failed to set coords: %v", err)
	}

	result, err := geometryToWKB(point)
	if err != nil {
		t.Errorf("geometryToWKB(point) should not return error, got %v", err)
	}
	if result == nil {
		t.Error("geometryToWKB(point) should return non-nil bytes")
	}
}
