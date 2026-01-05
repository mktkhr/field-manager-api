// Package query は圃場機能のクエリインターフェースを提供する
package query

import (
	"context"

	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
)

// FieldQuery は圃場の照会インターフェース
type FieldQuery interface {
	// List は圃場一覧を取得する
	// limit: 取得件数, offset: スキップ件数
	List(ctx context.Context, limit, offset int32) ([]*entity.Field, error)

	// Count は圃場の総数を取得する
	Count(ctx context.Context) (int64, error)
}
