package logger

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"time"
)

// メソッドトレースの開始を記録し、終了時の処理を返す
func TraceMethod(ctx context.Context, methodName string, params ...interface{}) func() {
	start := time.Now()

	// パラメータを構造化ログ用にフォーマット
	logArgs := make([]interface{}, 0, len(params)*2+4)
	logArgs = append(logArgs, "method", methodName)
	logArgs = append(logArgs, "phase", "start")

	// パラメータを追加(nilチェックあり)
	for i, param := range params {
		if param != nil {
			// 構造体や複雑な型は文字列化
			paramValue := formatParameter(param)
			logArgs = append(logArgs, fmt.Sprintf("param_%d", i), paramValue)
		}
	}

	DebugWithContext(ctx, "メソッド実行開始", logArgs...)

	// 終了処理を返す
	return func() {
		duration := time.Since(start)
		DebugWithContext(ctx, "メソッド実行終了",
			"method", methodName,
			"phase", "end",
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String())
	}
}

// 自動でメソッド名を取得してトレース
func TraceMethodAuto(ctx context.Context, params ...interface{}) func() {
	pc, _, _, ok := runtime.Caller(1)
	methodName := "unknown"
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			methodName = fn.Name()
		}
	}
	return TraceMethod(ctx, methodName, params...)
}

// パラメータを適切な形式でフォーマット
func formatParameter(param interface{}) interface{} {
	if param == nil {
		return nil
	}

	v := reflect.ValueOf(param)
	t := reflect.TypeOf(param)

	// ポインタの場合は中身を取得
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	// プリミティブ型はそのまま返す
	switch t.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return v.Interface()
	case reflect.Struct:
		// 構造体の場合、主要フィールドのみ表示
		result := make(map[string]interface{})

		// よく使われるフィールド名を優先的に取得
		fieldNames := []string{"ID", "Id", "Name", "Title", "Code", "Type", "Status"}
		for _, fieldName := range fieldNames {
			if field := v.FieldByName(fieldName); field.IsValid() && field.CanInterface() {
				result[fieldName] = field.Interface()
			}
		}

		// 主要フィールドが見つからない場合は構造体名のみ
		if len(result) == 0 {
			return fmt.Sprintf("<%s>", t.Name())
		}
		return result
	default:
		// その他の型は型名のみ表示
		return fmt.Sprintf("<%s>", t.String())
	}
}

// エラー発生時のメソッドトレース
func TraceMethodError(ctx context.Context, methodName string, err error, additionalInfo ...interface{}) {
	logArgs := make([]interface{}, 0, len(additionalInfo)+4)
	logArgs = append(logArgs, "method", methodName)
	logArgs = append(logArgs, "phase", "error")
	logArgs = append(logArgs, "error", err.Error())

	// 追加情報を付与
	logArgs = append(logArgs, additionalInfo...)

	ErrorWithContext(ctx, "メソッド実行エラー", logArgs...)
}

// 成功時のメソッド結果ログ
func TraceMethodSuccess(ctx context.Context, methodName string, result interface{}, additionalInfo ...interface{}) {
	logArgs := make([]interface{}, 0, len(additionalInfo)+4)
	logArgs = append(logArgs, "method", methodName)
	logArgs = append(logArgs, "phase", "success")

	// 結果の要約情報を追加
	if result != nil {
		logArgs = append(logArgs, "result", formatParameter(result))
	}

	// 追加情報を付与
	logArgs = append(logArgs, additionalInfo...)

	DebugWithContext(ctx, "メソッド実行成功", logArgs...)
}
