package logger

import (
	"log/slog"
	"testing"

	"github.com/mktkhr/field-manager-api/internal/config"
)

func TestSetupDevelopment(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:       "INFO",
		Environment: "development",
	}

	Setup(cfg)

	// slog.Defaultが設定されていることを確認
	if slog.Default() == nil {
		t.Error("slog.Default() should not be nil after Setup")
	}
}

func TestSetupProduction(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:       "INFO",
		Environment: "production",
	}

	Setup(cfg)

	if slog.Default() == nil {
		t.Error("slog.Default() should not be nil after Setup")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  slog.Level
	}{
		{
			name:  "DEBUG",
			level: "DEBUG",
			want:  slog.LevelDebug,
		},
		{
			name:  "INFO",
			level: "INFO",
			want:  slog.LevelInfo,
		},
		{
			name:  "WARN",
			level: "WARN",
			want:  slog.LevelWarn,
		},
		{
			name:  "ERROR",
			level: "ERROR",
			want:  slog.LevelError,
		},
		{
			name:  "unknown defaults to INFO",
			level: "UNKNOWN",
			want:  slog.LevelInfo,
		},
		{
			name:  "empty defaults to INFO",
			level: "",
			want:  slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLogLevel(tt.level); got != tt.want {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}
