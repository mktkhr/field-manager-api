// Package logger はアプリケーション全体で使用するログ機能を提供する
package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/mktkhr/field-manager-api/internal/config"
)

// Setup はLoggerConfigに基づいてログ設定を初期化する
// production環境ではJSON形式、それ以外では色付きテキスト形式で出力する
func Setup(cfg config.LoggerConfig) {
	level := parseLogLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level: level,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
