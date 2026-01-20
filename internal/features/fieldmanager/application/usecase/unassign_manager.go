package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/repository"
)

// UnassignManagerInput は管理者解除ユースケースの入力
type UnassignManagerInput struct {
	FieldID        uuid.UUID
	FieldManagerID uuid.UUID
}

// UnassignManagerUseCase は管理者解除ユースケース
type UnassignManagerUseCase struct {
	repo   repository.FieldManagerRepository
	logger *slog.Logger
}

// NewUnassignManagerUseCase はUnassignManagerUseCaseを作成する
func NewUnassignManagerUseCase(
	repo repository.FieldManagerRepository,
	logger *slog.Logger,
) *UnassignManagerUseCase {
	return &UnassignManagerUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute は管理者解除を実行する
func (u *UnassignManagerUseCase) Execute(ctx context.Context, input UnassignManagerInput) error {
	// 管理関係の存在確認
	fm, err := u.repo.FindByID(ctx, input.FieldManagerID)
	if err != nil {
		u.logger.Error("管理関係の取得に失敗しました",
			slog.String("fieldManagerId", input.FieldManagerID.String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("管理関係の取得に失敗しました: %w", err)
	}
	if fm == nil {
		return fmt.Errorf("指定された管理関係が存在しません: %s", input.FieldManagerID)
	}

	// 圃場IDの整合性チェック
	if fm.FieldID != input.FieldID {
		return fmt.Errorf("管理関係が指定された圃場に属していません: fieldId=%s, actualFieldId=%s",
			input.FieldID, fm.FieldID)
	}

	// 削除
	if err := u.repo.Delete(ctx, input.FieldManagerID); err != nil {
		u.logger.Error("管理関係の削除に失敗しました",
			slog.String("fieldManagerId", input.FieldManagerID.String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("管理関係の削除に失敗しました: %w", err)
	}

	u.logger.Info("管理者を解除しました",
		slog.String("fieldId", input.FieldID.String()),
		slog.String("fieldManagerId", input.FieldManagerID.String()),
		slog.String("managerType", string(fm.ManagerType)),
		slog.String("managerId", fm.ManagerID.String()))

	return nil
}
