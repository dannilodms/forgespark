package restapi

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// Register adds an operation to the API. The handler parses nothing and
// answers nothing by itself: it receives the input Huma validated, calls one
// use case and maps the result. An error it returns is answered in the error
// format: a broken rule listed in failures with its status and its code,
// anything else as a 500 whose raw error goes to the logs.
//
// The spec lists, under each status of the operation, the codes it answers
// with.
func Register[I, O any](api *API, operation huma.Operation, failures Failures, handler func(context.Context, *I) (*O, error)) {
	operation.MaxBodyBytes = api.limits.BodyBytes
	operation.BodyReadTimeout = api.limits.BodyReadTimeout
	operation.Errors = failures.statuses()
	// Huma fills this map with the responses of the operation, so they can
	// be described once it has.
	operation.Responses = map[string]*huma.Response{}

	huma.Register(api.API, operation, func(ctx context.Context, input *I) (*O, error) {
		output, err := handler(ctx, input)
		if err != nil {
			return nil, failures.answerTo(ctx, err)
		}
		return output, nil
	})

	for status, response := range operation.Responses {
		response.Description = failures.describe(status, response.Description)
	}
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

// describe adds to the description of a response the codes the operation
// answers with under that status: the broken rules it declares, and the two
// codes any operation answers with.
func (f Failures) describe(status, description string) string {
	number, _ := strconv.Atoi(status)
	if number < http.StatusBadRequest {
		return description
	}
	codes := []string{}
	for _, failure := range f[number] {
		codes = append(codes, fmt.Sprintf("`%s`: %s.", failure.Code(), failure.Error()))
	}
	switch number {
	case http.StatusUnprocessableEntity:
		codes = append(codes, "`"+validationFailed+"`: the request does not follow this contract, and `errors` names each field.")
	case http.StatusInternalServerError:
		codes = append(codes, "`"+codeOfStatus(number)+"`: the failure is in the logs, under the `trace_id`.")
	}
	return strings.Join(append([]string{description + "."}, codes...), " ")
}

func containsCode(declared []Failure, code string) bool {
	return slices.ContainsFunc(declared, func(failure Failure) bool { return failure.Code() == code })
}
