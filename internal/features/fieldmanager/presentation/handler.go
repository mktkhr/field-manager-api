// Package presentation は圃場管理者機能のHTTPハンドラーを提供する
package presentation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/application/usecase"
	"github.com/mktkhr/field-manager-api/internal/features/fieldmanager/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/openapi"
)

const (
	// DefaultPageSize はデフォルトのページサイズ
	DefaultPageSize = 20
	// MaxPageSize は最大ページサイズ
	MaxPageSize = 1000
)

// FieldManagerHandler は圃場管理者APIのハンドラー
type FieldManagerHandler struct {
	assignManagerUC       *usecase.AssignManagerUseCase
	unassignManagerUC     *usecase.UnassignManagerUseCase
	listFieldsByManagerUC *usecase.ListFieldsByManagerUseCase
	listManagersByFieldUC *usecase.ListManagersByFieldUseCase
	logger                *slog.Logger
}

// NewFieldManagerHandler はFieldManagerHandlerを作成する
func NewFieldManagerHandler(
	assignManagerUC *usecase.AssignManagerUseCase,
	unassignManagerUC *usecase.UnassignManagerUseCase,
	listFieldsByManagerUC *usecase.ListFieldsByManagerUseCase,
	listManagersByFieldUC *usecase.ListManagersByFieldUseCase,
	logger *slog.Logger,
) *FieldManagerHandler {
	return &FieldManagerHandler{
		assignManagerUC:       assignManagerUC,
		unassignManagerUC:     unassignManagerUC,
		listFieldsByManagerUC: listFieldsByManagerUC,
		listManagersByFieldUC: listManagersByFieldUC,
		logger:                logger,
	}
}

// ListManagersByField は圃場の管理者一覧を取得する
func (h *FieldManagerHandler) ListManagersByField(ctx context.Context, request openapi.ListManagersByFieldRequestObject) (openapi.ListManagersByFieldResponseObject, error) {
	output, err := h.listManagersByFieldUC.Execute(ctx, usecase.ListManagersByFieldInput{
		FieldID: request.FieldId,
	})
	if err != nil {
		if isNotFoundError(err) {
			return openapi.ListManagersByField404JSONResponse{
				NotFoundJSONResponse: openapi.NotFoundJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "not_found",
						Message: err.Error(),
					}},
				},
			}, nil
		}
		h.logger.Error("圃場管理者一覧取得に失敗しました",
			slog.String("fieldId", request.FieldId.String()),
			slog.String("error", err.Error()))
		return openapi.ListManagersByField500JSONResponse{
			InternalServerErrorJSONResponse: openapi.InternalServerErrorJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "internal_error",
					Message: "圃場管理者一覧の取得に失敗しました",
				}},
			},
		}, nil
	}

	managers := make([]openapi.FieldManager, 0, len(output.FieldManagers))
	for _, fm := range output.FieldManagers {
		managers = append(managers, h.toOpenAPIFieldManager(fm))
	}

	return openapi.ListManagersByField200JSONResponse{
		Data: openapi.FieldManagerListData{
			Managers: managers,
		},
		Errors: nil,
	}, nil
}

// AssignManager は管理者を割り当てる
func (h *FieldManagerHandler) AssignManager(ctx context.Context, request openapi.AssignManagerRequestObject) (openapi.AssignManagerResponseObject, error) {
	if request.Body == nil {
		return openapi.AssignManager400JSONResponse{
			BadRequestJSONResponse: openapi.BadRequestJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "invalid_parameter",
					Message: "リクエストボディが必要です",
				}},
			},
		}, nil
	}

	output, err := h.assignManagerUC.Execute(ctx, usecase.AssignManagerInput{
		FieldID:     request.FieldId,
		ManagerType: string(request.Body.FieldManagerType),
		ManagerID:   request.Body.ManagerId,
		CreatedBy:   nil, // TODO: 認証情報から取得
	})
	if err != nil {
		if isValidationError(err) {
			return openapi.AssignManager400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: err.Error(),
					}},
				},
			}, nil
		}
		if isNotFoundError(err) {
			return openapi.AssignManager404JSONResponse{
				NotFoundJSONResponse: openapi.NotFoundJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "not_found",
						Message: err.Error(),
					}},
				},
			}, nil
		}
		if isConflictError(err) {
			return openapi.AssignManager409JSONResponse{
				ConflictJSONResponse: openapi.ConflictJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "conflict",
						Message: err.Error(),
					}},
				},
			}, nil
		}
		h.logger.Error("管理者割り当てに失敗しました",
			slog.String("fieldId", request.FieldId.String()),
			slog.String("error", err.Error()))
		return openapi.AssignManager500JSONResponse{
			InternalServerErrorJSONResponse: openapi.InternalServerErrorJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "internal_error",
					Message: "管理者の割り当てに失敗しました",
				}},
			},
		}, nil
	}

	return openapi.AssignManager201JSONResponse{
		Data:   h.toOpenAPIFieldManager(output.FieldManager),
		Errors: nil,
	}, nil
}

// UnassignManager は管理者を解除する
func (h *FieldManagerHandler) UnassignManager(ctx context.Context, request openapi.UnassignManagerRequestObject) (openapi.UnassignManagerResponseObject, error) {
	err := h.unassignManagerUC.Execute(ctx, usecase.UnassignManagerInput{
		FieldID:        request.FieldId,
		FieldManagerID: request.FieldManagerId,
	})
	if err != nil {
		if isNotFoundError(err) {
			return openapi.UnassignManager404JSONResponse{
				NotFoundJSONResponse: openapi.NotFoundJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "not_found",
						Message: err.Error(),
					}},
				},
			}, nil
		}
		if isValidationError(err) {
			return openapi.UnassignManager400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: err.Error(),
					}},
				},
			}, nil
		}
		h.logger.Error("管理者解除に失敗しました",
			slog.String("fieldId", request.FieldId.String()),
			slog.String("fieldManagerId", request.FieldManagerId.String()),
			slog.String("error", err.Error()))
		return openapi.UnassignManager500JSONResponse{
			InternalServerErrorJSONResponse: openapi.InternalServerErrorJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "internal_error",
					Message: "管理者の解除に失敗しました",
				}},
			},
		}, nil
	}

	return openapi.UnassignManager204Response{}, nil
}

// ListFieldsByManager は管理配下の圃場一覧を取得する
func (h *FieldManagerHandler) ListFieldsByManager(ctx context.Context, request openapi.ListFieldsByManagerRequestObject) (openapi.ListFieldsByManagerResponseObject, error) {
	// カーソルをデコード
	var cursorCreatedAt *time.Time
	var cursorID *uuid.UUID
	if request.Params.Cursor != nil && *request.Params.Cursor != "" {
		decoded, err := decodeCursor(*request.Params.Cursor)
		if err != nil {
			return openapi.ListFieldsByManager400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: "カーソルの形式が不正です",
					}},
				},
			}, nil
		}
		cursorCreatedAt = &decoded.CreatedAt
		cursorID = &decoded.ID
	}

	// ページサイズ
	pageSize := DefaultPageSize
	if request.Params.PageSize != nil {
		ps := *request.Params.PageSize
		if ps < 1 || ps > MaxPageSize {
			return openapi.ListFieldsByManager400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: "ページサイズは1から1000の範囲で指定してください",
					}},
				},
			}, nil
		}
		pageSize = ps
	}

	output, err := h.listFieldsByManagerUC.Execute(ctx, usecase.ListFieldsByManagerInput{
		ManagerType:     string(request.ManagerType),
		ManagerID:       request.ManagerId,
		CursorCreatedAt: cursorCreatedAt,
		CursorID:        cursorID,
		PageSize:        pageSize,
	})
	if err != nil {
		if isValidationError(err) {
			return openapi.ListFieldsByManager400JSONResponse{
				BadRequestJSONResponse: openapi.BadRequestJSONResponse{
					Data: nil,
					Errors: &[]openapi.Error{{
						Code:    "invalid_parameter",
						Message: err.Error(),
					}},
				},
			}, nil
		}
		h.logger.Error("管理配下圃場一覧取得に失敗しました",
			slog.String("managerType", string(request.ManagerType)),
			slog.String("managerId", request.ManagerId.String()),
			slog.String("error", err.Error()))
		return openapi.ListFieldsByManager500JSONResponse{
			InternalServerErrorJSONResponse: openapi.InternalServerErrorJSONResponse{
				Data: nil,
				Errors: &[]openapi.Error{{
					Code:    "internal_error",
					Message: "管理配下圃場一覧の取得に失敗しました",
				}},
			},
		}, nil
	}

	// 圃場IDのリストを作成
	fieldIDs := make([]uuid.UUID, 0, len(output.FieldManagers))
	for _, fm := range output.FieldManagers {
		fieldIDs = append(fieldIDs, fm.FieldID)
	}

	// 次のカーソルを生成
	var nextCursor *string
	if output.HasMore && len(output.FieldManagers) > 0 {
		lastFM := output.FieldManagers[len(output.FieldManagers)-1]
		cursor := encodeCursor(lastFM.CreatedAt, lastFM.ID)
		nextCursor = &cursor
	}

	return openapi.ListFieldsByManager200JSONResponse{
		Data: openapi.FieldIdListData{
			FieldIds:   fieldIDs,
			NextCursor: nextCursor,
		},
		Errors: nil,
	}, nil
}

// toOpenAPIFieldManager はエンティティをOpenAPIの型に変換する
func (h *FieldManagerHandler) toOpenAPIFieldManager(fm *entity.FieldManager) openapi.FieldManager {
	return openapi.FieldManager{
		Id:               fm.ID,
		FieldId:          fm.FieldID,
		FieldManagerType: openapi.FieldManagerType(fm.ManagerType),
		ManagerId:        fm.ManagerID,
		CreatedAt:        fm.CreatedAt,
		CreatedBy:        fm.CreatedBy,
	}
}

// Cursor はカーソルの内部構造
type Cursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        uuid.UUID `json:"id"`
}

// encodeCursor はカーソルをBase64エンコードする
func encodeCursor(createdAt time.Time, id uuid.UUID) string {
	cursor := Cursor{CreatedAt: createdAt, ID: id}
	data, _ := json.Marshal(cursor)
	return base64.URLEncoding.EncodeToString(data)
}

// decodeCursor はBase64エンコードされたカーソルをデコードする
func decodeCursor(encoded string) (*Cursor, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, err
	}
	return &cursor, nil
}

// isNotFoundError はエラーがNotFoundエラーかどうかを判定する
func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "存在しません")
}

// isValidationError はエラーがバリデーションエラーかどうかを判定する
func isValidationError(err error) bool {
	keywords := []string{"無効な", "不正", "属していません"}
	for _, keyword := range keywords {
		if strings.Contains(err.Error(), keyword) {
			return true
		}
	}
	return false
}

// isConflictError はエラーが競合エラーかどうかを判定する
func isConflictError(err error) bool {
	return strings.Contains(err.Error(), "既に存在します")
}
