package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/repository"
)

// ListManagersByFieldInput は圃場管理者一覧取得ユースケースの入力
type ListManagersByFieldInput struct {
	FieldID uuid.UUID
}

// ListManagersByFieldOutput は圃場管理者一覧取得ユースケースの出力
type ListManagersByFieldOutput struct {
	FieldManagers []*entity.FieldManager
}

// ListManagersByFieldUseCase は圃場管理者一覧取得ユースケース
type ListManagersByFieldUseCase struct {
	repo         repository.FieldManagerRepository
	fieldChecker FieldExistsChecker
	logger       *slog.Logger
}

// NewListManagersByFieldUseCase はListManagersByFieldUseCaseを作成する
func NewListManagersByFieldUseCase(
	repo repository.FieldManagerRepository,
	fieldChecker FieldExistsChecker,
	logger *slog.Logger,
) *ListManagersByFieldUseCase {
	return &ListManagersByFieldUseCase{
		repo:         repo,
		fieldChecker: fieldChecker,
		logger:       logger,
	}
}

// Execute は圃場管理者一覧取得を実行する
func (u *ListManagersByFieldUseCase) Execute(ctx context.Context, input ListManagersByFieldInput) (*ListManagersByFieldOutput, error) {
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

	managers, err := u.repo.ListByField(ctx, input.FieldID)
	if err != nil {
		u.logger.Error("圃場管理者一覧の取得に失敗しました",
			slog.String("fieldId", input.FieldID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("圃場管理者一覧の取得に失敗しました: %w", err)
	}

	u.logger.Debug("圃場管理者一覧を取得しました",
		slog.String("fieldId", input.FieldID.String()),
		slog.Int("count", len(managers)))

	return &ListManagersByFieldOutput{
		FieldManagers: managers,
	}, nil
}
