package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

// flushDeadline is how long the last spans have to leave when a command
// ends.
const flushDeadline = 5 * time.Second

// Tracing creates the spans of one command and, when it has an address,
// exports them.
type Tracing struct {
	// Provider goes to whatever opens a span: Requests, the pool of the
	// database, an HTTP client.
	Provider trace.TracerProvider

	exports  bool
	shutdown func(context.Context) error
}

// StartTracing builds the tracer of a command. Every trace is kept: nothing
// is sampled, whatever the caller of a request decided. The one exception is
// a request that Requests marks as recording no span, a probe, whose spans
// are all dropped.
//
// The address is the base URL of an OTLP/HTTP receiver, such as
// http://collector:4318. With no address nothing is exported, and a span
// still has its trace identifier. A failed export is logged and never stops
// the command.
func StartTracing(service, exportURL string, logger *slog.Logger) (*Tracing, error) {
	options := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(everySpanButTheProbes{}),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(service))),
	}
	if exportURL == "" {
		return started(false, options), nil
	}

	traces, err := url.JoinPath(exportURL, "v1", "traces")
	if err != nil {
		return nil, fmt.Errorf("the address telemetry is exported to is not valid: %w", err)
	}
	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(traces))
	if err != nil {
		return nil, fmt.Errorf("build the exporter of traces: %w", err)
	}
	return started(true, append(options, sdktrace.WithBatcher(loggedExporter{SpanExporter: exporter, logger: logger}))), nil
}

func started(exports bool, options []sdktrace.TracerProviderOption) *Tracing {
	provider := sdktrace.NewTracerProvider(options...)
	return &Tracing{Provider: provider, exports: exports, shutdown: provider.Shutdown}
}

// Exports tells whether the spans leave the process. They do not when the
// configuration holds no address.
func (t *Tracing) Exports() bool {
	return t.exports
}

// Stop sends the spans still waiting, under a short deadline of its own, and
// ends the tracer. A command calls it last.
func (t *Tracing) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), flushDeadline)
	defer cancel()
	// A failure to export was already logged, and nothing is left to stop.
	_ = t.shutdown(ctx)
}

// loggedExporter writes a failed export to the logs, as a failed outbound
// call, instead of handing it to the global error handler of OpenTelemetry,
// which prints plain text.
type loggedExporter struct {
	sdktrace.SpanExporter
	logger *slog.Logger
}

func (e loggedExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if err := e.SpanExporter.ExportSpans(ctx, spans); err != nil {
		e.logger.Error("export of traces failed", slog.String("error", err.Error()))
	}
	return nil
}

// everySpanButTheProbes keeps every span, except one started under a request
// that records none, such as liveness and readiness (see Requests). A span it
// drops still carries the trace of its parent, so a line names the trace of
// the caller, and every span under it is dropped too.
type everySpanButTheProbes struct{}

func (everySpanButTheProbes) ShouldSample(parameters sdktrace.SamplingParameters) sdktrace.SamplingResult {
	if recordsNoSpan(parameters.ParentContext) {
		return sdktrace.SamplingResult{
			Decision:   sdktrace.Drop,
			Tracestate: trace.SpanContextFromContext(parameters.ParentContext).TraceState(),
		}
	}
	return sdktrace.AlwaysSample().ShouldSample(parameters)
}

func (everySpanButTheProbes) Description() string {
	return "EverySpanButTheProbes"
}
