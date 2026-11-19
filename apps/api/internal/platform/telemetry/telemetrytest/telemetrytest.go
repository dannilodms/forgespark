// Package telemetrytest gives a test a stand-in for the telemetry collector,
// so that it can read the spans a command exported. Tests run with no
// telemetry otherwise: a command given no address exports nothing.
package telemetrytest

import (
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	collectorpb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Collector receives traces over OTLP/HTTP and keeps them. Its URL is the
// value of OTEL_EXPORTER_OTLP_ENDPOINT for the command under test. A command
// exports its last spans when it stops, so read them after that.
type Collector struct {
	// URL is the base address of the collector.
	URL string

	mu       sync.Mutex
	paths    []string
	received []*collectorpb.ExportTraceServiceRequest
}

// Span is one span that reached the collector.
type Span struct {
	// Service is the name of the service that sent the span.
	Service string
	Name    string
	// TraceID, SpanID and ParentSpanID are in hexadecimal, as a `traceparent`
	// header and a log line write them. ParentSpanID is empty for a span with
	// no parent.
	TraceID      string
	SpanID       string
	ParentSpanID string
	// Attributes holds the attributes whose value is a text.
	Attributes map[string]string
	// Links are the spans this span points at without descending from them,
	// as the span that runs a job points at the one that saved it.
	Links []SpanLink
}

// SpanLink names a span another span points at, in hexadecimal.
type SpanLink struct {
	TraceID string
	SpanID  string
}

// NewCollector starts a collector that lives until the test ends and answers
// every export with 200.
func NewCollector(t testing.TB) *Collector {
	t.Helper()
	return NewCollectorAnswering(t, http.StatusOK)
}

// NewCollectorAnswering starts a collector that answers every export with
// the status, to try a failed export.
func NewCollectorAnswering(t testing.TB, status int) *Collector {
	t.Helper()
	collector := &Collector{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collector.keep(t, r)
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	collector.URL = server.URL
	return collector
}

func (c *Collector) keep(t testing.TB, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("read an export of traces: %v", err)
		return
	}
	var export collectorpb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &export); err != nil {
		t.Errorf("an export of traces is not OTLP: %v", err)
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paths = append(c.paths, r.URL.Path)
	c.received = append(c.received, &export)
}

// Paths lists the path of each export received.
func (c *Collector) Paths() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string{}, c.paths...)
}

// Spans lists every span received so far.
func (c *Collector) Spans() []Span {
	c.mu.Lock()
	defer c.mu.Unlock()
	var spans []Span
	for _, export := range c.received {
		spans = append(spans, spansOfExport(export)...)
	}
	return spans
}

// Text writes everything received so far as text, names and values of every
// attribute, event and status included. A test searches it for a value that
// must never reach a span.
func (c *Collector) Text() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var text string
	for _, export := range c.received {
		text += prototext.Format(export)
	}
	return text
}

func spansOfExport(export *collectorpb.ExportTraceServiceRequest) []Span {
	var spans []Span
	for _, resource := range export.GetResourceSpans() {
		spans = append(spans, spansOfResource(resource)...)
	}
	return spans
}

func spansOfResource(resource *tracepb.ResourceSpans) []Span {
	service := textAttributes(resource.GetResource().GetAttributes())["service.name"]
	var spans []Span
	for _, scope := range resource.GetScopeSpans() {
		spans = append(spans, spansOfScope(service, scope)...)
	}
	return spans
}

func spansOfScope(service string, scope *tracepb.ScopeSpans) []Span {
	var spans []Span
	for _, span := range scope.GetSpans() {
		spans = append(spans, Span{
			Service:      service,
			Name:         span.GetName(),
			TraceID:      hex.EncodeToString(span.GetTraceId()),
			SpanID:       hex.EncodeToString(span.GetSpanId()),
			ParentSpanID: hex.EncodeToString(span.GetParentSpanId()),
			Attributes:   textAttributes(span.GetAttributes()),
			Links:        linksOf(span.GetLinks()),
		})
	}
	return spans
}

func linksOf(links []*tracepb.Span_Link) []SpanLink {
	var read []SpanLink
	for _, link := range links {
		read = append(read, SpanLink{
			TraceID: hex.EncodeToString(link.GetTraceId()),
			SpanID:  hex.EncodeToString(link.GetSpanId()),
		})
	}
	return read
}

func textAttributes(attributes []*commonpb.KeyValue) map[string]string {
	texts := map[string]string{}
	for _, attribute := range attributes {
		texts[attribute.GetKey()] = attribute.GetValue().GetStringValue()
	}
	return texts
}
