package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/sergeyWh1te/go-template/internal/env"
)

// New builds the application logger. An unset or unrecognized LOG_LEVEL falls
// back to info rather than failing startup — losing the service over a typo in
// a log setting is worse than logging at the wrong level.
func New(cfg *env.AppConfig) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}

	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if strings.EqualFold(cfg.LogFormat, "json") {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func parseLevel(s string) slog.Level {
	var level slog.Level
	// UnmarshalText accepts "debug"/"INFO"/"warn"/"error" and also offsets such
	// as "error+2"; anything else keeps the info default.
	if err := level.UnmarshalText([]byte(strings.TrimSpace(s))); err != nil {
		return slog.LevelInfo
	}

	return level
}
