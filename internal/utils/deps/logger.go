package deps

import "log/slog"

type Logger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	With(args ...any) *slog.Logger
}
