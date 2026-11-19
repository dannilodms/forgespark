package restapi

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
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
	if operation.Responses == nil {
		operation.Responses = map[string]*huma.Response{}
	}

	huma.Register(api.API, operation, func(ctx context.Context, input *I) (*O, error) {
		output, err := handler(ctx, input)
		if err != nil {
			return nil, failures.answerTo(ctx, err)
		}
		return output, nil
	})

	// Huma declares a 422 for every operation that has parameters or a body.
	// One that checks neither, and declares no broken rule under 422, never
	// answers it.
	if operation.SkipValidateParams && operation.SkipValidateBody && len(failures[http.StatusUnprocessableEntity]) == 0 {
		delete(operation.Responses, strconv.Itoa(http.StatusUnprocessableEntity))
	}
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
	inFields := []string{}
	for _, failure := range f[number] {
		_, inField := failure.(fieldRule)
		described := fmt.Sprintf("`%s`: %s.", failure.Code(), failure.Error())
		codes = appendIf(codes, described, !inField)
		inFields = appendIf(inFields, described, inField)
	}
	switch number {
	case http.StatusUnprocessableEntity:
		codes = append(codes, "`"+validationFailed+"`: the request does not follow this contract, or a field breaks a rule, "+
			"and `errors` names each field with the code `"+required+"` or `"+invalid+"`, or the code of the rule.")
		codes = appendIf(codes, "The rules of a field: "+strings.Join(inFields, " "), len(inFields) > 0)
	case http.StatusInternalServerError:
		codes = append(codes, "`"+codeOfStatus(number)+"`: the failure is in the logs, under the `trace_id`.")
	}
	return strings.Join(append([]string{description + "."}, codes...), " ")
}

func appendIf(texts []string, text string, wanted bool) []string {
	if !wanted {
		return texts
	}
	return append(texts, text)
}

// containsCode tells whether the operation declares a broken rule with this
// code. A rule about a field is declared for the fields only.
func containsCode(declared []kernel.BrokenRule, code string) bool {
	return slices.ContainsFunc(declared, func(failure kernel.BrokenRule) bool {
		_, inField := failure.(fieldRule)
		return !inField && failure.Code() == code
	})
}
