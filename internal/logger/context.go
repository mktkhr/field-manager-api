package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	// RequestIDContextKey はContextでリクエストIDを保存するキー
	RequestIDContextKey contextKey = "request_id"
	// LoggerContextKey はContextでロガーを保存するキー
	LoggerContextKey contextKey = "logger"
)

// WithRequestID はContextにリクエストIDを設定する
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDContextKey, requestID)
}

// GetRequestIDFromContext はContextからリクエストIDを取得する
func GetRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDContextKey).(string); ok {
		return id
	}
	return ""
}

// WithLogger はContextにロガーを設定する
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}

// GetLoggerFromContext はContextからロガーを取得する
// ロガーが設定されていない場合はデフォルトロガーを返す
func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerContextKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// NewRequestLogger はリクエスト専用のロガーを作成する
func NewRequestLogger(requestID string, userAgent string, clientIP string) *slog.Logger {
	return slog.Default().With(
		slog.String("request_id", requestID),
		slog.String("user_agent", userAgent),
		slog.String("client_ip", clientIP),
	)
}

// InfoWithContext はリクエストIDを含めてInfoログを出力する
func InfoWithContext(ctx context.Context, msg string, args ...any) {
	requestID := GetRequestIDFromContext(ctx)
	if requestID != "" {
		newArgs := append([]any{"request_id", requestID}, args...)
		slog.InfoContext(ctx, msg, newArgs...)
	} else {
		slog.InfoContext(ctx, msg, args...)
	}
}

// ErrorWithContext はリクエストIDを含めてErrorログを出力する
func ErrorWithContext(ctx context.Context, msg string, args ...any) {
	requestID := GetRequestIDFromContext(ctx)
	if requestID != "" {
		newArgs := append([]any{"request_id", requestID}, args...)
		slog.ErrorContext(ctx, msg, newArgs...)
	} else {
		slog.ErrorContext(ctx, msg, args...)
	}
}

// WarnWithContext はリクエストIDを含めてWarnログを出力する
func WarnWithContext(ctx context.Context, msg string, args ...any) {
	requestID := GetRequestIDFromContext(ctx)
	if requestID != "" {
		newArgs := append([]any{"request_id", requestID}, args...)
		slog.WarnContext(ctx, msg, newArgs...)
	} else {
		slog.WarnContext(ctx, msg, args...)
	}
}

// DebugWithContext はリクエストIDを含めてDebugログを出力する
func DebugWithContext(ctx context.Context, msg string, args ...any) {
	requestID := GetRequestIDFromContext(ctx)
	if requestID != "" {
		newArgs := append([]any{"request_id", requestID}, args...)
		slog.DebugContext(ctx, msg, newArgs...)
	} else {
		slog.DebugContext(ctx, msg, args...)
	}
}
