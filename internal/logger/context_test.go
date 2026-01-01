package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-id"

	newCtx := WithRequestID(ctx, requestID)

	if got := newCtx.Value(RequestIDContextKey); got != requestID {
		t.Errorf("WithRequestID() = %v, 期待値 %v", got, requestID)
	}
}

func TestGetRequestIDFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "リクエストIDが設定されている場合",
			ctx:  context.WithValue(context.Background(), RequestIDContextKey, "test-id"),
			want: "test-id",
		},
		{
			name: "リクエストIDが設定されていない場合",
			ctx:  context.Background(),
			want: "",
		},
		{
			name: "異なる型の値が設定されている場合",
			ctx:  context.WithValue(context.Background(), RequestIDContextKey, 123),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRequestIDFromContext(tt.ctx); got != tt.want {
				t.Errorf("GetRequestIDFromContext() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	newCtx := WithLogger(ctx, logger)

	if got := newCtx.Value(LoggerContextKey); got != logger {
		t.Errorf("WithLogger()でロガーが正しく設定されませんでした")
	}
}

func TestGetLoggerFromContext(t *testing.T) {
	customLogger := slog.Default().With("custom", "logger")

	tests := []struct {
		name        string
		ctx         context.Context
		wantDefault bool
	}{
		{
			name:        "ロガーが設定されている場合",
			ctx:         context.WithValue(context.Background(), LoggerContextKey, customLogger),
			wantDefault: false,
		},
		{
			name:        "ロガーが設定されていない場合",
			ctx:         context.Background(),
			wantDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLoggerFromContext(tt.ctx)
			if got == nil {
				t.Error("GetLoggerFromContext()はnilを返すべきではない")
			}
		})
	}
}

func TestNewRequestLogger(t *testing.T) {
	logger := NewRequestLogger("req-123", "Mozilla/5.0", "192.168.1.1")

	if logger == nil {
		t.Error("NewRequestLogger()はnilを返すべきではない")
	}
}

func TestInfoWithContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "リクエストIDあり",
			ctx:  WithRequestID(context.Background(), "test-id"),
		},
		{
			name: "リクエストIDなし",
			ctx:  context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// パニックしないことを確認
			InfoWithContext(tt.ctx, "test message", "key", "value")
		})
	}
}

func TestErrorWithContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "リクエストIDあり",
			ctx:  WithRequestID(context.Background(), "test-id"),
		},
		{
			name: "リクエストIDなし",
			ctx:  context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ErrorWithContext(tt.ctx, "test error", "key", "value")
		})
	}
}

func TestWarnWithContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "リクエストIDあり",
			ctx:  WithRequestID(context.Background(), "test-id"),
		},
		{
			name: "リクエストIDなし",
			ctx:  context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WarnWithContext(tt.ctx, "test warn", "key", "value")
		})
	}
}

func TestDebugWithContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "リクエストIDあり",
			ctx:  WithRequestID(context.Background(), "test-id"),
		},
		{
			name: "リクエストIDなし",
			ctx:  context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DebugWithContext(tt.ctx, "test debug", "key", "value")
		})
	}
}
