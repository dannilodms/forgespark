package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"

// Requests wraps the handler of the API with the telemetry of a request. It
// opens the span of each request, under the trace of an incoming
// `traceparent`, and writes the only line a request may leave in the logs:
// the one of a 5xx or of a panic, with its raw error.
//
// A successful request and a 4xx write no line. A span and a line carry the
// method, the route and the status, and never a path, a query string, a
// header or a body.
//
// A request to one of the paths withoutSpan names, liveness and readiness,
// records no span, and nothing under it does, its queries included: a probe
// every few seconds would fill the traces. The tracer of StartTracing drops
// them. Its line, when it fails, is still written, under the trace of an
// incoming `traceparent`.
//
// A panic that reaches this handler answers 500 with no body.
func Requests(next http.Handler, logger *slog.Logger, tracerProvider trace.TracerProvider, withoutSpan ...string) http.Handler {
	return requests{next: next, logger: logger, tracer: tracerProvider.Tracer(instrumentationName), withoutSpan: withoutSpan}
}

// RecordFailure keeps the raw error behind the 5xx a request is about to
// answer, for the line Requests writes when the request ends. The answer
// never carries it. Outside a request, and for an answer below 500, it does
// nothing.
func RecordFailure(ctx context.Context, err error) {
	kept, inRequest := ctx.Value(failureKey{}).(*failure)
	if !inRequest {
		return
	}
	kept.err = err
}

type failureKey struct{}

// failure is the raw error of one request. Only the goroutine of the request
// writes it, because no goroutine outlives its request.
type failure struct {
	err error
}

// noSpanKey marks the context of a request that records no span.
type noSpanKey struct{}

// recordsNoSpan tells whether a span started in the context belongs to a
// request that records none.
func recordsNoSpan(ctx context.Context) bool {
	return ctx.Value(noSpanKey{}) != nil
}

type requests struct {
	next        http.Handler
	logger      *slog.Logger
	tracer      trace.Tracer
	withoutSpan []string
}

func (t requests) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	if slices.Contains(t.withoutSpan, r.URL.Path) {
		ctx = context.WithValue(ctx, noSpanKey{}, true)
	}
	ctx, span := t.tracer.Start(ctx, methodOf(r), trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	kept := &failure{}
	answer := &response{ResponseWriter: w}
	// The mux writes the route it matched on this copy of the request.
	request := r.WithContext(context.WithValue(ctx, failureKey{}, kept))
	defer t.record(span, request, answer, kept)

	t.next.ServeHTTP(answer, request)
}

// record ends the telemetry of one request. It is deferred, so it also sees
// a panic of the handler.
func (t requests) record(span trace.Span, request *http.Request, answer *response, kept *failure) {
	panicked := recover()
	if panicked == http.ErrAbortHandler {
		// The handler asked the server to drop the connection.
		panic(panicked)
	}
	if panicked != nil {
		answer.writeHeaderOnce(http.StatusInternalServerError)
	}

	method, route, status := methodOf(request), routeOf(request), answer.sent()
	span.SetAttributes(semconv.HTTPRequestMethodKey.String(method), semconv.HTTPResponseStatusCode(status))
	if route != "" {
		span.SetName(method + " " + route)
		span.SetAttributes(semconv.HTTPRoute(route))
	}
	if panicked == nil && status < http.StatusInternalServerError {
		return
	}

	span.SetStatus(codes.Error, "")
	line := []slog.Attr{slog.String("method", method), slog.String("route", route), slog.Int("status", status)}
	if panicked != nil {
		line = append(line, slog.String("error", fmt.Sprint(panicked)), slog.String("stack", string(debug.Stack())))
		t.logger.LogAttrs(request.Context(), slog.LevelError, "request panicked", line...)
		return
	}
	if kept.err != nil {
		line = append(line, slog.String("error", kept.err.Error()))
	}
	t.logger.LogAttrs(request.Context(), slog.LevelError, "request failed", line...)
}

// knownMethods are the methods a span and a line name. Any other one is
// written as `_OTHER`, as the semantic conventions of OpenTelemetry say, so
// that nothing a caller invents reaches the telemetry.
var knownMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace,
}

func methodOf(request *http.Request) string {
	if !slices.Contains(knownMethods, request.Method) {
		return "_OTHER"
	}
	return request.Method
}

// routeOf gives the route the mux matched, without its method: `/things/{id}`
// for the pattern `GET /things/{id}`. It is empty when no route matched.
func routeOf(request *http.Request) string {
	_, route, hasMethod := strings.Cut(request.Pattern, " ")
	if !hasMethod {
		return request.Pattern
	}
	return route
}

// response remembers the status a handler answered with.
type response struct {
	http.ResponseWriter
	status int
}

func (r *response) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *response) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

// Unwrap lets http.ResponseController reach the writer of the server, which
// a stream needs to flush.
func (r *response) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *response) writeHeaderOnce(status int) {
	if r.status != 0 {
		return
	}
	r.WriteHeader(status)
}

// sent gives the status of the answer. A handler that wrote nothing answered
// 200.
func (r *response) sent() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}
