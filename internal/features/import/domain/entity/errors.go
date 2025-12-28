package entity

import "errors"

var (
	// ErrInvalidStatusTransition は無効なステータス遷移エラー
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)
