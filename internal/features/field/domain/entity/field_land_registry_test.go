package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNewFieldLandRegistry はNewFieldLandRegistryがFieldIDとタイムスタンプを正しく設定したFieldLandRegistryを生成することをテストする
func TestNewFieldLandRegistry(t *testing.T) {
	fieldID := uuid.New()

	registry := NewFieldLandRegistry(fieldID)

	if registry.FieldID != fieldID {
		t.Errorf("FieldID = %v, want %v", registry.FieldID, fieldID)
	}
	if registry.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if registry.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

// TestFieldLandRegistrySetters は各Setterメソッドが有効な値を正しく設定することをテストする
func TestFieldLandRegistrySetters(t *testing.T) {
	registry := &FieldLandRegistry{}

	farmerNumber := "12345"
	registry.SetFarmerNumber(farmerNumber)
	if registry.FarmerNumber == nil || *registry.FarmerNumber != farmerNumber {
		t.Errorf("FarmerNumber = %v, want %q", registry.FarmerNumber, farmerNumber)
	}

	address := "東京都千代田区"
	registry.SetAddress(address)
	if registry.Address == nil || *registry.Address != address {
		t.Errorf("Address = %v, want %q", registry.Address, address)
	}

	areaSqm := 1000
	registry.SetAreaSqm(areaSqm)
	if registry.AreaSqm == nil || int(*registry.AreaSqm) != areaSqm {
		t.Errorf("AreaSqm = %v, want %d", registry.AreaSqm, areaSqm)
	}

	landCategoryCode := "01"
	registry.SetLandCategoryCode(landCategoryCode)
	if registry.LandCategoryCode == nil || *registry.LandCategoryCode != landCategoryCode {
		t.Errorf("LandCategoryCode = %v, want %q", registry.LandCategoryCode, landCategoryCode)
	}

	idleLandStatusCode := "1"
	registry.SetIdleLandStatusCode(idleLandStatusCode)
	if registry.IdleLandStatusCode == nil || *registry.IdleLandStatusCode != idleLandStatusCode {
		t.Errorf("IdleLandStatusCode = %v, want %q", registry.IdleLandStatusCode, idleLandStatusCode)
	}

	now := time.Now()
	registry.SetDescriptiveStudyData(&now)
	if registry.DescriptiveStudyData == nil || !registry.DescriptiveStudyData.Equal(now) {
		t.Errorf("DescriptiveStudyData = %v, want %v", registry.DescriptiveStudyData, now)
	}
}

// TestFieldLandRegistrySettersWithEmpty は各Setterメソッドが空値やゼロ値の場合にnilを設定することをテストする
func TestFieldLandRegistrySettersWithEmpty(t *testing.T) {
	registry := &FieldLandRegistry{}

	registry.SetFarmerNumber("")
	if registry.FarmerNumber != nil {
		t.Error("FarmerNumber should be nil for empty string")
	}

	registry.SetAddress("")
	if registry.Address != nil {
		t.Error("Address should be nil for empty string")
	}

	registry.SetAreaSqm(0)
	if registry.AreaSqm != nil {
		t.Error("AreaSqm should be nil for zero value")
	}

	registry.SetLandCategoryCode("")
	if registry.LandCategoryCode != nil {
		t.Error("LandCategoryCode should be nil for empty string")
	}

	registry.SetIdleLandStatusCode("")
	if registry.IdleLandStatusCode != nil {
		t.Error("IdleLandStatusCode should be nil for empty string")
	}

	registry.SetDescriptiveStudyData(nil)
	if registry.DescriptiveStudyData != nil {
		t.Error("DescriptiveStudyData should be nil")
	}
}
