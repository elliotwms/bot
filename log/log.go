package log

import (
	"context"
	"log/slog"
)

// DiscardHandler discards all log output.
// DiscardHandler.Enabled returns false for all Levels.
// todo remove when slog.DiscardHandler is available in go 1.24 https://github.com/golang/go/issues/62005
var DiscardHandler slog.Handler = discardHandler{}

type discardHandler struct{}

func (dh discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (dh discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (dh discardHandler) WithAttrs(attrs []slog.Attr) slog.Handler  { return dh }
func (dh discardHandler) WithGroup(name string) slog.Handler        { return dh }

func WithErr(err error) slog.Attr {
	return slog.Any("error", err)
}
