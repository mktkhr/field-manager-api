package entity

import (
	"time"

	"github.com/google/uuid"
)

// FieldLandRegistry は農地台帳エンティティ
type FieldLandRegistry struct {
	ID                   uuid.UUID
	FieldID              uuid.UUID
	FarmerNumber         *string
	Address              *string
	AreaSqm              *int32
	LandCategoryCode     *string
	IdleLandStatusCode   *string
	DescriptiveStudyData *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewFieldLandRegistry は新しいFieldLandRegistryを作成する
func NewFieldLandRegistry(fieldID uuid.UUID) *FieldLandRegistry {
	now := time.Now()
	return &FieldLandRegistry{
		ID:        uuid.New(),
		FieldID:   fieldID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetFarmerNumber は耕作者番号を設定する
func (r *FieldLandRegistry) SetFarmerNumber(farmerNumber string) {
	if farmerNumber != "" {
		r.FarmerNumber = &farmerNumber
	}
}

// SetAddress は住所を設定する
func (r *FieldLandRegistry) SetAddress(address string) {
	if address != "" {
		r.Address = &address
	}
}

// SetAreaSqm は面積を設定する
func (r *FieldLandRegistry) SetAreaSqm(area int) {
	if area != 0 {
		a := int32(area)
		r.AreaSqm = &a
	}
}

// SetLandCategoryCode は土地種別コードを設定する
func (r *FieldLandRegistry) SetLandCategoryCode(code string) {
	if code != "" {
		r.LandCategoryCode = &code
	}
}

// SetIdleLandStatusCode は遊休農地状況コードを設定する
func (r *FieldLandRegistry) SetIdleLandStatusCode(code string) {
	if code != "" {
		r.IdleLandStatusCode = &code
	}
}

// SetDescriptiveStudyData は実態調査日を設定する
func (r *FieldLandRegistry) SetDescriptiveStudyData(date *time.Time) {
	r.DescriptiveStudyData = date
}
