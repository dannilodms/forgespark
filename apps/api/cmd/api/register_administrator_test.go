package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// concludedIn concludes the first setup on the temporary database, with the
// administrator of setupBody, through an API it starts. It returns the
// address of the API.
func concludedIn(t *testing.T, temporary databasetest.Database) string {
	t.Helper()
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	answer := concludeSetup(t, api, setupBody(code))
	require.Equal(t, http.StatusNoContent, answer.status, answer.body)
	return api
}

// administratorBody is the body of registering a new administrator, every
// field filled in, as a person types it.
func administratorBody(code string) map[string]any {
	return map[string]any{
		"code": code,
		"administrator": map[string]any{
			"name":               "Bruno Lima",
			"email":              "Bruno.Lima@Fabrica.example",
			"country":            "PT",
			"national_id":        "123 456 789",
			"telephone":          "+351 21 000 1234",
			"preferred_language": "es",
		},
	}
}

// registerAdministrator posts the body to the API.
func registerAdministrator(t *testing.T, api string, body map[string]any) received {
	t.Helper()
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	return sendJSON(t, http.MethodPost, api+"/api/setup/administrators", string(encoded))
}

// backIn is an install whose setup is concluded, with a second setup code
// issued to get back in. It returns the address of the API and the code.
func backIn(t *testing.T, temporary databasetest.Database) (api, code string) {
	t.Helper()
	api = concludedIn(t, temporary)
	_, code = issueSetupCodeIn(t, temporary)
	return api, code
}

// holders lists, for each role, the e-mails of the users who hold it.
func holders(t *testing.T, owner *pgx.Conn) map[string][]string {
	t.Helper()
	rows, err := owner.Query(context.Background(), `
		SELECT roles.key, coalesce(array_agg(users.email ORDER BY users.email) FILTER (WHERE users.id IS NOT NULL), '{}')
		FROM roles
		LEFT JOIN user_roles ON user_roles.role_id = roles.id
		LEFT JOIN users ON users.id = user_roles.user_id
		GROUP BY roles.key`)
	require.NoError(t, err)
	held := map[string][]string{}
	var key string
	var emails []string
	_, err = pgx.ForEachRow(rows, []any{&key, &emails}, func() error {
		held[key] = emails
		return nil
	})
	require.NoError(t, err)
	return held
}

func TestIssueSetupCodeSaysACodeIssuedAfterTheSetupRegistersANewAdministratorAndChangesNoUser(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	concludedIn(t, temporary)

	printed, _ := issueSetupCodeIn(t, temporary)

	assert.Contains(t, printed, "It opens:          the registration of a new administrator.")
	assert.Contains(t, printed, "The users that exist are not changed.")
	assert.NotContains(t, printed, "the first setup")
}

func TestRegisteringAnAdministratorAnswers201AndSavesTheUserWithTheAdministratorRole(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api, code := backIn(t, temporary)
	owner := asOwner(t, temporary)

	answer := registerAdministrator(t, api, administratorBody(code))

	require.Equal(t, http.StatusCreated, answer.status, answer.body)
	assert.Empty(t, answer.body)
	var registered string
	require.NoError(t, owner.QueryRow(context.Background(),
		"SELECT row_to_json(users)::jsonb - 'id' FROM users WHERE email <> $1", theFirstAdministrator).Scan(&registered))
	assert.JSONEq(t, `{
		"name": "Bruno Lima", "email": "bruno.lima@fabrica.example", "country": "PT",
		"national_id": "123 456 789", "telephone": "+351210001234", "preferred_language": "es",
		"photo_file_id": null, "deactivated": false
	}`, registered)
	assert.Equal(t, map[string][]string{
		"administrator": {theFirstAdministrator, "bruno.lima@fabrica.example"},
		"developer":     {},
	}, holders(t, owner))
}

// theFirstAdministrator is the e-mail of the administrator setupBody
// registers, as stored.
const theFirstAdministrator = "ana.souza@fabrica.example"

// whatExisted is, as text, every row of the consultancy and of the roles, and
// the row of the user with this e-mail with the roles they hold.
func whatExisted(t *testing.T, owner *pgx.Conn, email string) string {
	t.Helper()
	var rows string
	require.NoError(t, owner.QueryRow(context.Background(), `
		SELECT concat_ws(' | ',
			(SELECT string_agg(row_to_json(consultancy)::text, ' | ') FROM consultancy),
			(SELECT string_agg(row_to_json(roles)::text, ' | ' ORDER BY id) FROM roles),
			(SELECT string_agg(row_to_json(users)::text, ' | ') FROM users WHERE email = $1),
			(SELECT string_agg(row_to_json(user_roles)::text, ' | ' ORDER BY role_id) FROM user_roles
				WHERE user_id = (SELECT id FROM users WHERE email = $1)))`, email).Scan(&rows))
	return rows
}

func TestRegisteringAnAdministratorLeavesTheConsultancyTheUsersThatExistedAndTheRolesAsTheyWere(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api, code := backIn(t, temporary)
	owner := asOwner(t, temporary)
	before := whatExisted(t, owner, theFirstAdministrator)
	require.Equal(t, 1, countOf(t, owner, "users"))

	answer := registerAdministrator(t, api, administratorBody(code))

	require.Equal(t, http.StatusCreated, answer.status, answer.body)
	assert.Equal(t, before, whatExisted(t, owner, theFirstAdministrator))
	assert.Equal(t, 2, countOf(t, owner, "users"))
}

func TestRegisteringAnAdministratorEndsTheSetupCodeAndLeavesOneEntryInTheAuditLog(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api, code := backIn(t, temporary)
	require.JSONEq(t, `{"concluded":true,"setup_code_waiting":true}`, setupState(t, api))

	require.Equal(t, http.StatusCreated, registerAdministrator(t, api, administratorBody(code)).status)

	rows, err := asOwner(t, temporary).Query(context.Background(), `
		SELECT audit_log.actor_kind, audit_log.subject_kind,
		       audit_log.actor_id = setup_codes.id, audit_log.subject_id = users.id,
		       audit_log.occurred_at = setup_codes.ended_at
		FROM audit_log
		JOIN setup_codes ON setup_codes.id = audit_log.actor_id
		JOIN users ON users.id = audit_log.subject_id
		WHERE audit_log.operation = 'administrator_registered'`)
	require.NoError(t, err)
	type entry struct {
		ActorKind, SubjectKind                                        string
		ActedByTheSetupCode, OnTheUserRegistered, WhenTheCodeWasEnded bool
	}
	entries, err := pgx.CollectRows(rows, pgx.RowToStructByPos[entry])
	require.NoError(t, err)
	assert.Equal(t, []entry{{
		ActorKind: "setup_code", SubjectKind: "user",
		ActedByTheSetupCode: true, OnTheUserRegistered: true, WhenTheCodeWasEnded: true,
	}}, entries)
	var registered string
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(),
		"SELECT users.email FROM audit_log JOIN users ON users.id = audit_log.subject_id WHERE operation = 'administrator_registered'").Scan(&registered))
	assert.Equal(t, "bruno.lima@fabrica.example", registered)
	assert.JSONEq(t, `{"concluded":true,"setup_code_waiting":false}`, setupState(t, api))
	requireSetupCodeInvalid(t, openSetup(t, api, code))
}

func TestTheEmailOfAUserIsRefusedWith409WhetherThatUserIsDeactivatedOrNot(t *testing.T) {
	t.Parallel()
	for name, deactivated := range map[string]bool{"a user": false, "a deactivated user": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			temporary := databasetest.Migrated(t)
			api, code := backIn(t, temporary)
			owner := asOwner(t, temporary)
			_, err := owner.Exec(context.Background(), "UPDATE users SET deactivated = $1", deactivated)
			require.NoError(t, err)
			before := everyRow(t, owner)
			body := administratorBody(code)
			object(body, "administrator")["email"] = "ANA.Souza@Fabrica.example"

			answer := registerAdministrator(t, api, body)

			require.Equal(t, http.StatusConflict, answer.status, answer.body)
			assert.Equal(t, "application/problem+json", answer.mediaType)
			assert.Equal(t, "email_in_use", answer.fields["code"])
			assert.Equal(t, before, everyRow(t, owner), "a refused registration changed what is stored")
			assert.Equal(t, http.StatusOK, openSetup(t, api, code).status, "the setup code stopped working")
		})
	}
}

func TestRegisteringAnAdministratorBeforeTheSetupIsConcludedAnswers409(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	owner := asOwner(t, temporary)
	before := everyRow(t, owner)

	answer := registerAdministrator(t, api, administratorBody(code))

	require.Equal(t, http.StatusConflict, answer.status, answer.body)
	assert.Equal(t, "application/problem+json", answer.mediaType)
	assert.Equal(t, "setup_not_concluded", answer.fields["code"])
	assert.Equal(t, before, everyRow(t, owner), "a refused registration changed what is stored")
	assert.Equal(t, "first_setup", openSetup(t, api, code).fields["opens"], "the setup code stopped working")
}

// fieldsRefused gives, for each field an answer names, the code of its rule.
func fieldsRefused(t *testing.T, answer received) map[string]any {
	t.Helper()
	require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
	require.Equal(t, "validation_failed", answer.fields["code"])
	refused := map[string]any{}
	errors, _ := answer.fields["errors"].([]any)
	for _, each := range errors {
		field, _ := each.(map[string]any)
		location, _ := field["location"].(string)
		refused[location] = field["code"]
	}
	return refused
}

func TestAFieldOrAPhotoThatBreaksARuleIsNamedAndNothingIsSaved(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api, code := backIn(t, temporary)
	owner := asOwner(t, temporary)
	outsideTheList := administratorBody(code)
	object(outsideTheList, "administrator")["country"] = "ZZ"
	neverSent := administratorBody(code)
	started := startUpload(t, api, uploadBody(code, "user_photo", "image/png", 2048))
	require.Equal(t, http.StatusCreated, started.status, started.body)
	object(neverSent, "administrator")["photo_file_id"] = started.fields["file_id"]
	before := everyRow(t, owner)

	refusedByTheDomain := registerAdministrator(t, api, outsideTheList)
	refusedByTheStorage := registerAdministrator(t, api, neverSent)

	assert.Equal(t, map[string]any{"body.administrator.country": "country_unknown"}, fieldsRefused(t, refusedByTheDomain))
	assert.Equal(t, map[string]any{"body.administrator.photo_file_id": "file_not_uploaded"}, fieldsRefused(t, refusedByTheStorage))
	assert.Equal(t, before, everyRow(t, owner), "a refused registration changed what is stored")
}

func TestThePhotoOfANewAdministratorIsConfirmedAndKept(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api, code := backIn(t, temporary)
	body := administratorBody(code)
	object(body, "administrator")["photo_file_id"] = sentImage(t, api, code, "user_photo")

	answer := registerAdministrator(t, api, body)

	require.Equal(t, http.StatusCreated, answer.status, answer.body)
	var kept bool
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(), `
		SELECT EXISTS (SELECT FROM users JOIN files ON files.id = users.photo_file_id
			WHERE users.email = 'bruno.lima@fabrica.example'
			AND files.purpose = 'user_photo' AND files.confirmed_at IS NOT NULL)`).Scan(&kept))
	assert.True(t, kept, "the new administrator has no confirmed photo")
}
