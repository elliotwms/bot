package bot

import (
	"context"
	"log/slog"
)

// dh discards all log output.
// dh.Enabled returns false for all Levels.
// todo remove when slog.DiscardHandler is available in go 1.24 https://github.com/golang/go/issues/62005
var dh slog.Handler = discardHandler{}

type discardHandler struct{}

func (dh discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (dh discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (dh discardHandler) WithAttrs(attrs []slog.Attr) slog.Handler  { return dh }
func (dh discardHandler) WithGroup(name string) slog.Handler        { return dh }

func withErr(err error) slog.Attr {
	return slog.Any("error", err)
}
