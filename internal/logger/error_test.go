package logger

import (
	"context"
	"errors"
	"testing"
)

func TestLogDatabaseError(t *testing.T) {
	ctx := context.Background()
	err := errors.New("database connection failed")

	// パニックしないことを確認
	LogDatabaseError(ctx, "INSERT", "users", err)
	LogDatabaseError(ctx, "SELECT", "posts", err, "user_id", 123)
}

func TestLogBusinessError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "エラーあり",
			err:  errors.New("business rule violation"),
		},
		{
			name: "エラーなし",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LogBusinessError(ctx, "user_age_validation", tt.err)
			LogBusinessError(ctx, "order_limit", tt.err, "limit", 100)
		})
	}
}

func TestLogValidationError(t *testing.T) {
	ctx := context.Background()

	LogValidationError(ctx, "email", "invalid@", "email_format")
	LogValidationError(ctx, "age", -1, "positive_number", "min", 0)
}

func TestLogExternalAPIError(t *testing.T) {
	ctx := context.Background()
	err := errors.New("connection timeout")

	LogExternalAPIError(ctx, "PaymentAPI", "https://api.example.com/pay", 500, err)
	LogExternalAPIError(ctx, "NotificationAPI", "https://api.example.com/notify", 503, err, "retry_count", 3)
}

func TestLogErrorWithContext(t *testing.T) {
	ctx := context.Background()
	err := errors.New("generic error")

	LogErrorWithContext(ctx, "custom", "something went wrong", err)
	LogErrorWithContext(ctx, "auth", "authentication failed", err, "user_id", "123")
}

func TestGetCallerMethod(t *testing.T) {
	result := getCallerMethod(1)

	if result == "" || result == "unknown" {
		t.Errorf("getCallerMethod() should return valid method info, got %q", result)
	}

	// ファイル名と行番号を含むことを確認
	if len(result) < 10 {
		t.Errorf("getCallerMethod() result too short: %q", result)
	}
}

func TestGetCallerMethodUnknown(t *testing.T) {
	// 非常に大きなskip値でruntime.Callerを失敗させる
	result := getCallerMethod(1000)

	if result != "unknown" {
		t.Errorf("getCallerMethod(1000) should return 'unknown', got %q", result)
	}
}

func TestExtractDatabaseErrorInfo(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{
			name:     "nilエラー",
			err:      nil,
			wantCode: "",
		},
		{
			name:     "PostgreSQL SQLSTATE",
			err:      errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)"),
			wantCode: "23505",
		},
		{
			name:     "duplicate key",
			err:      errors.New("duplicate key error"),
			wantCode: "DUPLICATE_KEY",
		},
		{
			name:     "unique constraint",
			err:      errors.New("unique constraint violation"),
			wantCode: "DUPLICATE_KEY",
		},
		{
			name:     "foreign key",
			err:      errors.New("foreign key constraint fails"),
			wantCode: "FOREIGN_KEY_VIOLATION",
		},
		{
			name:     "不明なエラー",
			err:      errors.New("some unknown error"),
			wantCode: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDatabaseErrorInfo(tt.err)
			if tt.err == nil {
				if result != nil {
					t.Errorf("extractDatabaseErrorInfo() should return nil for nil error")
				}
				return
			}
			if result == nil {
				t.Fatal("extractDatabaseErrorInfo() should not return nil for non-nil error")
			}
			if result.Code != tt.wantCode {
				t.Errorf("extractDatabaseErrorInfo().Code = %q, want %q", result.Code, tt.wantCode)
			}
		})
	}
}

func TestExtractSQLState(t *testing.T) {
	tests := []struct {
		name     string
		errorStr string
		want     string
	}{
		{
			name:     "正常なSQLSTATE",
			errorStr: "ERROR: duplicate key (SQLSTATE 23505)",
			want:     "23505",
		},
		{
			name:     "SQLSTATEなし",
			errorStr: "some error without sqlstate",
			want:     "UNKNOWN",
		},
		{
			name:     "SQLSTATEが短すぎる",
			errorStr: "SQLSTATE 23",
			want:     "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractSQLState(tt.errorStr); got != tt.want {
				t.Errorf("extractSQLState() = %q, want %q", got, tt.want)
			}
		})
	}
}
