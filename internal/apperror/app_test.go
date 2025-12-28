package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestAppErrorMethods(t *testing.T) {
	cause := errors.New("original error")
	err := newAppError("TEST_CODE", "test message", http.StatusBadRequest, cause)

	if err.Error() != "test message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test message")
	}
	if err.Code() != "TEST_CODE" {
		t.Errorf("Code() = %q, want %q", err.Code(), "TEST_CODE")
	}
	if err.Message() != "test message" {
		t.Errorf("Message() = %q, want %q", err.Message(), "test message")
	}
	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusBadRequest)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, want %v", err.Cause(), cause)
	}
	if err.StackTrace() == "" {
		t.Error("StackTrace() should not be empty")
	}
}

func TestAppErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	err := newAppError("TEST_CODE", "test message", http.StatusBadRequest, cause)

	appErr, ok := err.(*appError)
	if !ok {
		t.Fatal("expected *appError type")
	}
	if appErr.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", appErr.Unwrap(), cause)
	}

	// errors.Isで動作確認
	if !errors.Is(err, cause) {
		t.Error("errors.Is should return true for wrapped cause")
	}
}

func TestNewAppErrorWithoutCause(t *testing.T) {
	err := newAppErrorWithoutCause("TEST_CODE", "test message", http.StatusOK)

	if err.Code() != "TEST_CODE" {
		t.Errorf("Code() = %q, want %q", err.Code(), "TEST_CODE")
	}
	if err.Message() != "test message" {
		t.Errorf("Message() = %q, want %q", err.Message(), "test message")
	}
	if err.HTTPStatus() != http.StatusOK {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusOK)
	}
	if err.Cause() != nil {
		t.Errorf("Cause() = %v, want nil", err.Cause())
	}
	if err.StackTrace() != "" {
		t.Error("StackTrace() should be empty for error without cause")
	}
}

func TestNotFoundError(t *testing.T) {
	err := NotFoundError("resource not found")

	if err.Code() != "NOT_FOUND" {
		t.Errorf("Code() = %q, want %q", err.Code(), "NOT_FOUND")
	}
	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusNotFound)
	}
}

func TestBadRequestError(t *testing.T) {
	err := BadRequestError("invalid input")

	if err.Code() != "BAD_REQUEST" {
		t.Errorf("Code() = %q, want %q", err.Code(), "BAD_REQUEST")
	}
	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusBadRequest)
	}
}

func TestConflictError(t *testing.T) {
	err := ConflictError("conflict occurred")

	if err.Code() != "CONFLICT" {
		t.Errorf("Code() = %q, want %q", err.Code(), "CONFLICT")
	}
	if err.HTTPStatus() != http.StatusConflict {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusConflict)
	}
}

func TestForbiddenError(t *testing.T) {
	err := ForbiddenError("access denied")

	if err.Code() != "FORBIDDEN" {
		t.Errorf("Code() = %q, want %q", err.Code(), "FORBIDDEN")
	}
	if err.HTTPStatus() != http.StatusForbidden {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusForbidden)
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := UnauthorizedError("authentication required")

	if err.Code() != "UNAUTHORIZED" {
		t.Errorf("Code() = %q, want %q", err.Code(), "UNAUTHORIZED")
	}
	if err.HTTPStatus() != http.StatusUnauthorized {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusUnauthorized)
	}
}

func TestInternalError(t *testing.T) {
	err := InternalError("internal server error")

	if err.Code() != "INTERNAL_ERROR" {
		t.Errorf("Code() = %q, want %q", err.Code(), "INTERNAL_ERROR")
	}
	if err.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusInternalServerError)
	}
}

func TestTimeoutError(t *testing.T) {
	err := TimeoutError("request timeout")

	if err.Code() != "TIMEOUT" {
		t.Errorf("Code() = %q, want %q", err.Code(), "TIMEOUT")
	}
	if err.HTTPStatus() != http.StatusGatewayTimeout {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusGatewayTimeout)
	}
}

func TestNotFoundErrorWithCause(t *testing.T) {
	cause := errors.New("db error")
	err := NotFoundErrorWithCause("user not found", cause)

	if err.Code() != "NOT_FOUND" {
		t.Errorf("Code() = %q, want %q", err.Code(), "NOT_FOUND")
	}
	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusNotFound)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, want %v", err.Cause(), cause)
	}
	if err.StackTrace() == "" {
		t.Error("StackTrace() should not be empty")
	}
}

func TestBadRequestErrorWithCause(t *testing.T) {
	cause := errors.New("validation error")
	err := BadRequestErrorWithCause("invalid email", cause)

	if err.Code() != "BAD_REQUEST" {
		t.Errorf("Code() = %q, want %q", err.Code(), "BAD_REQUEST")
	}
	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusBadRequest)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, want %v", err.Cause(), cause)
	}
}

func TestInternalErrorWithCause(t *testing.T) {
	cause := errors.New("db connection failed")
	err := InternalErrorWithCause("database error", cause)

	if err.Code() != "INTERNAL_ERROR" {
		t.Errorf("Code() = %q, want %q", err.Code(), "INTERNAL_ERROR")
	}
	if err.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusInternalServerError)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, want %v", err.Cause(), cause)
	}
}

func TestConflictErrorWithCause(t *testing.T) {
	cause := errors.New("duplicate key")
	err := ConflictErrorWithCause("resource already exists", cause)

	if err.Code() != "CONFLICT" {
		t.Errorf("Code() = %q, want %q", err.Code(), "CONFLICT")
	}
	if err.HTTPStatus() != http.StatusConflict {
		t.Errorf("HTTPStatus() = %d, want %d", err.HTTPStatus(), http.StatusConflict)
	}
	if err.Cause() != cause {
		t.Errorf("Cause() = %v, want %v", err.Cause(), cause)
	}
}

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
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCaptureStack(t *testing.T) {
	stack := captureStack(1)

	if stack == "" {
		t.Error("captureStack() should not return empty string")
	}
	// スタックトレースにはこの関数名が含まれるはず
	if !containsString(stack, "TestCaptureStack") {
		t.Error("captureStack() should contain the caller function name")
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
