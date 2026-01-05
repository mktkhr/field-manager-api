// Package query は圃場機能のクエリ実装を提供する
package query

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	appQuery "github.com/mktkhr/field-manager-api/internal/features/field/application/query"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
)

// fieldQuery はFieldQueryの実装
type fieldQuery struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewFieldQuery は新しいFieldQueryを作成する
func NewFieldQuery(db *pgxpool.Pool) appQuery.FieldQuery {
	return &fieldQuery{
		db:      db,
		queries: sqlc.New(db),
	}
}

// List は圃場一覧を取得する
func (q *fieldQuery) List(ctx context.Context, limit, offset int32) ([]*entity.Field, error) {
	rows, err := q.queries.ListFields(ctx, &sqlc.ListFieldsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("圃場一覧の取得に失敗: %w", err)
	}

	fields := make([]*entity.Field, len(rows))
	for i, row := range rows {
		field, err := q.toEntity(row)
		if err != nil {
			return nil, fmt.Errorf("圃場エンティティ変換に失敗: %w", err)
		}
		fields[i] = field
	}
	return fields, nil
}

// Count は圃場の総数を取得する
func (q *fieldQuery) Count(ctx context.Context) (int64, error) {
	count, err := q.queries.CountFields(ctx)
	if err != nil {
		return 0, fmt.Errorf("圃場総数の取得に失敗: %w", err)
	}
	return count, nil
}

// toEntity はSQLCモデルをエンティティに変換する
func (q *fieldQuery) toEntity(row *sqlc.Field) (*entity.Field, error) {
	if row == nil {
		return nil, nil
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

	// Geometry変換(interface{} -> *geom.Polygon)
	if row.Geometry != nil {
		geometry, err := q.parseGeometry(row.Geometry)
		if err != nil {
			return nil, fmt.Errorf("geometry解析に失敗: %w", err)
		}
		if polygon, ok := geometry.(*geom.Polygon); ok {
			field.Geometry = polygon
		}
	}

	// Centroid変換(interface{} -> *geom.Point)
	if row.Centroid != nil {
		centroid, err := q.parseGeometry(row.Centroid)
		if err != nil {
			return nil, fmt.Errorf("centroid解析に失敗: %w", err)
		}
		if point, ok := centroid.(*geom.Point); ok {
			field.Centroid = point
		}
	}

	return field, nil
}

// parseGeometry はWKBバイト列をgeom.Tに変換する
func (q *fieldQuery) parseGeometry(data interface{}) (geom.T, error) {
	if data == nil {
		return nil, nil
	}

	var bytes []byte
	switch v := data.(type) {
	case []byte:
		bytes = v
	default:
		return nil, fmt.Errorf("不正なgeometry型: %T", data)
	}

	if len(bytes) == 0 {
		return nil, nil
	}

	g, err := wkb.Unmarshal(bytes)
	if err != nil {
		return nil, fmt.Errorf("WKBデコードに失敗: %w", err)
	}
	return g, nil
}
