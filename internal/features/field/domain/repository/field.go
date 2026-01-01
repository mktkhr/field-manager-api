package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/field/domain/entity"
)

// FieldRepository は圃場のリポジトリインターフェース
type FieldRepository interface {
	// FindByID はIDで圃場を取得する
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Field, error)

	// Create は圃場を作成する
	Create(ctx context.Context, field *entity.Field) error

	// Update は圃場を更新する
	Update(ctx context.Context, field *entity.Field) error

	// Delete は圃場を削除する
	Delete(ctx context.Context, id uuid.UUID) error

	// Upsert は圃場をUPSERTする
	Upsert(ctx context.Context, field *entity.Field) error
}
