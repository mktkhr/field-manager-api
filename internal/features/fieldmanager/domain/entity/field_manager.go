// Package entity は圃場管理者機能のエンティティを提供する
package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ManagerType は管理者タイプを表す
type ManagerType string

const (
	// ManagerTypeOrganization は組織タイプ
	ManagerTypeOrganization ManagerType = "organization"
	// ManagerTypeUser はユーザータイプ
	ManagerTypeUser ManagerType = "user"
	// ManagerTypeRegion は地域タイプ
	ManagerTypeRegion ManagerType = "region"
)

// IsValid はManagerTypeが有効な値かチェックする
func (m ManagerType) IsValid() bool {
	switch m {
	case ManagerTypeOrganization, ManagerTypeUser, ManagerTypeRegion:
		return true
	}
	return false
}

// ParseManagerType は文字列からManagerTypeを解析する
func ParseManagerType(s string) (ManagerType, error) {
	mt := ManagerType(s)
	if !mt.IsValid() {
		return "", fmt.Errorf("無効な管理者タイプです: %s(有効値: organization, user, region)", s)
	}
	return mt, nil
}

// FieldManager は圃場管理関係エンティティ
type FieldManager struct {
	ID          uuid.UUID
	FieldID     uuid.UUID
	ManagerType ManagerType
	ManagerID   uuid.UUID
	CreatedAt   time.Time
	CreatedBy   *uuid.UUID
}

// NewFieldManager は新しいFieldManagerを作成する
func NewFieldManager(fieldID uuid.UUID, managerType ManagerType, managerID uuid.UUID, createdBy *uuid.UUID) (*FieldManager, error) {
	if !managerType.IsValid() {
		return nil, fmt.Errorf("無効な管理者タイプです: %s", managerType)
	}

	return &FieldManager{
		ID:          uuid.New(),
		FieldID:     fieldID,
		ManagerType: managerType,
		ManagerID:   managerID,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
	}, nil
}
