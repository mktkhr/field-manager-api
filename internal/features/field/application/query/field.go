// Package query は圃場機能のクエリインターフェースを提供する
package query

import (
	"context"

	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
)

// FieldQuery は圃場の照会インターフェース
type FieldQuery interface {
	// ListByCursor はカーソルベースで圃場一覧を取得する
	// cursor: 前ページの最後の圃場のカーソル(nilの場合は先頭から)
	// limit: 取得件数
	ListByCursor(ctx context.Context, cursor *entity.FieldCursor, limit int32) ([]*entity.Field, error)
}
