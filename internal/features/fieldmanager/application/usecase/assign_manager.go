// Package usecase は圃場管理者機能のユースケースを提供する
package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/repository"
)

// FieldExistsChecker は圃場の存在確認インターフェース
type FieldExistsChecker interface {
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

// AssignManagerInput は管理者割り当てユースケースの入力
type AssignManagerInput struct {
	FieldID     uuid.UUID
	ManagerType string
	ManagerID   uuid.UUID
	CreatedBy   *uuid.UUID
}

// AssignManagerOutput は管理者割り当てユースケースの出力
type AssignManagerOutput struct {
	FieldManager *entity.FieldManager
}

// AssignManagerUseCase は管理者割り当てユースケース
type AssignManagerUseCase struct {
	repo         repository.FieldManagerRepository
	fieldChecker FieldExistsChecker
	logger       *slog.Logger
}

// NewAssignManagerUseCase はAssignManagerUseCaseを作成する
func NewAssignManagerUseCase(
	repo repository.FieldManagerRepository,
	fieldChecker FieldExistsChecker,
	logger *slog.Logger,
) *AssignManagerUseCase {
	return &AssignManagerUseCase{
		repo:         repo,
		fieldChecker: fieldChecker,
		logger:       logger,
	}
}

// Execute は管理者割り当てを実行する
func (u *AssignManagerUseCase) Execute(ctx context.Context, input AssignManagerInput) (*AssignManagerOutput, error) {
	// 管理者タイプの検証
	managerType, err := entity.ParseManagerType(input.ManagerType)
	if err != nil {
		u.logger.Warn("無効な管理者タイプ",
			slog.String("managerType", input.ManagerType),
			slog.String("error", err.Error()))
		return nil, err
	}

	// 圃場の存在確認
	exists, err := u.fieldChecker.ExistsByID(ctx, input.FieldID)
	if err != nil {
		u.logger.Error("圃場存在確認に失敗しました",
			slog.String("fieldId", input.FieldID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("圃場存在確認に失敗しました: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("指定された圃場が存在しません: %s", input.FieldID)
	}

	// 重複チェック
	duplicateExists, err := u.repo.Exists(ctx, input.FieldID, managerType, input.ManagerID)
	if err != nil {
		u.logger.Error("重複確認に失敗しました",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("重複確認に失敗しました: %w", err)
	}
	if duplicateExists {
		return nil, fmt.Errorf("管理関係は既に存在します: fieldId=%s, managerType=%s, managerId=%s",
			input.FieldID, managerType, input.ManagerID)
	}

	// エンティティ作成
	fm, err := entity.NewFieldManager(input.FieldID, managerType, input.ManagerID, input.CreatedBy)
	if err != nil {
		return nil, err
	}

	// 保存
	if err := u.repo.Create(ctx, fm); err != nil {
		u.logger.Error("管理関係の作成に失敗しました",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("管理関係の作成に失敗しました: %w", err)
	}

	u.logger.Info("管理者を割り当てました",
		slog.String("fieldId", input.FieldID.String()),
		slog.String("managerType", string(managerType)),
		slog.String("managerId", input.ManagerID.String()))

	return &AssignManagerOutput{
		FieldManager: fm,
	}, nil
}
