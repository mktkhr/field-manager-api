package logger

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTraceMethod(t *testing.T) {
	ctx := context.Background()

	t.Run("パラメータなし", func(t *testing.T) {
		done := TraceMethod(ctx, "TestMethod")
		if done == nil {
			t.Error("TraceMethod() should return a non-nil function")
		}
		done()
	})

	t.Run("パラメータあり", func(t *testing.T) {
		done := TraceMethod(ctx, "TestMethod", "param1", 123, nil)
		if done == nil {
			t.Error("TraceMethod() should return a non-nil function")
		}
		done()
	})

	t.Run("リクエストIDありのコンテキスト", func(t *testing.T) {
		ctxWithID := WithRequestID(ctx, "trace-request-id")
		done := TraceMethod(ctxWithID, "TestMethod", "value")
		done()
	})

	t.Run("実行時間の計測", func(t *testing.T) {
		done := TraceMethod(ctx, "SlowMethod")
		time.Sleep(10 * time.Millisecond)
		done()
	})
}

func TestTraceMethodAuto(t *testing.T) {
	ctx := context.Background()

	t.Run("自動メソッド名取得", func(t *testing.T) {
		done := TraceMethodAuto(ctx)
		if done == nil {
			t.Error("TraceMethodAuto() should return a non-nil function")
		}
		done()
	})

	t.Run("パラメータ付き", func(t *testing.T) {
		done := TraceMethodAuto(ctx, "param1", 42)
		done()
	})
}

func TestFormatParameter(t *testing.T) {
	tests := []struct {
		name     string
		param    interface{}
		wantNil  bool
		wantType string
	}{
		{
			name:    "nil",
			param:   nil,
			wantNil: true,
		},
		{
			name:     "string",
			param:    "hello",
			wantType: "string",
		},
		{
			name:     "int",
			param:    123,
			wantType: "int",
		},
		{
			name:     "int8",
			param:    int8(8),
			wantType: "int8",
		},
		{
			name:     "int16",
			param:    int16(16),
			wantType: "int16",
		},
		{
			name:     "int32",
			param:    int32(32),
			wantType: "int32",
		},
		{
			name:     "int64",
			param:    int64(64),
			wantType: "int64",
		},
		{
			name:     "uint",
			param:    uint(1),
			wantType: "uint",
		},
		{
			name:     "uint8",
			param:    uint8(8),
			wantType: "uint8",
		},
		{
			name:     "uint16",
			param:    uint16(16),
			wantType: "uint16",
		},
		{
			name:     "uint32",
			param:    uint32(32),
			wantType: "uint32",
		},
		{
			name:     "uint64",
			param:    uint64(64),
			wantType: "uint64",
		},
		{
			name:     "float32",
			param:    float32(1.5),
			wantType: "float32",
		},
		{
			name:     "float64",
			param:    float64(2.5),
			wantType: "float64",
		},
		{
			name:     "bool",
			param:    true,
			wantType: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatParameter(tt.param)
			if tt.wantNil {
				if result != nil {
					t.Errorf("formatParameter() = %v, want nil", result)
				}
				return
			}
			if result == nil {
				t.Errorf("formatParameter() = nil, want non-nil")
			}
		})
	}
}

func TestFormatParameterPointer(t *testing.T) {
	t.Run("nilポインタ", func(t *testing.T) {
		var p *string = nil
		result := formatParameter(p)
		if result != nil {
			t.Errorf("formatParameter(nil pointer) = %v, want nil", result)
		}
	})

	t.Run("stringポインタ", func(t *testing.T) {
		s := "hello"
		result := formatParameter(&s)
		if result != "hello" {
			t.Errorf("formatParameter(*string) = %v, want 'hello'", result)
		}
	})

	t.Run("intポインタ", func(t *testing.T) {
		i := 42
		result := formatParameter(&i)
		if result != 42 {
			t.Errorf("formatParameter(*int) = %v, want 42", result)
		}
	})
}

type testStructWithID struct {
	ID   int
	Name string
}

type testStructWithTitle struct {
	Title  string
	Status string
}

type testStructWithCode struct {
	Code string
	Type string
}

type testStructEmpty struct {
	Data []byte
}

func TestFormatParameterStruct(t *testing.T) {
	t.Run("IDフィールドを持つ構造体", func(t *testing.T) {
		s := testStructWithID{ID: 1, Name: "test"}
		result := formatParameter(s)
		if result == nil {
			t.Error("formatParameter(struct) should not return nil")
		}
		if m, ok := result.(map[string]interface{}); ok {
			if m["ID"] != 1 || m["Name"] != "test" {
				t.Errorf("formatParameter(struct) = %v, want ID=1, Name=test", m)
			}
		}
	})

	t.Run("Titleフィールドを持つ構造体", func(t *testing.T) {
		s := testStructWithTitle{Title: "My Title", Status: "active"}
		result := formatParameter(s)
		if m, ok := result.(map[string]interface{}); ok {
			if m["Title"] != "My Title" || m["Status"] != "active" {
				t.Errorf("formatParameter(struct) = %v", m)
			}
		}
	})

	t.Run("Codeフィールドを持つ構造体", func(t *testing.T) {
		s := testStructWithCode{Code: "ABC", Type: "test"}
		result := formatParameter(s)
		if m, ok := result.(map[string]interface{}); ok {
			if m["Code"] != "ABC" || m["Type"] != "test" {
				t.Errorf("formatParameter(struct) = %v", m)
			}
		}
	})

	t.Run("主要フィールドを持たない構造体", func(t *testing.T) {
		s := testStructEmpty{Data: []byte{1, 2, 3}}
		result := formatParameter(s)
		if str, ok := result.(string); ok {
			if str != "<testStructEmpty>" {
				t.Errorf("formatParameter(struct) = %v, want <testStructEmpty>", str)
			}
		} else {
			t.Errorf("formatParameter(struct) should return string for struct without known fields")
		}
	})

	t.Run("構造体ポインタ", func(t *testing.T) {
		s := &testStructWithID{ID: 2, Name: "pointer"}
		result := formatParameter(s)
		if result == nil {
			t.Error("formatParameter(*struct) should not return nil")
		}
	})
}

func TestFormatParameterOtherTypes(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		s := []int{1, 2, 3}
		result := formatParameter(s)
		if str, ok := result.(string); ok {
			if str != "<[]int>" {
				t.Errorf("formatParameter(slice) = %v, want <[]int>", str)
			}
		}
	})

	t.Run("map", func(t *testing.T) {
		m := map[string]int{"a": 1}
		result := formatParameter(m)
		if str, ok := result.(string); ok {
			if str != "<map[string]int>" {
				t.Errorf("formatParameter(map) = %v, want <map[string]int>", str)
			}
		}
	})

	t.Run("channel", func(t *testing.T) {
		ch := make(chan int)
		result := formatParameter(ch)
		if str, ok := result.(string); ok {
			if str != "<chan int>" {
				t.Errorf("formatParameter(chan) = %v, want <chan int>", str)
			}
		}
	})

	t.Run("function", func(t *testing.T) {
		fn := func() {}
		result := formatParameter(fn)
		if result == nil {
			t.Error("formatParameter(func) should not return nil")
		}
	})
}

func TestTraceMethodError(t *testing.T) {
	ctx := context.Background()
	err := errors.New("test error")

	t.Run("追加情報なし", func(t *testing.T) {
		// パニックしないことを確認
		TraceMethodError(ctx, "TestMethod", err)
	})

	t.Run("追加情報あり", func(t *testing.T) {
		TraceMethodError(ctx, "TestMethod", err, "key", "value", "count", 5)
	})

	t.Run("リクエストIDありのコンテキスト", func(t *testing.T) {
		ctxWithID := WithRequestID(ctx, "error-request-id")
		TraceMethodError(ctxWithID, "TestMethod", err)
	})
}

func TestTraceMethodSuccess(t *testing.T) {
	ctx := context.Background()

	t.Run("結果なし", func(t *testing.T) {
		TraceMethodSuccess(ctx, "TestMethod", nil)
	})

	t.Run("プリミティブ結果", func(t *testing.T) {
		TraceMethodSuccess(ctx, "TestMethod", "success result")
	})

	t.Run("構造体結果", func(t *testing.T) {
		result := testStructWithID{ID: 1, Name: "result"}
		TraceMethodSuccess(ctx, "TestMethod", result)
	})

	t.Run("追加情報あり", func(t *testing.T) {
		TraceMethodSuccess(ctx, "TestMethod", "result", "extra", "info")
	})

	t.Run("リクエストIDありのコンテキスト", func(t *testing.T) {
		ctxWithID := WithRequestID(ctx, "success-request-id")
		TraceMethodSuccess(ctxWithID, "TestMethod", "result")
	})
}
