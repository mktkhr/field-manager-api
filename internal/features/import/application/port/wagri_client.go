package port

import (
	"context"

	"github.com/mktkhr/field-manager-api/internal/features/import/domain/entity"
)

// WagriClient はwagri API操作のインターフェース
type WagriClient interface {
	// FetchFieldsByCityCode は市区町村コードで圃場データを取得する
	FetchFieldsByCityCode(ctx context.Context, cityCode string) (*entity.WagriResponse, error)

	// FetchFieldsByCityCodeToStream は市区町村コードで圃場データを取得し、ストリームとして返す
	FetchFieldsByCityCodeToStream(ctx context.Context, cityCode string) ([]byte, error)
}
