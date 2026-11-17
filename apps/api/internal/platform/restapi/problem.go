package restapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/validation"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
)

// problemContentType is the media type of every failure (RFC 9457).
const problemContentType = "application/problem+json"

// validationFailed is the code of a request that does not follow the
// contract. Every other failure that is not a broken rule takes its code
// from the name of its status.
const validationFailed = "validation_failed"

// The codes of a field that does not follow the contract: one that is
// required and missing, and one that is there and is not accepted.
const (
	required = "required"
	invalid  = "invalid"
)

// Problem is the body of every failure: RFC 9457 with the stable code, its
// parameters and the trace identifier of the request. `type` is left out,
// which the RFC reads as `about:blank`.
type Problem struct {
	Title   string         `json:"title" doc:"The name of the status."`
	Status  int            `json:"status" doc:"The status of the answer."`
	Detail  string         `json:"detail,omitempty" doc:"What went wrong, in English. A client shows a text of its own, chosen by the code."`
	Code    string         `json:"code" doc:"The stable code of the failure, which a client translates."`
	Params  map[string]any `json:"params" doc:"The values the text of the code is filled with."`
	TraceID string         `json:"trace_id" doc:"The trace identifier of the request, to find it in the logs."`
	Errors  []FieldError   `json:"errors,omitempty" nullable:"false" doc:"The fields that were refused, when the failure is about fields."`
}

// FieldError is one field of a request that was refused.
type FieldError struct {
	Location string         `json:"location" doc:"Where the field is: body.name, query.limit, path.id."`
	Code     string         `json:"code" doc:"The stable code of what is wrong with the field."`
	Params   map[string]any `json:"params" doc:"The values the text of the code is filled with."`
	Message  string         `json:"message" doc:"What is wrong with the field, in English."`
}

// Error makes a Problem the error a handler returns.
func (p *Problem) Error() string {
	return p.Code
}

// GetStatus tells Huma the status to answer with.
func (p *Problem) GetStatus() int {
	return p.Status
}

// ContentType tells Huma the media type of a failure.
func (p *Problem) ContentType(string) string {
	return problemContentType
}

// Huma builds every error it answers with through these two functions, and
// reads the schema of an error from what the first one returns.
func init() {
	huma.NewError = func(status int, message string, causes ...error) huma.StatusError {
		return problemOfHuma(context.Background(), status, message, causes)
	}
	huma.NewErrorWithContext = func(ctx huma.Context, status int, message string, causes ...error) huma.StatusError {
		return problemOfHuma(ctx.Context(), status, message, causes)
	}
}

// problemOfHuma answers a failure Huma found: a request that does not follow
// the contract, or an error of a handler that is not a broken rule. The
// second is a 5xx: its raw error goes to the logs and never to the answer.
func problemOfHuma(ctx context.Context, status int, message string, causes []error) *Problem {
	if status >= http.StatusInternalServerError {
		telemetry.RecordFailure(ctx, errors.Join(append([]error{errors.New(message)}, causes...)...))
		return newProblem(ctx, status, codeOfStatus(status), "")
	}
	problem := newProblem(ctx, status, codeOfStatus(status), message)
	for _, cause := range causes {
		problem.Errors = appendFieldError(problem.Errors, cause)
	}
	return problem
}

func newProblem(ctx context.Context, status int, code, detail string) *Problem {
	return &Problem{
		Title:   http.StatusText(status),
		Status:  status,
		Detail:  detail,
		Code:    code,
		Params:  map[string]any{},
		TraceID: telemetry.TraceID(ctx),
	}
}

// codeOfStatus names a failure that has no code of its own after its status:
// `not_found`, `method_not_allowed`, `too_many_requests`.
func codeOfStatus(status int) string {
	if status == http.StatusUnprocessableEntity {
		return validationFailed
	}
	return strings.ReplaceAll(strings.ToLower(http.StatusText(status)), " ", "_")
}

// appendFieldError adds the field a cause names. The value that was refused
// is never repeated, and a cause that names no field adds nothing.
func appendFieldError(fields []FieldError, cause error) []FieldError {
	detailer, namesField := cause.(huma.ErrorDetailer)
	if !namesField {
		return fields
	}
	detail := detailer.ErrorDetail()
	field := FieldError{Location: detail.Location, Code: invalid, Params: map[string]any{}, Message: detail.Message}
	// Huma places a required property that is missing at the object that
	// lacks it, and names the property in its message alone.
	var missing string
	if _, err := fmt.Sscanf(detail.Message, validation.MsgExpectedRequiredProperty, &missing); err == nil {
		field.Location += "." + missing
		field.Code = required
	}
	return append(fields, field)
}

// Failure is a broken rule, as `domain` and `application` declare one: an
// error with a stable code. Its text is in English and safe to answer with.
// One that has parameters also has `Params() map[string]any`.
type Failure interface {
	error
	Code() string
}

// Failures lists, for each status, the broken rules an operation answers
// with. An example of each error is enough: its code is read from it.
type Failures map[int][]Failure

// answerTo turns the error of a use case into the answer. A broken rule the
// operation declares answers with its status, code and parameters. Anything
// else is left to Huma, which answers 500.
func (f Failures) answerTo(ctx context.Context, err error) error {
	var broken Failure
	if !errors.As(err, &broken) {
		return err
	}
	status, declared := f.statusOf(broken.Code())
	if !declared {
		return fmt.Errorf("the operation does not declare the code %s: %w", broken.Code(), err)
	}
	problem := newProblem(ctx, status, broken.Code(), broken.Error())
	if parameterised, has := broken.(interface{ Params() map[string]any }); has {
		problem.Params = parameterised.Params()
	}
	return problem
}

func (f Failures) statusOf(code string) (int, bool) {
	statuses := f.statuses()
	position := slices.IndexFunc(statuses, func(status int) bool { return containsCode(f[status], code) })
	if position < 0 {
		return 0, false
	}
	return statuses[position], true
}
