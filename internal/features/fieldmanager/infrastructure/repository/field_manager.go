// Package repository は圃場管理者機能のリポジトリ実装を提供する
package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/repository"
)

// fieldManagerRepository はFieldManagerRepositoryの実装
type fieldManagerRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewFieldManagerRepository は新しいFieldManagerRepositoryを作成する
func NewFieldManagerRepository(db *pgxpool.Pool, logger *slog.Logger) repository.FieldManagerRepository {
	return &fieldManagerRepository{
		db:     db,
		logger: logger,
	}
}

// FindByID はIDで管理関係を取得する
func (r *fieldManagerRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.FieldManager, error) {
	query := `
		SELECT id, field_id, manager_type, manager_id, created_at, created_by
		FROM field_managers
		WHERE id = $1
	`

	var fm entity.FieldManager
	var managerType string
	var createdBy *uuid.UUID

	err := r.db.QueryRow(ctx, query, id).Scan(
		&fm.ID,
		&fm.FieldID,
		&managerType,
		&fm.ManagerID,
		&fm.CreatedAt,
		&createdBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("管理関係の取得に失敗: %w", err)
	}

	fm.ManagerType = entity.ManagerType(managerType)
	fm.CreatedBy = createdBy

	return &fm, nil
}

// FindByFieldAndManager は圃場IDと管理者情報で管理関係を取得する
func (r *fieldManagerRepository) FindByFieldAndManager(ctx context.Context, fieldID uuid.UUID, managerType entity.ManagerType, managerID uuid.UUID) (*entity.FieldManager, error) {
	query := `
		SELECT id, field_id, manager_type, manager_id, created_at, created_by
		FROM field_managers
		WHERE field_id = $1 AND manager_type = $2 AND manager_id = $3
	`

	var fm entity.FieldManager
	var mt string
	var createdBy *uuid.UUID

	err := r.db.QueryRow(ctx, query, fieldID, string(managerType), managerID).Scan(
		&fm.ID,
		&fm.FieldID,
		&mt,
		&fm.ManagerID,
		&fm.CreatedAt,
		&createdBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("管理関係の取得に失敗: %w", err)
	}

	fm.ManagerType = entity.ManagerType(mt)
	fm.CreatedBy = createdBy

	return &fm, nil
}

// ListByField は圃場IDで管理者一覧を取得する
func (r *fieldManagerRepository) ListByField(ctx context.Context, fieldID uuid.UUID) ([]*entity.FieldManager, error) {
	query := `
		SELECT id, field_id, manager_type, manager_id, created_at, created_by
		FROM field_managers
		WHERE field_id = $1
		ORDER BY manager_type, created_at DESC
	`

	rows, err := r.db.Query(ctx, query, fieldID)
	if err != nil {
		return nil, fmt.Errorf("管理者一覧の取得に失敗: %w", err)
	}
	defer rows.Close()

	var managers []*entity.FieldManager
	for rows.Next() {
		fm, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		managers = append(managers, fm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("管理者一覧の読み取りに失敗: %w", err)
	}

	return managers, nil
}

// ListByManager は管理者タイプと管理者IDで管理配下の圃場管理関係一覧を取得する
func (r *fieldManagerRepository) ListByManager(ctx context.Context, managerType entity.ManagerType, managerID uuid.UUID, params repository.CursorParams) ([]*entity.FieldManager, error) {
	query := `
		SELECT id, field_id, manager_type, manager_id, created_at, created_by
		FROM field_managers
		WHERE manager_type = $1 AND manager_id = $2
			AND CASE
				WHEN $3::timestamptz IS NULL THEN TRUE
				ELSE (created_at < $3)
					 OR (created_at = $3 AND id < $4)
			END
		ORDER BY created_at DESC, id DESC
		LIMIT $5
	`

	rows, err := r.db.Query(ctx, query,
		string(managerType),
		managerID,
		params.CursorCreatedAt,
		params.CursorID,
		params.PageLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("管理配下一覧の取得に失敗: %w", err)
	}
	defer rows.Close()

	var managers []*entity.FieldManager
	for rows.Next() {
		fm, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		managers = append(managers, fm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("管理配下一覧の読み取りに失敗: %w", err)
	}

	return managers, nil
}

// Create は管理関係を作成する
func (r *fieldManagerRepository) Create(ctx context.Context, fm *entity.FieldManager) error {
	query := `
		INSERT INTO field_managers (id, field_id, manager_type, manager_id, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		fm.ID,
		fm.FieldID,
		string(fm.ManagerType),
		fm.ManagerID,
		fm.CreatedAt,
		fm.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("管理関係の作成に失敗: %w", err)
	}

	return nil
}

// Delete はIDで管理関係を削除する
func (r *fieldManagerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM field_managers WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("管理関係の削除に失敗: %w", err)
	}

	return nil
}

// Exists は管理関係の存在を確認する
func (r *fieldManagerRepository) Exists(ctx context.Context, fieldID uuid.UUID, managerType entity.ManagerType, managerID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM field_managers
			WHERE field_id = $1 AND manager_type = $2 AND manager_id = $3
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, fieldID, string(managerType), managerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("管理関係の存在確認に失敗: %w", err)
	}

	return exists, nil
}

// scanRow は行をFieldManagerエンティティにスキャンする
func (r *fieldManagerRepository) scanRow(rows pgx.Rows) (*entity.FieldManager, error) {
	var fm entity.FieldManager
	var managerType string
	var createdBy *uuid.UUID

	err := rows.Scan(
		&fm.ID,
		&fm.FieldID,
		&managerType,
		&fm.ManagerID,
		&fm.CreatedAt,
		&createdBy,
	)
	if err != nil {
		return nil, fmt.Errorf("管理関係のスキャンに失敗: %w", err)
	}

	fm.ManagerType = entity.ManagerType(managerType)
	fm.CreatedBy = createdBy

	return &fm, nil
}
