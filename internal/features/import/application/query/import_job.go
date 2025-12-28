package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

// ImportJobQuery はインポートジョブの照会インターフェース
type ImportJobQuery interface {
	// FindByID はIDでインポートジョブを取得する
	FindByID(ctx context.Context, id uuid.UUID) (*entity.ImportJob, error)

	// List はインポートジョブ一覧を取得する
	List(ctx context.Context, limit, offset int32) ([]*entity.ImportJob, error)

	// ListByCityCode は市区町村コードでインポートジョブ一覧を取得する
	ListByCityCode(ctx context.Context, cityCode string, limit, offset int32) ([]*entity.ImportJob, error)

	// Count はインポートジョブの総数を取得する
	Count(ctx context.Context) (int64, error)

	// CountByStatus はステータス別のインポートジョブ数を取得する
	CountByStatus(ctx context.Context, status entity.ImportStatus) (int64, error)
}
