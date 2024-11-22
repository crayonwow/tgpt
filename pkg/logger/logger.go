package logger

import (
	"log/slog"
)

func New(config Config) *slog.Logger {
	lvl := slog.Level(0)
	if err := lvl.UnmarshalText([]byte(config.Level)); err != nil {
		lvl = slog.LevelError
	}

	ops := &slog.HandlerOptions{
		AddSource: true,
		Level:     lvl,
	}

	var handler slog.Handler
	switch config.HandlerType {
	case HandlerTypeJSON:
		handler = slog.NewJSONHandler(config.Out, ops)
	default:
		handler = slog.NewTextHandler(config.Out, ops)
	}

	return slog.New(handler)
}

func NewDefault(config Config) {
	slog.SetDefault(New(config))
}
