package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

// LogDatabaseError はデータベース操作エラー用のログを出力する
func LogDatabaseError(ctx context.Context, operation string, table string, err error, additionalInfo ...interface{}) {
	callerMethod := getCallerMethod(2)

	logArgs := make([]interface{}, 0, len(additionalInfo)+8)
	logArgs = append(logArgs,
		"error_type", "database",
		"operation", operation,
		"table", table,
		"method", callerMethod,
		"error", err.Error())

	if dbErrorInfo := extractDatabaseErrorInfo(err); dbErrorInfo != nil {
		logArgs = append(logArgs, "db_error_code", dbErrorInfo.Code)
		logArgs = append(logArgs, "db_error_detail", dbErrorInfo.Detail)
	}

	logArgs = append(logArgs, additionalInfo...)

	ErrorWithContext(ctx, fmt.Sprintf("%sでデータベースエラーが発生", operation), logArgs...)
}

// LogBusinessError はビジネスロジックエラー用のログを出力する
func LogBusinessError(ctx context.Context, businessRule string, err error, additionalInfo ...interface{}) {
	callerMethod := getCallerMethod(2)

	logArgs := make([]interface{}, 0, len(additionalInfo)+6)
	logArgs = append(logArgs,
		"error_type", "business",
		"business_rule", businessRule,
		"method", callerMethod)

	if err != nil {
		logArgs = append(logArgs, "error", err.Error())
	} else {
		logArgs = append(logArgs, "error", "error is nil")
	}

	logArgs = append(logArgs, additionalInfo...)

	WarnWithContext(ctx, fmt.Sprintf("ビジネスルールエラー: %s", businessRule), logArgs...)
}

// LogValidationError はバリデーションエラー用のログを出力する
func LogValidationError(ctx context.Context, field string, value interface{}, rule string, additionalInfo ...interface{}) {
	callerMethod := getCallerMethod(2)

	logArgs := make([]interface{}, 0, len(additionalInfo)+8)
	logArgs = append(logArgs,
		"error_type", "validation",
		"field", field,
		"value", value,
		"rule", rule,
		"method", callerMethod)

	logArgs = append(logArgs, additionalInfo...)

	WarnWithContext(ctx, fmt.Sprintf("バリデーションエラー: %s", field), logArgs...)
}

// LogExternalAPIError は外部API呼び出しエラー用のログを出力する
func LogExternalAPIError(ctx context.Context, apiName string, endpoint string, statusCode int, err error, additionalInfo ...interface{}) {
	callerMethod := getCallerMethod(2)

	logArgs := make([]interface{}, 0, len(additionalInfo)+10)
	logArgs = append(logArgs,
		"error_type", "external_api",
		"api_name", apiName,
		"endpoint", endpoint,
		"status_code", statusCode,
		"method", callerMethod,
		"error", err.Error())

	logArgs = append(logArgs, additionalInfo...)

	ErrorWithContext(ctx, fmt.Sprintf("外部API呼び出しエラー: %s", apiName), logArgs...)
}

// LogErrorWithContext は汎用エラーログを出力する(コンテキスト情報を自動付与)
func LogErrorWithContext(ctx context.Context, errorType string, message string, err error, additionalInfo ...interface{}) {
	callerMethod := getCallerMethod(2)

	logArgs := make([]interface{}, 0, len(additionalInfo)+6)
	logArgs = append(logArgs,
		"error_type", errorType,
		"method", callerMethod,
		"error", err.Error())

	logArgs = append(logArgs, additionalInfo...)

	ErrorWithContext(ctx, message, logArgs...)
}

func getCallerMethod(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	methodName := fn.Name()

	parts := strings.Split(methodName, "/")
	if len(parts) > 0 {
		methodName = parts[len(parts)-1]
	}

	fileParts := strings.Split(file, "/")
	fileName := fileParts[len(fileParts)-1]

	return fmt.Sprintf("%s (%s:%d)", methodName, fileName, line)
}

// DatabaseErrorInfo はデータベースエラー情報を保持する
type DatabaseErrorInfo struct {
	Code   string
	Detail string
}

func extractDatabaseErrorInfo(err error) *DatabaseErrorInfo {
	if err == nil {
		return nil
	}

	errorStr := err.Error()

	// PostgreSQLエラーコードの抽出
	if strings.Contains(errorStr, "SQLSTATE") {
		return &DatabaseErrorInfo{
			Code:   extractSQLState(errorStr),
			Detail: errorStr,
		}
	}

	// 一般的なデータベースエラー
	if strings.Contains(errorStr, "duplicate") || strings.Contains(errorStr, "unique") {
		return &DatabaseErrorInfo{
			Code:   "DUPLICATE_KEY",
			Detail: errorStr,
		}
	}

	if strings.Contains(errorStr, "foreign key") {
		return &DatabaseErrorInfo{
			Code:   "FOREIGN_KEY_VIOLATION",
			Detail: errorStr,
		}
	}

	return &DatabaseErrorInfo{
		Code:   "UNKNOWN",
		Detail: errorStr,
	}
}

func extractSQLState(errorStr string) string {
	start := strings.Index(errorStr, "SQLSTATE")
	if start == -1 {
		return "UNKNOWN"
	}

	start += 9
	end := start + 5

	if end <= len(errorStr) {
		return errorStr[start:end]
	}

	return "UNKNOWN"
}
