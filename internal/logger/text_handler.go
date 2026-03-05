package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

type SimpleHandler struct {
	out   io.Writer
	level slog.Level
	mu    sync.Mutex
}

func NewTextHandler(out io.Writer, level slog.Level) *SimpleHandler {
	return &SimpleHandler{out: out, level: level}
}

func (h *SimpleHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *SimpleHandler) Handle(_ context.Context, r slog.Record) error {
	if !h.Enabled(context.TODO(), r.Level) {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	var b strings.Builder

	// <time>
	b.WriteString(r.Time.Format("2006-01-02T15:04:05-07:00"))
	b.WriteString(" ")

	// <level>
	b.WriteString("[")
	b.WriteString(r.Level.String())
	b.WriteString("] ")

	// <msg>
	b.WriteString(r.Message)
	b.WriteString(" | ")

	// <attrs>
	r.Attrs(func(a slog.Attr) bool {
		b.WriteString(" ")
		b.WriteString(a.Key)
		b.WriteString("=")
		b.WriteString(fmt.Sprint(a.Value.Any()))
		return true
	})

	b.WriteString("\n")

	_, err := h.out.Write([]byte(b.String()))
	return err
}

func (h *SimpleHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *SimpleHandler) WithGroup(_ string) slog.Handler      { return h }
