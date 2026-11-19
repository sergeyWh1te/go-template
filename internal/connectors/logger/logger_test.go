package logger

import (
	"log/slog"
	"testing"
)

func TestParseLevel_known_and_unknown_levels(t *testing.T) {
	cases := []struct {
		in   string
		want slog.Level
	}{
		{in: "debug", want: slog.LevelDebug},
		{in: "INFO", want: slog.LevelInfo},
		{in: " warn ", want: slog.LevelWarn},
		{in: "error", want: slog.LevelError},
		{in: "error+2", want: slog.LevelError + 2},
		// Unset or unrecognized levels fall back to info instead of failing startup.
		{in: "", want: slog.LevelInfo},
		{in: "verbose", want: slog.LevelInfo},
	}

	for _, tc := range cases {
		if got := parseLevel(tc.in); got != tc.want {
			t.Errorf("parseLevel(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
