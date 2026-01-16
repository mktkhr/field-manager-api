package entity

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FieldCursor は圃場一覧取得用のカーソル
type FieldCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

// NewFieldCursor は新しいFieldCursorを作成する
func NewFieldCursor(createdAt time.Time, id uuid.UUID) *FieldCursor {
	return &FieldCursor{
		CreatedAt: createdAt,
		ID:        id,
	}
}

// Encode はカーソルをBase64エンコードされた文字列に変換する
func (c *FieldCursor) Encode() (string, error) {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("カーソルのJSON変換に失敗: %w", err)
	}
	return base64.URLEncoding.EncodeToString(jsonBytes), nil
}

// DecodeFieldCursor はBase64文字列からカーソルをデコードする
// 空文字列の場合はnilを返す(エラーではない)
func DecodeFieldCursor(encoded string) (*FieldCursor, error) {
	if encoded == "" {
		return nil, nil
	}

	jsonBytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("カーソルのBase64デコードに失敗: %w", err)
	}

	var cursor FieldCursor
	if err := json.Unmarshal(jsonBytes, &cursor); err != nil {
		return nil, fmt.Errorf("カーソルのJSON解析に失敗: %w", err)
	}

	return &cursor, nil
}
