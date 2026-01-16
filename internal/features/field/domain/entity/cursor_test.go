package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFieldCursor_Encode_Success はカーソルのエンコードが正常に動作することをテストする
func TestFieldCursor_Encode_Success(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	cursor := NewFieldCursor(createdAt, id)

	encoded, err := cursor.Encode()

	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

// TestFieldCursor_Encode_Decode_Roundtrip はエンコード・デコードのラウンドトリップをテストする
func TestFieldCursor_Encode_Decode_Roundtrip(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 123456000, time.UTC)
	original := NewFieldCursor(createdAt, id)

	encoded, err := original.Encode()
	require.NoError(t, err)

	decoded, err := DecodeFieldCursor(encoded)

	require.NoError(t, err)
	require.NotNil(t, decoded)
	assert.Equal(t, original.ID, decoded.ID)
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt))
}

// TestDecodeFieldCursor_EmptyString は空文字列でnilが返ることをテストする
func TestDecodeFieldCursor_EmptyString(t *testing.T) {
	cursor, err := DecodeFieldCursor("")

	require.NoError(t, err)
	assert.Nil(t, cursor)
}

// TestDecodeFieldCursor_InvalidBase64 は不正なBase64でエラーが返ることをテストする
func TestDecodeFieldCursor_InvalidBase64(t *testing.T) {
	cursor, err := DecodeFieldCursor("not-valid-base64!!!")

	assert.Error(t, err)
	assert.Nil(t, cursor)
	assert.Contains(t, err.Error(), "Base64デコードに失敗")
}

// TestDecodeFieldCursor_InvalidJSON は不正なJSONでエラーが返ることをテストする
func TestDecodeFieldCursor_InvalidJSON(t *testing.T) {
	// Base64エンコードされた不正なJSON
	invalidJSON := "bm90LWpzb24="

	cursor, err := DecodeFieldCursor(invalidJSON)

	assert.Error(t, err)
	assert.Nil(t, cursor)
	assert.Contains(t, err.Error(), "JSON解析に失敗")
}

// TestNewFieldCursor は新しいカーソルが正しく作成されることをテストする
func TestNewFieldCursor(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()

	cursor := NewFieldCursor(createdAt, id)

	assert.Equal(t, id, cursor.ID)
	assert.True(t, createdAt.Equal(cursor.CreatedAt))
}
