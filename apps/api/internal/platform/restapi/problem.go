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

	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
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
		telemetry.RecordFailure(ctx, rawError(message, causes))
		return newProblem(ctx, status, codeOfStatus(status), "")
	}
	problem := newProblem(ctx, status, codeOfStatus(status), message)
	for _, cause := range causes {
		problem.Errors = appendFieldError(problem.Errors, cause)
	}
	return problem
}

// validationFailedDetail is the detail of a request that does not follow the
// contract, as Huma words it, and of one whose fields break a rule.
const validationFailedDetail = "validation failed"

// emailRefused begins the message Huma gives an e-mail its parser could not
// read, followed by the error of the parser, which may quote the address.
var emailRefused = strings.TrimSuffix(validation.MsgExpectedRFC5322Email, "%v")

// rawError is what the logs keep of a 5xx: what Huma says happened, and the
// errors behind it.
func rawError(message string, causes []error) error {
	behind := errors.Join(causes...)
	if behind == nil {
		return errors.New(message)
	}
	return fmt.Errorf("%s: %w", message, behind)
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
// is never repeated, and a cause that names no field adds nothing. A field is
// named once, for the first rule it breaks: Huma checks the length of a text
// before its pattern and its format.
func appendFieldError(fields []FieldError, cause error) []FieldError {
	detailer, namesField := cause.(huma.ErrorDetailer)
	if !namesField {
		return fields
	}
	field := fieldErrorOf(detailer.ErrorDetail())
	named := slices.ContainsFunc(fields, func(each FieldError) bool { return each.Location == field.Location })
	if named {
		return fields
	}
	return append(fields, field)
}

func fieldErrorOf(detail *huma.ErrorDetail) FieldError {
	field := FieldError{Location: detail.Location, Code: invalid, Params: map[string]any{}, Message: detail.Message}
	// Huma places a required property that is missing at the object that
	// lacks it, and names the property in its message alone.
	var missing string
	if _, err := fmt.Sscanf(detail.Message, validation.MsgExpectedRequiredProperty, &missing); err == nil {
		field.Location += "." + missing
		field.Code = required
	}
	// A text the contract asks at least one character of is a required field
	// left empty.
	if detail.Message == fmt.Sprintf(validation.MsgExpectedMinLength, 1) {
		field.Code = required
	}
	if strings.HasPrefix(detail.Message, emailRefused) {
		field.Message = validation.MsgExpectedRFC5322EmailBare
	}
	return field
}

// Failures lists, for each status, the broken rules an operation answers
// with, as `domain` and `application` declare them (kernel.BrokenRule). An
// example of each error is enough: its code is read from it. A rule about one
// field is listed under 422 with InField.
type Failures map[int][]kernel.BrokenRule

// InField declares a rule that a field of the request may break, which the
// use case returns inside a kernel.InvalidFieldsError. It is answered as a
// failed validation, with an entry in `errors` that names the field and
// carries the code of the rule. List it under 422.
func InField(rule kernel.BrokenRule) kernel.BrokenRule {
	return fieldRule{BrokenRule: rule}
}

// LimitedByCaddy stands for the limit Caddy sets in front of an operation for
// each IP address (Rate limit in CODING_STANDARDS.md). Caddy answers 429 with
// the code `too_many_requests`, in the error format, before the request
// reaches the API. An operation under that limit lists it under 429, so the
// spec and the client generated from it know the answer. No use case returns
// it.
var LimitedByCaddy kernel.BrokenRule = limitedByCaddy{}

type limitedByCaddy struct{}

func (limitedByCaddy) Error() string {
	return "too many requests from this address, answered by Caddy before the API; " +
		"the header Retry-After says how many seconds to wait"
}

func (limitedByCaddy) Code() string {
	return codeOfStatus(http.StatusTooManyRequests)
}

// NotFound is the answer, 404, of an operation whose path names nothing that
// exists, such as a login provider that is not turned on. Its code is the
// one of its status, `not_found`. An operation that answers it lists it under
// 404, and its handler returns it.
var NotFound kernel.BrokenRule = notFound{}

type notFound struct{}

func (notFound) Error() string {
	return "the path names nothing that exists"
}

func (notFound) Code() string {
	return codeOfStatus(http.StatusNotFound)
}

// fieldRule is a rule an operation declares about one field.
type fieldRule struct {
	kernel.BrokenRule
}

// answerTo turns the error of a use case into the answer. Fields that break a
// rule answer as a failed validation that names each one. A broken rule the
// operation declares answers with its status, code and parameters. Anything
// else is left to Huma, which answers 500.
func (f Failures) answerTo(ctx context.Context, err error) error {
	var invalid kernel.InvalidFieldsError
	if errors.As(err, &invalid) {
		return f.answerToFields(ctx, invalid, err)
	}
	var broken kernel.BrokenRule
	if !errors.As(err, &broken) {
		return err
	}
	status, declared := f.statusOf(broken.Code())
	if !declared {
		return fmt.Errorf("the operation does not declare the code %s: %w", broken.Code(), err)
	}
	problem := newProblem(ctx, status, broken.Code(), broken.Error())
	if params := paramsOf(broken); params != nil {
		problem.Params = params
	}
	return problem
}

// answerToFields answers fields that break a rule of the domain as a failed
// validation, the shape of a request that does not follow the contract: one
// entry in `errors` for each field, at its place in the body, with the code
// of the rule. Every rule must be declared with InField, as any broken rule.
func (f Failures) answerToFields(ctx context.Context, invalid kernel.InvalidFieldsError, err error) error {
	undeclared := slices.IndexFunc(invalid.Fields, func(field kernel.FieldError) bool {
		return !f.declaresInField(field.Rule.Code())
	})
	if undeclared >= 0 {
		return fmt.Errorf("the operation does not declare the code %s: %w", invalid.Fields[undeclared].Rule.Code(), err)
	}
	problem := newProblem(ctx, http.StatusUnprocessableEntity, validationFailed, validationFailedDetail)
	for _, field := range invalid.Fields {
		problem.Errors = append(problem.Errors, FieldError{
			Location: "body." + field.Field,
			Code:     field.Rule.Code(),
			Params:   paramsOrNone(field.Rule),
			Message:  field.Rule.Error(),
		})
	}
	return problem
}

func (f Failures) declaresInField(code string) bool {
	return slices.ContainsFunc(f[http.StatusUnprocessableEntity], func(declared kernel.BrokenRule) bool {
		_, inField := declared.(fieldRule)
		return inField && declared.Code() == code
	})
}

// paramsOrNone returns the parameters of a broken rule, and an empty object
// when it has none.
func paramsOrNone(broken kernel.BrokenRule) map[string]any {
	params := paramsOf(broken)
	if params == nil {
		return map[string]any{}
	}
	return params
}

// paramsOf returns the parameters of a broken rule, and nil when it has none.
func paramsOf(broken kernel.BrokenRule) map[string]any {
	parameterised, has := broken.(interface{ Params() map[string]any })
	if !has {
		return nil
	}
	return parameterised.Params()
}

func (f Failures) statusOf(code string) (int, bool) {
	statuses := f.statuses()
	position := slices.IndexFunc(statuses, func(status int) bool { return containsCode(f[status], code) })
	if position < 0 {
		return 0, false
	}
	return statuses[position], true
}
