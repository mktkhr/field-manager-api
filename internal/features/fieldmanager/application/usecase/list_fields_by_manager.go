package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/repository"
)

// ListFieldsByManagerInput は管理配下圃場一覧取得ユースケースの入力
type ListFieldsByManagerInput struct {
	ManagerType     string
	ManagerID       uuid.UUID
	CursorCreatedAt *time.Time
	CursorID        *uuid.UUID
	PageSize        int
}

// ListFieldsByManagerOutput は管理配下圃場一覧取得ユースケースの出力
type ListFieldsByManagerOutput struct {
	FieldManagers []*entity.FieldManager
	HasMore       bool
}

// ListFieldsByManagerUseCase は管理配下圃場一覧取得ユースケース
type ListFieldsByManagerUseCase struct {
	repo   repository.FieldManagerRepository
	logger *slog.Logger
}

// NewListFieldsByManagerUseCase はListFieldsByManagerUseCaseを作成する
func NewListFieldsByManagerUseCase(
	repo repository.FieldManagerRepository,
	logger *slog.Logger,
) *ListFieldsByManagerUseCase {
	return &ListFieldsByManagerUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute は管理配下圃場一覧取得を実行する
func (u *ListFieldsByManagerUseCase) Execute(ctx context.Context, input ListFieldsByManagerInput) (*ListFieldsByManagerOutput, error) {
	// 管理者タイプの検証
	managerType, err := entity.ParseManagerType(input.ManagerType)
	if err != nil {
		u.logger.Warn("無効な管理者タイプ",
			slog.String("managerType", input.ManagerType),
			slog.String("error", err.Error()))
		return nil, err
	}

	// ページサイズのデフォルト設定
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	// 1件多く取得してhasMoreを判定
	params := repository.CursorParams{
		CursorCreatedAt: input.CursorCreatedAt,
		CursorID:        input.CursorID,
		PageLimit:       pageSize + 1,
	}

	managers, err := u.repo.ListByManager(ctx, managerType, input.ManagerID, params)
	if err != nil {
		u.logger.Error("管理配下圃場一覧の取得に失敗しました",
			slog.String("managerType", string(managerType)),
			slog.String("managerId", input.ManagerID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("管理配下圃場一覧の取得に失敗しました: %w", err)
	}

	hasMore := len(managers) > pageSize
	if hasMore {
		managers = managers[:pageSize]
	}

	u.logger.Debug("管理配下圃場一覧を取得しました",
		slog.String("managerType", string(managerType)),
		slog.String("managerId", input.ManagerID.String()),
		slog.Int("count", len(managers)),
		slog.Bool("hasMore", hasMore))

	return &ListFieldsByManagerOutput{
		FieldManagers: managers,
		HasMore:       hasMore,
	}, nil
}
