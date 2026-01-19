// Package query は圃場検索機能のクエリ実装を提供する
package query

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/application/query"
	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/domain/entity"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
)

// fieldSearchQuery はFieldSearchQueryの実装
type fieldSearchQuery struct {
	db *pgxpool.Pool
}

// NewFieldSearchQuery は新しいFieldSearchQueryを作成する
func NewFieldSearchQuery(db *pgxpool.Pool) query.FieldSearchQuery {
	return &fieldSearchQuery{
		db: db,
	}
}

// SearchByBBox はH3+GISTハイブリッドアプローチでBBox内の圃場を検索する
func (q *fieldSearchQuery) SearchByBBox(ctx context.Context, bbox *entity.BoundingBox, h3Cells []string, resolution int) ([]*entity.SearchedField, error) {
	if bbox == nil {
		return nil, fmt.Errorf("BoundingBoxがnilです")
	}

	// H3セルが空の場合は空の結果を返す
	if len(h3Cells) == 0 {
		return []*entity.SearchedField{}, nil
	}

	// 解像度に対応するH3インデックスカラム名を取得
	h3Column, err := q.getH3ColumnName(resolution)
	if err != nil {
		return nil, err
	}

	// クエリ実行
	// Phase1: H3 B-Treeインデックスで候補絞り込み
	// Phase2: centroid(POINT)でBBox判定
	sql := fmt.Sprintf(`
		SELECT
			id,
			ST_AsBinary(geometry) AS geometry,
			ST_AsBinary(centroid) AS centroid,
			area_sqm,
			city_code,
			name,
			soil_type_id
		FROM fields
		WHERE
			%s = ANY($1::varchar[])
			AND centroid && ST_MakeEnvelope($2, $3, $4, $5, 4326)
	`, h3Column)

	rows, err := q.db.Query(ctx, sql,
		h3Cells,
		bbox.SwLng(), bbox.SwLat(), // 南西(経度, 緯度)
		bbox.NeLng(), bbox.NeLat(), // 北東(経度, 緯度)
	)
	if err != nil {
		return nil, fmt.Errorf("圃場検索クエリの実行に失敗: %w", err)
	}
	defer rows.Close()

	// 結果をエンティティに変換
	var fields []*entity.SearchedField
	for rows.Next() {
		field, err := q.scanRow(rows)
		if err != nil {
			return nil, fmt.Errorf("圃場データのスキャンに失敗: %w", err)
		}
		fields = append(fields, field)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("圃場検索結果の読み取りに失敗: %w", err)
	}

	return fields, nil
}

// getH3ColumnName は解像度に対応するH3インデックスカラム名を返す
func (q *fieldSearchQuery) getH3ColumnName(resolution int) (string, error) {
	switch resolution {
	case 3:
		return "h3_index_res3", nil
	case 5:
		return "h3_index_res5", nil
	case 7:
		return "h3_index_res7", nil
	case 9:
		return "h3_index_res9", nil
	default:
		return "", fmt.Errorf("サポートされていないH3解像度: %d(サポート: 3, 5, 7, 9)", resolution)
	}
}

// scanRow は行をSearchedFieldエンティティにスキャンする
func (q *fieldSearchQuery) scanRow(rows interface{ Scan(dest ...any) error }) (*entity.SearchedField, error) {
	var (
		field       entity.SearchedField
		geometryWKB []byte
		centroidWKB []byte
	)

	err := rows.Scan(
		&field.ID,
		&geometryWKB,
		&centroidWKB,
		&field.AreaSqm,
		&field.CityCode,
		&field.Name,
		&field.SoilTypeID,
	)
	if err != nil {
		return nil, err
	}

	// Geometry変換(WKB -> *geom.Polygon)
	if len(geometryWKB) > 0 {
		g, err := wkb.Unmarshal(geometryWKB)
		if err != nil {
			return nil, fmt.Errorf("geometry WKBデコードに失敗: %w", err)
		}
		if polygon, ok := g.(*geom.Polygon); ok {
			field.Geometry = polygon
		}
	}

	// Centroid変換(WKB -> *geom.Point)
	if len(centroidWKB) > 0 {
		c, err := wkb.Unmarshal(centroidWKB)
		if err != nil {
			return nil, fmt.Errorf("centroid WKBデコードに失敗: %w", err)
		}
		if point, ok := c.(*geom.Point); ok {
			field.Centroid = point
		}
	}

	return &field, nil
}
