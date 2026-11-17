package restapi_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldsThatBreakARuleOfTheDomainAnswerAsAFailedValidationNamingEachField(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things", `{"label":"reserved","owner":{"email":"ana.souza@fabrica.example"}}`)

	assert.Equal(t, http.StatusUnprocessableEntity, answer.status)
	assert.JSONEq(t, `{
		"title": "Unprocessable Entity",
		"status": 422,
		"detail": "validation failed",
		"code": "validation_failed",
		"params": {},
		"trace_id": "`+traceID+`",
		"errors": [
			{"location": "body.label", "code": "thing_label_reserved", "params": {}, "message": "the label is reserved"},
			{"location": "body.owner.email", "code": "too_long", "params": {"max_length": 5}, "message": "the text is longer than the field allows"}
		]
	}`, answer.raw)
	assert.NotContains(t, answer.raw, "ana.souza", "the answer repeats the value that was refused")
	assert.Empty(t, api.logs.String(), "a 4xx wrote a line to the logs")
}

func TestAFieldRuleTheOperationDoesNotDeclareAnswers500(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things", `{"label":"rude","owner":{"email":"ana@fabrica.example"}}`)

	assert.Equal(t, http.StatusInternalServerError, answer.status)
	assert.JSONEq(t, `{"title":"Internal Server Error","status":500,"code":"internal_server_error","params":{},"trace_id":"`+traceID+`"}`, answer.raw)
	logs := api.logs.String()
	assert.Equal(t, 1, strings.Count(logs, "\n"), "one line for a 5xx: %s", logs)
	assert.Contains(t, logs, "the operation does not declare the code thing_label_rude")
}

func TestAnEmptyTextWhereOneCharacterIsTheLeastIsNamedAsRequired(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things", `{"label":"","owner":{"email":""}}`)

	assert.Equal(t, http.StatusUnprocessableEntity, answer.status)
	assert.Equal(t, "validation_failed", answer.problem.Code)
	require.Len(t, answer.problem.Errors, 2)
	assert.Equal(t, "body.label", answer.problem.Errors[0].Location)
	assert.Equal(t, "required", answer.problem.Errors[0].Code)
	assert.Equal(t, "body.owner.email", answer.problem.Errors[1].Location)
	assert.Equal(t, "required", answer.problem.Errors[1].Code)
}

// The parser of e-mail addresses quotes, in some of its errors, the part of
// the address it could not read.
func TestAnEmailRefusedByTheContractIsNotRepeatedInTheMessage(t *testing.T) {
	t.Parallel()
	api := serve(t)
	refused := map[string]string{
		"two addresses":       `ana@fabrica.example, guarded.souza@fabrica.example`,
		"a name and brackets": `Guarded Souza <guarded.souza@fabrica.example>`,
		"no at sign":          `guarded.souza.fabrica.example`,
	}
	for name, email := range refused {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			answer := api.send(t, http.MethodPost, "/api/things", `{"label":"ledger","owner":{"email":"`+email+`"}}`)

			assert.Equal(t, http.StatusUnprocessableEntity, answer.status)
			require.Len(t, answer.problem.Errors, 1)
			assert.Equal(t, "body.owner.email", answer.problem.Errors[0].Location)
			assert.Equal(t, "invalid", answer.problem.Errors[0].Code)
			assert.NotContains(t, strings.ToLower(answer.raw), "guarded", "the answer repeats the value that was refused")
		})
	}
}

func TestTheSpecListsTheRulesOfAFieldAnOperationAnswersWith(t *testing.T) {
	t.Parallel()
	api := serve(t)

	described := api.api.OpenAPI().Paths["/api/things"].Post.Responses["422"].Description

	assert.Contains(t, described, "`errors`")
	assert.Contains(t, described, "`thing_label_reserved`: the label is reserved.")
	assert.Contains(t, described, "`too_long`: the text is longer than the field allows.")
	assert.Contains(t, described, "`required`")
}
