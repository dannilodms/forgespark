package s3

import (
	"context"
	"regexp"

	"github.com/aws/smithy-go/tracing"
)

// No key of an object goes into a log or a span: the errors of this adapter
// go into the logs of a request or of a job, and the SDK records on the span
// of each call its raw error and what the storage answered.
//
// A key reaches a text in two ways. A request that could not be sent is
// named by its URL, whose path holds the key of an object and whose query, on
// a later page of the listing, holds a token the storage makes from one
// (Garage encodes the last key in base64). And the storage may repeat a key
// in its answer.

var (
	// pastTheAddress is what follows the scheme and the host of a URL.
	pastTheAddress = regexp.MustCompile(`(https?://[^/\s"?#]+)[^\s"]*`)
	// objectKey is the form of an object key, as domain.NewObjectKey makes
	// one: 128 bits in hexadecimal.
	objectKey = regexp.MustCompile(`\b[0-9a-f]{32}\b`)
)

// withoutKeys cuts each URL of the text to its scheme and its host, and puts
// [key] in place of each object key.
func withoutKeys(text string) string {
	return objectKey.ReplaceAllString(pastTheAddress.ReplaceAllString(text, "$1"), "[key]")
}

// failed is the error of a call to the storage, said as what failed, with no
// key of an object in its text. The error is still what it wraps, for
// errors.Is and errors.As: a call cut by a shutdown is still a context
// cancelled.
func failed(what string, err error) error {
	return keyless{text: what + ": " + withoutKeys(err.Error()), err: err}
}

// keyless is an error of the storage told without a key.
type keyless struct {
	text string
	err  error
}

func (e keyless) Error() string { return e.text }

func (e keyless) Unwrap() error { return e.err }

// keylessTracing is the tracing of the SDK with no key of an object in a
// text a span records.
type keylessTracing struct {
	tracing.TracerProvider
}

func (p keylessTracing) Tracer(scope string, options ...tracing.TracerOption) tracing.Tracer {
	return keylessTracer{Tracer: p.TracerProvider.Tracer(scope, options...)}
}

type keylessTracer struct {
	tracing.Tracer
}

// StartSpan starts the span of the SDK. The context keeps the span of the SDK
// itself, which the tracing of the SDK reads back for the spans under it.
func (t keylessTracer) StartSpan(ctx context.Context, name string, options ...tracing.SpanOption) (context.Context, tracing.Span) {
	ctx, span := t.Tracer.StartSpan(ctx, name, options...)
	return ctx, keylessSpan{Span: span}
}

type keylessSpan struct {
	tracing.Span
}

// SetProperty records the property, a text without its keys.
func (s keylessSpan) SetProperty(key, value any) {
	text, isText := value.(string)
	if !isText {
		s.Span.SetProperty(key, value)
		return
	}
	s.Span.SetProperty(key, withoutKeys(text))
}
