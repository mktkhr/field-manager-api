package entity

import (
	"time"

	"github.com/google/uuid"
)

// SoilType は土壌タイプエンティティ
type SoilType struct {
	ID          uuid.UUID
	LargeCode   string
	MiddleCode  string
	SmallCode   string
	SmallName   string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewSoilType は新しいSoilTypeを作成する
func NewSoilType(largeCode, middleCode, smallCode, smallName string) *SoilType {
	now := time.Now()
	return &SoilType{
		ID:         uuid.New(),
		LargeCode:  largeCode,
		MiddleCode: middleCode,
		SmallCode:  smallCode,
		SmallName:  smallName,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
