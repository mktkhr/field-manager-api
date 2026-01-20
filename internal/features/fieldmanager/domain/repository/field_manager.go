// Package repository は圃場管理者機能のリポジトリインターフェースを提供する
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/entity"
)

// CursorParams はカーソルページネーションのパラメータ
type CursorParams struct {
	CursorCreatedAt *time.Time
	CursorID        *uuid.UUID
	PageLimit       int
}

// FieldManagerRepository は圃場管理関係のリポジトリインターフェース
type FieldManagerRepository interface {
	// FindByID はIDで管理関係を取得する
	FindByID(ctx context.Context, id uuid.UUID) (*entity.FieldManager, error)

	// FindByFieldAndManager は圃場IDと管理者情報で管理関係を取得する
	FindByFieldAndManager(ctx context.Context, fieldID uuid.UUID, managerType entity.ManagerType, managerID uuid.UUID) (*entity.FieldManager, error)

	// ListByField は圃場IDで管理者一覧を取得する
	ListByField(ctx context.Context, fieldID uuid.UUID) ([]*entity.FieldManager, error)

	// ListByManager は管理者タイプと管理者IDで管理配下の圃場管理関係一覧を取得する(カーソルページネーション)
	ListByManager(ctx context.Context, managerType entity.ManagerType, managerID uuid.UUID, params CursorParams) ([]*entity.FieldManager, error)

	// Create は管理関係を作成する
	Create(ctx context.Context, fm *entity.FieldManager) error

	// Delete はIDで管理関係を削除する
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists は管理関係の存在を確認する
	Exists(ctx context.Context, fieldID uuid.UUID, managerType entity.ManagerType, managerID uuid.UUID) (bool, error)
}
