package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/repository"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
)

// masterRepository はMasterRepositoryの実装
type masterRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewMasterRepository は新しいMasterRepositoryを作成する
func NewMasterRepository(db *pgxpool.Pool) repository.MasterRepository {
	return &masterRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// UpsertSoilType は土壌タイプをUPSERTする
func (r *masterRepository) UpsertSoilType(ctx context.Context, soilType *entity.SoilType) (*uuid.UUID, error) {
	row, err := r.queries.UpsertSoilType(ctx, &sqlc.UpsertSoilTypeParams{
		LargeCode:  soilType.LargeCode,
		MiddleCode: soilType.MiddleCode,
		SmallCode:  soilType.SmallCode,
		SmallName:  soilType.SmallName,
	})
	if err != nil {
		return nil, err
	}
	return &row.ID, nil
}

// UpsertLandCategory は土地種別をUPSERTする
func (r *masterRepository) UpsertLandCategory(ctx context.Context, category *entity.LandCategory) error {
	_, err := r.queries.UpsertLandCategory(ctx, &sqlc.UpsertLandCategoryParams{
		Code: category.Code,
		Name: category.Name,
	})
	return err
}

// UpsertIdleLandStatus は遊休農地状況をUPSERTする
func (r *masterRepository) UpsertIdleLandStatus(ctx context.Context, status *entity.IdleLandStatus) error {
	_, err := r.queries.UpsertIdleLandStatus(ctx, &sqlc.UpsertIdleLandStatusParams{
		Code: status.Code,
		Name: status.Name,
	})
	return err
}

// FindSoilTypeBySmallCode は小分類コードで土壌タイプを取得する
func (r *masterRepository) FindSoilTypeBySmallCode(ctx context.Context, smallCode string) (*entity.SoilType, error) {
	row, err := r.queries.GetSoilTypeBySmallCode(ctx, smallCode)
	if err != nil {
		return nil, err
	}
	return &entity.SoilType{
		ID:         row.ID,
		LargeCode:  row.LargeCode,
		MiddleCode: row.MiddleCode,
		SmallCode:  row.SmallCode,
		SmallName:  row.SmallName,
	}, nil
}

// FindLandCategoryByCode はコードで土地種別を取得する
func (r *masterRepository) FindLandCategoryByCode(ctx context.Context, code string) (*entity.LandCategory, error) {
	row, err := r.queries.GetLandCategory(ctx, code)
	if err != nil {
		return nil, err
	}
	return &entity.LandCategory{
		Code:        row.Code,
		Name:        row.Name,
		Description: row.Description,
	}, nil
}

// FindIdleLandStatusByCode はコードで遊休農地状況を取得する
func (r *masterRepository) FindIdleLandStatusByCode(ctx context.Context, code string) (*entity.IdleLandStatus, error) {
	row, err := r.queries.GetIdleLandStatus(ctx, code)
	if err != nil {
		return nil, err
	}
	return &entity.IdleLandStatus{
		Code:        row.Code,
		Name:        row.Name,
		Description: row.Description,
	}, nil
}

// fieldLandRegistryRepository はFieldLandRegistryRepositoryの実装
type fieldLandRegistryRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewFieldLandRegistryRepository は新しいFieldLandRegistryRepositoryを作成する
func NewFieldLandRegistryRepository(db *pgxpool.Pool) repository.FieldLandRegistryRepository {
	return &fieldLandRegistryRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// FindByFieldID は圃場IDで農地台帳を取得する
func (r *fieldLandRegistryRepository) FindByFieldID(ctx context.Context, fieldID uuid.UUID) ([]*entity.FieldLandRegistry, error) {
	rows, err := r.queries.ListFieldLandRegistriesByFieldID(ctx, fieldID)
	if err != nil {
		return nil, err
	}

	registries := make([]*entity.FieldLandRegistry, len(rows))
	for i, row := range rows {
		registries[i] = r.toEntity(row)
	}
	return registries, nil
}

// Create は農地台帳を作成する
func (r *fieldLandRegistryRepository) Create(ctx context.Context, registry *entity.FieldLandRegistry) error {
	var descriptiveStudyData pgtype.Date
	if registry.DescriptiveStudyData != nil {
		descriptiveStudyData = pgtype.Date{
			Time:  *registry.DescriptiveStudyData,
			Valid: true,
		}
	}

	_, err := r.queries.CreateFieldLandRegistry(ctx, &sqlc.CreateFieldLandRegistryParams{
		FieldID:              registry.FieldID,
		FarmerNumber:         registry.FarmerNumber,
		Address:              registry.Address,
		AreaSqm:              registry.AreaSqm,
		LandCategoryCode:     registry.LandCategoryCode,
		IdleLandStatusCode:   registry.IdleLandStatusCode,
		DescriptiveStudyData: descriptiveStudyData,
	})
	return err
}

// DeleteByFieldID は圃場IDで農地台帳を削除する
func (r *fieldLandRegistryRepository) DeleteByFieldID(ctx context.Context, fieldID uuid.UUID) error {
	return r.queries.DeleteFieldLandRegistriesByFieldID(ctx, fieldID)
}

// DeleteByFieldIDs は複数の圃場IDで農地台帳を一括削除する
func (r *fieldLandRegistryRepository) DeleteByFieldIDs(ctx context.Context, fieldIDs []uuid.UUID) error {
	return r.queries.DeleteFieldLandRegistriesByFieldIDs(ctx, fieldIDs)
}

// CreateBatch は農地台帳をバッチで作成する
func (r *fieldLandRegistryRepository) CreateBatch(ctx context.Context, registries []*entity.FieldLandRegistry) error {
	for _, registry := range registries {
		if err := r.Create(ctx, registry); err != nil {
			return err
		}
	}
	return nil
}

// toEntity はSQLCモデルをエンティティに変換する
func (r *fieldLandRegistryRepository) toEntity(row *sqlc.FieldLandRegistry) *entity.FieldLandRegistry {
	if row == nil {
		return nil
	}

	registry := &entity.FieldLandRegistry{
		ID:                 row.ID,
		FieldID:            row.FieldID,
		FarmerNumber:       row.FarmerNumber,
		Address:            row.Address,
		AreaSqm:            row.AreaSqm,
		LandCategoryCode:   row.LandCategoryCode,
		IdleLandStatusCode: row.IdleLandStatusCode,
	}

	if row.DescriptiveStudyData.Valid {
		t := row.DescriptiveStudyData.Time
		registry.DescriptiveStudyData = &t
	}
	if row.CreatedAt.Valid {
		registry.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		registry.UpdatedAt = row.UpdatedAt.Time
	}

	return registry
}
