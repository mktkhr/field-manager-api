// Package apperror はアプリケーション全体で使用する共通エラー型を提供する
package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// AppError はアプリケーションエラーを表すインターフェース
// HTTPステータスコード、エラーコード、原因、スタックトレースを保持する
type AppError interface {
	error
	Code() string
	Message() string
	HTTPStatus() int
	Cause() error
	StackTrace() string
}

type appError struct {
	code       string
	message    string
	httpStatus int
	cause      error
	stack      string
}

func (e *appError) Error() string      { return e.message }
func (e *appError) Code() string       { return e.code }
func (e *appError) Message() string    { return e.message }
func (e *appError) HTTPStatus() int    { return e.httpStatus }
func (e *appError) Cause() error       { return e.cause }
func (e *appError) StackTrace() string { return e.stack }
func (e *appError) Unwrap() error      { return e.cause }

func newAppError(code string, msg string, status int, cause error) AppError {
	return &appError{
		code:       code,
		message:    msg,
		httpStatus: status,
		cause:      cause,
		stack:      captureStack(3),
	}
}

func newAppErrorWithoutCause(code string, msg string, status int) AppError {
	return &appError{
		code:       code,
		message:    msg,
		httpStatus: status,
	}
}

// NotFoundError はリソースが見つからない場合のエラーを生成する(HTTP 404)
func NotFoundError(msg string) AppError {
	return newAppErrorWithoutCause("NOT_FOUND", msg, http.StatusNotFound)
}

// BadRequestError はリクエストが不正な場合のエラーを生成する(HTTP 400)
func BadRequestError(msg string) AppError {
	return newAppErrorWithoutCause("BAD_REQUEST", msg, http.StatusBadRequest)
}

// ConflictError はリソースの競合が発生した場合のエラーを生成する(HTTP 409)
func ConflictError(msg string) AppError {
	return newAppErrorWithoutCause("CONFLICT", msg, http.StatusConflict)
}

// ForbiddenError はアクセスが禁止されている場合のエラーを生成する(HTTP 403)
func ForbiddenError(msg string) AppError {
	return newAppErrorWithoutCause("FORBIDDEN", msg, http.StatusForbidden)
}

// UnauthorizedError は認証が必要な場合のエラーを生成する(HTTP 401)
func UnauthorizedError(msg string) AppError {
	return newAppErrorWithoutCause("UNAUTHORIZED", msg, http.StatusUnauthorized)
}

// InternalError はサーバー内部エラーを生成する(HTTP 500)
func InternalError(msg string) AppError {
	return newAppErrorWithoutCause("INTERNAL_ERROR", msg, http.StatusInternalServerError)
}

// TimeoutError はタイムアウトエラーを生成する(HTTP 504)
func TimeoutError(msg string) AppError {
	return newAppErrorWithoutCause("TIMEOUT", msg, http.StatusGatewayTimeout)
}

// NotFoundErrorWithCause は原因エラー付きのNotFoundエラーを生成する
// スタックトレースも記録される
func NotFoundErrorWithCause(msg string, cause error) AppError {
	return newAppError("NOT_FOUND", msg, http.StatusNotFound, cause)
}

// BadRequestErrorWithCause は原因エラー付きのBadRequestエラーを生成する
// スタックトレースも記録される
func BadRequestErrorWithCause(msg string, cause error) AppError {
	return newAppError("BAD_REQUEST", msg, http.StatusBadRequest, cause)
}

// InternalErrorWithCause は原因エラー付きのInternalエラーを生成する
// スタックトレースも記録される
func InternalErrorWithCause(msg string, cause error) AppError {
	return newAppError("INTERNAL_ERROR", msg, http.StatusInternalServerError, cause)
}

// ConflictErrorWithCause は原因エラー付きのConflictエラーを生成する
// スタックトレースも記録される
func ConflictErrorWithCause(msg string, cause error) AppError {
	return newAppError("CONFLICT", msg, http.StatusConflict, cause)
}

// IsNotFoundError は指定されたエラーがNotFoundエラーかどうかを判定する
// errors.Asを使用するため、ラップされたエラーにも対応する
func IsNotFoundError(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr.Code() == "NOT_FOUND"
	}
	return false
}

func captureStack(skip int) string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])

	var sb strings.Builder
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return sb.String()
}
