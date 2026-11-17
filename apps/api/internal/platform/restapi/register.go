package restapi

import (
	"context"
	"slices"

	"github.com/danielgtaylor/huma/v2"
)

// Register adds an operation to the API. The handler parses nothing and
// answers nothing by itself: it receives the input Huma validated, calls one
// use case and maps the result. An error it returns is answered in the error
// format: a broken rule listed in failures with its status and its code,
// anything else as a 500 whose raw error goes to the logs.
func Register[I, O any](api *API, operation huma.Operation, failures Failures, handler func(context.Context, *I) (*O, error)) {
	operation.MaxBodyBytes = api.limits.BodyBytes
	operation.BodyReadTimeout = api.limits.BodyReadTimeout
	operation.Errors = failures.statuses()

	huma.Register(api.API, operation, func(ctx context.Context, input *I) (*O, error) {
		output, err := handler(ctx, input)
		if err != nil {
			return nil, failures.answerTo(ctx, err)
		}
		return output, nil
	})
}

// statuses lists the statuses the failures answer with, in order, so the
// generated spec does not change from one run to the next.
func (f Failures) statuses() []int {
	statuses := make([]int, 0, len(f))
	for status := range f {
		statuses = append(statuses, status)
	}
	slices.Sort(statuses)
	return statuses
}

func containsCode(declared []Failure, code string) bool {
	return slices.ContainsFunc(declared, func(failure Failure) bool { return failure.Code() == code })
}
