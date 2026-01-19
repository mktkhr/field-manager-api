// Package query は圃場検索機能のクエリインターフェースを提供する
package query

import (
	"context"

	"github.com/mktkhr/field-manager-api/internal/features/fieldsearch/domain/entity"
)

// FieldSearchQuery は圃場検索のクエリインターフェース
type FieldSearchQuery interface {
	// SearchByBBox はH3+GISTハイブリッドアプローチでBBox内の圃場を検索する
	// bbox: 検索範囲
	// h3Cells: H3セル(事前絞り込み用)
	// resolution: 使用するH3解像度(3, 5, 7, 9のいずれか)
	SearchByBBox(ctx context.Context, bbox *entity.BoundingBox, h3Cells []string, resolution int) ([]*entity.SearchedField, error)
}
