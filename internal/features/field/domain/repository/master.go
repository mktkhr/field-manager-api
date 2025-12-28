package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
)

// MasterRepository はマスタデータのリポジトリインターフェース
type MasterRepository interface {
	// UpsertSoilType は土壌タイプをUPSERTする
	UpsertSoilType(ctx context.Context, soilType *entity.SoilType) (*uuid.UUID, error)

	// UpsertLandCategory は土地種別をUPSERTする
	UpsertLandCategory(ctx context.Context, category *entity.LandCategory) error

	// UpsertIdleLandStatus は遊休農地状況をUPSERTする
	UpsertIdleLandStatus(ctx context.Context, status *entity.IdleLandStatus) error

	// FindSoilTypeBySmallCode は小分類コードで土壌タイプを取得する
	FindSoilTypeBySmallCode(ctx context.Context, smallCode string) (*entity.SoilType, error)

	// FindLandCategoryByCode はコードで土地種別を取得する
	FindLandCategoryByCode(ctx context.Context, code string) (*entity.LandCategory, error)

	// FindIdleLandStatusByCode はコードで遊休農地状況を取得する
	FindIdleLandStatusByCode(ctx context.Context, code string) (*entity.IdleLandStatus, error)
}

// FieldLandRegistryRepository は農地台帳のリポジトリインターフェース
type FieldLandRegistryRepository interface {
	// FindByFieldID は圃場IDで農地台帳を取得する
	FindByFieldID(ctx context.Context, fieldID uuid.UUID) ([]*entity.FieldLandRegistry, error)

	// Create は農地台帳を作成する
	Create(ctx context.Context, registry *entity.FieldLandRegistry) error

	// DeleteByFieldID は圃場IDで農地台帳を削除する
	DeleteByFieldID(ctx context.Context, fieldID uuid.UUID) error

	// DeleteByFieldIDs は複数の圃場IDで農地台帳を一括削除する
	DeleteByFieldIDs(ctx context.Context, fieldIDs []uuid.UUID) error

	// CreateBatch は農地台帳をバッチで作成する
	CreateBatch(ctx context.Context, registries []*entity.FieldLandRegistry) error
}
