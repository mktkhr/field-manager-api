package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// TestAppErrorMethods はappErrorの各メソッド(Error, Code, Message, HTTPStatus, Cause, StackTrace)が正しい値を返すことをテストする
func TestAppErrorMethods(t *testing.T) {
	cause := errors.New("original error")
	err := newAppError("TEST_CODE", "test message", http.StatusBadRequest, cause)

	if err.Error() != "test message" {
		t.Errorf("Error() = %q, 期待値 %q", err.Error(), "test message")
	}
	if err.Code() != "TEST_CODE" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "TEST_CODE")
	}
	if err.Message() != "test message" {
		t.Errorf("Message() = %q, 期待値 %q", err.Message(), "test message")
	}
	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusBadRequest)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, 期待値 %v", err.Cause(), cause)
	}
	if err.StackTrace() == "" {
		t.Error("StackTrace()が空です")
	}
}

// TestAppErrorUnwrap はappErrorのUnwrapメソッドとerrors.Isによるエラーチェーンが正しく動作することをテストする
func TestAppErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	err := newAppError("TEST_CODE", "test message", http.StatusBadRequest, cause)

	appErr, ok := err.(*appError)
	if !ok {
		t.Fatal("*appError型を期待したが異なる型が返された")
	}
	if appErr.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, 期待値 %v", appErr.Unwrap(), cause)
	}

	// errors.Isで動作確認
	if !errors.Is(err, cause) {
		t.Error("errors.Isはラップされた原因に対してtrueを返すべき")
	}
}

// TestNewAppErrorWithoutCause は原因なしのエラー作成時にCauseがnilでStackTraceが空であることをテストする
func TestNewAppErrorWithoutCause(t *testing.T) {
	err := newAppErrorWithoutCause("TEST_CODE", "test message", http.StatusOK)

	if err.Code() != "TEST_CODE" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "TEST_CODE")
	}
	if err.Message() != "test message" {
		t.Errorf("Message() = %q, 期待値 %q", err.Message(), "test message")
	}
	if err.HTTPStatus() != http.StatusOK {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusOK)
	}
	if err.Cause() != nil {
		t.Errorf("Cause() = %v, 期待値 nil", err.Cause())
	}
	if err.StackTrace() != "" {
		t.Error("原因なしのエラーではStackTrace()が空であるべき")
	}
}

// TestNotFoundError はNotFoundErrorが正しいコードとHTTPステータスを返すことをテストする
func TestNotFoundError(t *testing.T) {
	err := NotFoundError("resource not found")

	if err.Code() != "NOT_FOUND" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "NOT_FOUND")
	}
	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusNotFound)
	}
}

// TestBadRequestError はBadRequestErrorが正しいコードとHTTPステータスを返すことをテストする
func TestBadRequestError(t *testing.T) {
	err := BadRequestError("invalid input")

	if err.Code() != "BAD_REQUEST" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "BAD_REQUEST")
	}
	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusBadRequest)
	}
}

// TestConflictError はConflictErrorが正しいコードとHTTPステータスを返すことをテストする
func TestConflictError(t *testing.T) {
	err := ConflictError("conflict occurred")

	if err.Code() != "CONFLICT" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "CONFLICT")
	}
	if err.HTTPStatus() != http.StatusConflict {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusConflict)
	}
}

// TestForbiddenError はForbiddenErrorが正しいコードとHTTPステータスを返すことをテストする
func TestForbiddenError(t *testing.T) {
	err := ForbiddenError("access denied")

	if err.Code() != "FORBIDDEN" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "FORBIDDEN")
	}
	if err.HTTPStatus() != http.StatusForbidden {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusForbidden)
	}
}

// TestUnauthorizedError はUnauthorizedErrorが正しいコードとHTTPステータスを返すことをテストする
func TestUnauthorizedError(t *testing.T) {
	err := UnauthorizedError("authentication required")

	if err.Code() != "UNAUTHORIZED" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "UNAUTHORIZED")
	}
	if err.HTTPStatus() != http.StatusUnauthorized {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusUnauthorized)
	}
}

// TestInternalError はInternalErrorが正しいコードとHTTPステータスを返すことをテストする
func TestInternalError(t *testing.T) {
	err := InternalError("internal server error")

	if err.Code() != "INTERNAL_ERROR" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "INTERNAL_ERROR")
	}
	if err.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusInternalServerError)
	}
}

// TestTimeoutError はTimeoutErrorが正しいコードとHTTPステータスを返すことをテストする
func TestTimeoutError(t *testing.T) {
	err := TimeoutError("request timeout")

	if err.Code() != "TIMEOUT" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "TIMEOUT")
	}
	if err.HTTPStatus() != http.StatusGatewayTimeout {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusGatewayTimeout)
	}
}

// TestNotFoundErrorWithCause はNotFoundErrorWithCauseが原因エラーとスタックトレースを正しく保持することをテストする
func TestNotFoundErrorWithCause(t *testing.T) {
	cause := errors.New("db error")
	err := NotFoundErrorWithCause("user not found", cause)

	if err.Code() != "NOT_FOUND" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "NOT_FOUND")
	}
	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusNotFound)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, 期待値 %v", err.Cause(), cause)
	}
	if err.StackTrace() == "" {
		t.Error("StackTrace()が空です")
	}
}

// TestBadRequestErrorWithCause はBadRequestErrorWithCauseが原因エラーを正しく保持することをテストする
func TestBadRequestErrorWithCause(t *testing.T) {
	cause := errors.New("validation error")
	err := BadRequestErrorWithCause("invalid email", cause)

	if err.Code() != "BAD_REQUEST" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "BAD_REQUEST")
	}
	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusBadRequest)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, 期待値 %v", err.Cause(), cause)
	}
}

// TestInternalErrorWithCause はInternalErrorWithCauseが原因エラーを正しく保持することをテストする
func TestInternalErrorWithCause(t *testing.T) {
	cause := errors.New("db connection failed")
	err := InternalErrorWithCause("database error", cause)

	if err.Code() != "INTERNAL_ERROR" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "INTERNAL_ERROR")
	}
	if err.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusInternalServerError)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, 期待値 %v", err.Cause(), cause)
	}
}

// TestConflictErrorWithCause はConflictErrorWithCauseが原因エラーを正しく保持することをテストする
func TestConflictErrorWithCause(t *testing.T) {
	cause := errors.New("duplicate key")
	err := ConflictErrorWithCause("resource already exists", cause)

	if err.Code() != "CONFLICT" {
		t.Errorf("Code() = %q, 期待値 %q", err.Code(), "CONFLICT")
	}
	if err.HTTPStatus() != http.StatusConflict {
		t.Errorf("HTTPStatus() = %d, 期待値 %d", err.HTTPStatus(), http.StatusConflict)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, 期待値 %v", err.Cause(), cause)
	}
}

// TestIsNotFoundError はIsNotFoundErrorが各種エラータイプに対して正しい判定結果を返すことをテストする
func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "NotFoundError returns true",
			err:  NotFoundError("not found"),
			want: true,
		},
		{
			name: "NotFoundErrorWithCause returns true",
			err:  NotFoundErrorWithCause("not found", errors.New("cause")),
			want: true,
		},
		{
			name: "wrapped NotFoundError returns true",
			err:  fmt.Errorf("wrapped: %w", NotFoundError("not found")),
			want: true,
		},
		{
			name: "BadRequestError returns false",
			err:  BadRequestError("bad request"),
			want: false,
		},
		{
			name: "standard error returns false",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "nil returns false",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsNotFoundError() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestCaptureStack はcaptureStackが呼び出し元の関数名を含むスタックトレースを返すことをテストする
func TestCaptureStack(t *testing.T) {
	stack := captureStack(1)

	if stack == "" {
		t.Error("captureStack()が空文字列を返しました")
	}
	// スタックトレースにはこの関数名が含まれるはず
	if !containsString(stack, "TestCaptureStack") {
		t.Error("captureStack()は呼び出し元の関数名を含むべき")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
