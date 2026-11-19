// Package telemetry holds the logger, the tracer and the span of a request.
// Only `adapters`, `platform` and the commands create telemetry. Rules:
// Telemetry in CODING_STANDARDS.md.
package telemetry

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// NewLogger returns the logger of a command: one JSON object for each line on
// the output, from the level up. A line written with the context of a request
// carries its trace identifier, as `trace_id`.
//
// What may be logged is listed in Telemetry. No line carries a body, a query
// string, a header or anything sensitive.
func NewLogger(output io.Writer, level slog.Level) *slog.Logger {
	lines := slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level})
	return slog.New(traceHandler{lines: lines})
}

// RouteTheStandardLogger makes the logger of the command the one of the whole
// process: what a library writes with the standard library's logger, or with
// the default logger of slog, becomes a line of it, in JSON, at `info`. Every
// command calls it once it has its logger.
func RouteTheStandardLogger(logger *slog.Logger) {
	slog.SetDefault(logger)
}

// TraceID returns the trace identifier of the request the context belongs
// to, and an empty text outside a request. It is the `trace_id` of a log line
// and of an error body.
func TraceID(ctx context.Context) string {
	span := trace.SpanContextFromContext(ctx)
	if !span.HasTraceID() {
		return ""
	}
	return span.TraceID().String()
}

// traceHandler adds the trace identifier to each line, at the root of the
// object. The attributes and the groups a logger was given are applied to
// each line after it, so a group never swallows the identifier.
type traceHandler struct {
	lines slog.Handler
	given []func(slog.Handler) slog.Handler
}

func (h traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.lines.Enabled(ctx, level)
}

func (h traceHandler) Handle(ctx context.Context, record slog.Record) error {
	lines := h.lines
	if traceID := TraceID(ctx); traceID != "" {
		lines = lines.WithAttrs([]slog.Attr{slog.String("trace_id", traceID)})
	}
	for _, apply := range h.given {
		lines = apply(lines)
	}
	return lines.Handle(ctx, record)
}

func (h traceHandler) WithAttrs(attributes []slog.Attr) slog.Handler {
	return h.with(func(lines slog.Handler) slog.Handler { return lines.WithAttrs(attributes) })
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return h.with(func(lines slog.Handler) slog.Handler { return lines.WithGroup(name) })
}

func (h traceHandler) with(apply func(slog.Handler) slog.Handler) traceHandler {
	given := make([]func(slog.Handler) slog.Handler, len(h.given), len(h.given)+1)
	copy(given, h.given)
	return traceHandler{lines: h.lines, given: append(given, apply)}
}
