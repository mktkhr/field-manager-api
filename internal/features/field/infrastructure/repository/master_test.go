package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
)

// TestFieldLandRegistryRepository_ToEntity はtoEntityメソッドがsqlc.FieldLandRegistryをEntity.FieldLandRegistryに正しく変換することをテストする
func TestFieldLandRegistryRepository_ToEntity(t *testing.T) {
	now := time.Now()
	farmerNumber := "12345678"
	address := "東京都千代田区1-2-3"
	areaSqm := int32(1000)
	landCategoryCode := "001"
	idleLandStatusCode := "A"
	descriptiveStudyDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		row  *sqlc.FieldLandRegistry
		want *entity.FieldLandRegistry
	}{
		{
			name: "nil row",
			row:  nil,
			want: nil,
		},
		{
			name: "minimal registry",
			row: &sqlc.FieldLandRegistry{
				ID:        uuid.New(),
				FieldID:   uuid.New(),
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.FieldLandRegistry{
				FarmerNumber:         nil,
				Address:              nil,
				AreaSqm:              nil,
				LandCategoryCode:     nil,
				IdleLandStatusCode:   nil,
				DescriptiveStudyData: nil,
			},
		},
		{
			name: "registry with farmer info",
			row: &sqlc.FieldLandRegistry{
				ID:           uuid.New(),
				FieldID:      uuid.New(),
				FarmerNumber: &farmerNumber,
				Address:      &address,
				CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.FieldLandRegistry{
				FarmerNumber: &farmerNumber,
				Address:      &address,
			},
		},
		{
			name: "registry with area and category",
			row: &sqlc.FieldLandRegistry{
				ID:               uuid.New(),
				FieldID:          uuid.New(),
				AreaSqm:          &areaSqm,
				LandCategoryCode: &landCategoryCode,
				CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.FieldLandRegistry{
				AreaSqm:          &areaSqm,
				LandCategoryCode: &landCategoryCode,
			},
		},
		{
			name: "full registry with all fields",
			row: &sqlc.FieldLandRegistry{
				ID:                   uuid.New(),
				FieldID:              uuid.New(),
				FarmerNumber:         &farmerNumber,
				Address:              &address,
				AreaSqm:              &areaSqm,
				LandCategoryCode:     &landCategoryCode,
				IdleLandStatusCode:   &idleLandStatusCode,
				DescriptiveStudyData: pgtype.Date{Time: descriptiveStudyDate, Valid: true},
				CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt:            pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
			},
			want: &entity.FieldLandRegistry{
				FarmerNumber:         &farmerNumber,
				Address:              &address,
				AreaSqm:              &areaSqm,
				LandCategoryCode:     &landCategoryCode,
				IdleLandStatusCode:   &idleLandStatusCode,
				DescriptiveStudyData: &descriptiveStudyDate,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &fieldLandRegistryRepository{}
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

			// FarmerNumber
			if (result.FarmerNumber == nil) != (tt.want.FarmerNumber == nil) {
				t.Errorf("FarmerNumber nil mismatch: got %v, want %v", result.FarmerNumber, tt.want.FarmerNumber)
			}
			if result.FarmerNumber != nil && tt.want.FarmerNumber != nil && *result.FarmerNumber != *tt.want.FarmerNumber {
				t.Errorf("FarmerNumber = %q, want %q", *result.FarmerNumber, *tt.want.FarmerNumber)
			}

			// Address
			if (result.Address == nil) != (tt.want.Address == nil) {
				t.Errorf("Address nil mismatch: got %v, want %v", result.Address, tt.want.Address)
			}
			if result.Address != nil && tt.want.Address != nil && *result.Address != *tt.want.Address {
				t.Errorf("Address = %q, want %q", *result.Address, *tt.want.Address)
			}
		})
	}
}

// TestFieldLandRegistryRepository_ToEntity_FieldMapping はtoEntityメソッドが全フィールドを正しくマッピングすることをテストする
func TestFieldLandRegistryRepository_ToEntity_FieldMapping(t *testing.T) {
	now := time.Now()
	updatedAt := now.Add(time.Hour)
	id := uuid.New()
	fieldID := uuid.New()
	farmerNumber := "12345678"
	address := "東京都千代田区1-2-3"
	areaSqm := int32(1000)
	landCategoryCode := "001"
	idleLandStatusCode := "A"
	descriptiveStudyDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	row := &sqlc.FieldLandRegistry{
		ID:                   id,
		FieldID:              fieldID,
		FarmerNumber:         &farmerNumber,
		Address:              &address,
		AreaSqm:              &areaSqm,
		LandCategoryCode:     &landCategoryCode,
		IdleLandStatusCode:   &idleLandStatusCode,
		DescriptiveStudyData: pgtype.Date{Time: descriptiveStudyDate, Valid: true},
		CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:            pgtype.Timestamptz{Time: updatedAt, Valid: true},
	}

	r := &fieldLandRegistryRepository{}
	result := r.toEntity(row)

	// ID
	if result.ID != id {
		t.Errorf("ID = %v, want %v", result.ID, id)
	}

	// FieldID
	if result.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", result.FieldID, fieldID)
	}

	// FarmerNumber
	if result.FarmerNumber == nil || *result.FarmerNumber != farmerNumber {
		t.Errorf("FarmerNumber = %v, want %v", result.FarmerNumber, &farmerNumber)
	}

	// Address
	if result.Address == nil || *result.Address != address {
		t.Errorf("Address = %v, want %v", result.Address, &address)
	}

	// AreaSqm
	if result.AreaSqm == nil || *result.AreaSqm != areaSqm {
		t.Errorf("AreaSqm = %v, want %v", result.AreaSqm, &areaSqm)
	}

	// LandCategoryCode
	if result.LandCategoryCode == nil || *result.LandCategoryCode != landCategoryCode {
		t.Errorf("LandCategoryCode = %v, want %v", result.LandCategoryCode, &landCategoryCode)
	}

	// IdleLandStatusCode
	if result.IdleLandStatusCode == nil || *result.IdleLandStatusCode != idleLandStatusCode {
		t.Errorf("IdleLandStatusCode = %v, want %v", result.IdleLandStatusCode, &idleLandStatusCode)
	}

	// DescriptiveStudyData
	if result.DescriptiveStudyData == nil {
		t.Error("DescriptiveStudyData should not be nil")
	} else if !result.DescriptiveStudyData.Equal(descriptiveStudyDate) {
		t.Errorf("DescriptiveStudyData = %v, want %v", result.DescriptiveStudyData, descriptiveStudyDate)
	}

	// CreatedAt
	if result.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// UpdatedAt
	if result.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

// TestFieldLandRegistryRepository_ToEntity_NullableFields はtoEntityメソッドがNULL値を正しくnilに変換することをテストする
func TestFieldLandRegistryRepository_ToEntity_NullableFields(t *testing.T) {
	now := time.Now()

	row := &sqlc.FieldLandRegistry{
		ID:                   uuid.New(),
		FieldID:              uuid.New(),
		FarmerNumber:         nil,
		Address:              nil,
		AreaSqm:              nil,
		LandCategoryCode:     nil,
		IdleLandStatusCode:   nil,
		DescriptiveStudyData: pgtype.Date{Valid: false},
		CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
	}

	r := &fieldLandRegistryRepository{}
	result := r.toEntity(row)

	if result.FarmerNumber != nil {
		t.Errorf("FarmerNumber = %v, want nil", result.FarmerNumber)
	}
	if result.Address != nil {
		t.Errorf("Address = %v, want nil", result.Address)
	}
	if result.AreaSqm != nil {
		t.Errorf("AreaSqm = %v, want nil", result.AreaSqm)
	}
	if result.LandCategoryCode != nil {
		t.Errorf("LandCategoryCode = %v, want nil", result.LandCategoryCode)
	}
	if result.IdleLandStatusCode != nil {
		t.Errorf("IdleLandStatusCode = %v, want nil", result.IdleLandStatusCode)
	}
	if result.DescriptiveStudyData != nil {
		t.Errorf("DescriptiveStudyData = %v, want nil", result.DescriptiveStudyData)
	}
}

// TestFieldLandRegistryRepository_ToEntity_InvalidTimestamps はtoEntityメソッドが無効なタイムスタンプをゼロ値に変換することをテストする
func TestFieldLandRegistryRepository_ToEntity_InvalidTimestamps(t *testing.T) {
	row := &sqlc.FieldLandRegistry{
		ID:        uuid.New(),
		FieldID:   uuid.New(),
		CreatedAt: pgtype.Timestamptz{Valid: false},
		UpdatedAt: pgtype.Timestamptz{Valid: false},
	}

	r := &fieldLandRegistryRepository{}
	result := r.toEntity(row)

	if !result.CreatedAt.IsZero() {
		t.Errorf("CreatedAt = %v, want zero time", result.CreatedAt)
	}
	if !result.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt = %v, want zero time", result.UpdatedAt)
	}
}
