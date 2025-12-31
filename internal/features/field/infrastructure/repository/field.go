package repository

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/repository"
	"github.com/mktkhr/field-manager-api/internal/features/shared/types"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
)

// fieldRepository はFieldRepositoryの実装
type fieldRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewFieldRepository は新しいFieldRepositoryを作成する
func NewFieldRepository(db *pgxpool.Pool) repository.FieldRepository {
	return &fieldRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// FindByID はIDで圃場を取得する
func (r *fieldRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Field, error) {
	row, err := r.queries.GetField(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toEntity(row), nil
}

// Create は圃場を作成する
func (r *fieldRepository) Create(ctx context.Context, field *entity.Field) error {
	// 実装は省略(基本的なCRUDはSQLCで対応)
	return nil
}

// Update は圃場を更新する
func (r *fieldRepository) Update(ctx context.Context, field *entity.Field) error {
	// 実装は省略(基本的なCRUDはSQLCで対応)
	return nil
}

// Delete は圃場を削除する
func (r *fieldRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteField(ctx, id)
}

// Upsert は圃場をUPSERTする
func (r *fieldRepository) Upsert(ctx context.Context, field *entity.Field) error {
	// 実装は省略
	return nil
}

// UpsertBatch は圃場をバッチでUPSERTする(wagriインポート用)
func (r *fieldRepository) UpsertBatch(ctx context.Context, inputs []types.FieldBatchInput) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := sqlc.New(tx)

	for _, input := range inputs {
		// 1. 土壌タイプをUPSERT(トランザクション内で直接実行)
		var soilTypeID *uuid.UUID
		if input.HasSoilType() {
			row, err := queries.UpsertSoilType(ctx, &sqlc.UpsertSoilTypeParams{
				LargeCode:  input.SoilType.LargeCode,
				MiddleCode: input.SoilType.MiddleCode,
				SmallCode:  input.SoilType.SmallCode,
				SmallName:  input.SoilType.SmallName,
			})
			if err != nil {
				return fmt.Errorf("土壌タイプUPSERT失敗: %w", err)
			}
			soilTypeID = &row.ID
		}

		// 2. PinInfoからマスタデータをUPSERT(トランザクション内で直接実行)
		for _, pinInfo := range input.PinInfoList {
			if pinInfo.LandCategoryCode != "" {
				if _, err := queries.UpsertLandCategory(ctx, &sqlc.UpsertLandCategoryParams{
					Code: pinInfo.LandCategoryCode,
					Name: pinInfo.LandCategory,
				}); err != nil {
					return fmt.Errorf("土地種別UPSERT失敗: %w", err)
				}
			}
			if pinInfo.IdleLandStatusCode != "" {
				if _, err := queries.UpsertIdleLandStatus(ctx, &sqlc.UpsertIdleLandStatusParams{
					Code: pinInfo.IdleLandStatusCode,
					Name: pinInfo.IdleLandStatus,
				}); err != nil {
					return fmt.Errorf("遊休農地状況UPSERT失敗: %w", err)
				}
			}
		}

		// 3. Field作成
		fieldID, err := uuid.Parse(input.ID)
		if err != nil {
			return fmt.Errorf("圃場ID変換失敗: %w", err)
		}

		// LinearPolygon -> Polygon変換
		geometryCoords := input.GetFirstCoordinates()
		polygon, err := entity.ConvertLinearPolygonToPolygon(geometryCoords)
		if err != nil {
			return fmt.Errorf("ジオメトリ変換失敗: %w", err)
		}

		field := entity.NewField(fieldID, input.CityCode)
		field.SetGeometry(polygon)
		if soilTypeID != nil {
			field.SetSoilType(*soilTypeID)
		}

		// GeometryをWKB形式に変換
		geometryWKB, err := geometryToWKB(field.Geometry)
		if err != nil {
			return fmt.Errorf("geometry WKB変換失敗: %w", err)
		}
		centroidWKB, err := geometryToWKB(field.Centroid)
		if err != nil {
			return fmt.Errorf("centroid WKB変換失敗: %w", err)
		}

		// UpsertFieldを実行
		_, err = queries.UpsertField(ctx, &sqlc.UpsertFieldParams{
			ID:          field.ID,
			GeometryWkb: geometryWKB,
			CentroidWkb: centroidWKB,
			H3IndexRes3: field.H3IndexRes3,
			H3IndexRes5: field.H3IndexRes5,
			H3IndexRes7: field.H3IndexRes7,
			H3IndexRes9: field.H3IndexRes9,
			CityCode:    field.CityCode,
			SoilTypeID:  uuidToNullUUID(field.SoilTypeID),
		})
		if err != nil {
			return fmt.Errorf("圃場UPSERT失敗: %w", err)
		}

		// 4. 農地台帳をREPLACE(tx内で直接実行)
		if err := queries.DeleteFieldLandRegistriesByFieldID(ctx, fieldID); err != nil {
			return fmt.Errorf("農地台帳削除失敗: %w", err)
		}

		for _, pinInfo := range input.PinInfoList {
			registry := entity.NewFieldLandRegistry(fieldID)
			registry.SetFarmerNumber(pinInfo.FarmerNumber)
			registry.SetAddress(pinInfo.Address)
			registry.SetAreaSqm(pinInfo.Area)
			registry.SetLandCategoryCode(pinInfo.LandCategoryCode)
			registry.SetIdleLandStatusCode(pinInfo.IdleLandStatusCode)
			registry.SetDescriptiveStudyData(pinInfo.ParseDescriptiveStudyData())

			var descriptiveStudyData pgtype.Date
			if registry.DescriptiveStudyData != nil {
				descriptiveStudyData = pgtype.Date{
					Time:  *registry.DescriptiveStudyData,
					Valid: true,
				}
			}

			if _, err := queries.CreateFieldLandRegistry(ctx, &sqlc.CreateFieldLandRegistryParams{
				FieldID:              registry.FieldID,
				FarmerNumber:         registry.FarmerNumber,
				Address:              registry.Address,
				AreaSqm:              registry.AreaSqm,
				LandCategoryCode:     registry.LandCategoryCode,
				IdleLandStatusCode:   registry.IdleLandStatusCode,
				DescriptiveStudyData: descriptiveStudyData,
			}); err != nil {
				return fmt.Errorf("農地台帳作成失敗: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("コミット失敗: %w", err)
	}

	return nil
}

// toEntity はSQLCモデルをエンティティに変換する
func (r *fieldRepository) toEntity(row *sqlc.Field) *entity.Field {
	if row == nil {
		return nil
	}

	field := &entity.Field{
		ID:          row.ID,
		AreaSqm:     row.AreaSqm,
		H3IndexRes3: row.H3IndexRes3,
		H3IndexRes5: row.H3IndexRes5,
		H3IndexRes7: row.H3IndexRes7,
		H3IndexRes9: row.H3IndexRes9,
		CityCode:    row.CityCode,
		Name:        row.Name,
	}

	if row.SoilTypeID.Valid {
		field.SoilTypeID = &row.SoilTypeID.UUID
	}
	if row.CreatedAt.Valid {
		field.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		field.UpdatedAt = row.UpdatedAt.Time
	}
	if row.CreatedBy.Valid {
		field.CreatedBy = &row.CreatedBy.UUID
	}
	if row.UpdatedBy.Valid {
		field.UpdatedBy = &row.UpdatedBy.UUID
	}

	return field
}

// uuidToNullUUID はuuid.UUIDポインタをuuid.NullUUIDに変換する
func uuidToNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

// geometryToWKB はgeom.TをWKB形式のバイト列に変換する
func geometryToWKB(g geom.T) ([]byte, error) {
	if g == nil {
		return nil, nil
	}

	// nilポインタチェック(interface{}としてnilでない場合も考慮)
	switch v := g.(type) {
	case *geom.Point:
		if v == nil {
			return nil, nil
		}
	case *geom.Polygon:
		if v == nil {
			return nil, nil
		}
	case *geom.LineString:
		if v == nil {
			return nil, nil
		}
	case *geom.MultiPoint:
		if v == nil {
			return nil, nil
		}
	case *geom.MultiPolygon:
		if v == nil {
			return nil, nil
		}
	case *geom.MultiLineString:
		if v == nil {
			return nil, nil
		}
	}

	return wkb.Marshal(g, binary.LittleEndian)
}
