package main

import (
	"net/http"
	"testing"

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

func TestIssueSetupCodeSaysACodeIssuedAfterTheSetupRegistersANewAdministratorAndChangesNoUser(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	concludedIn(t, temporary)

	printed, _ := issueSetupCodeIn(t, temporary)

	assert.Contains(t, printed, "It opens:          the registration of a new administrator.")
	assert.Contains(t, printed, "The users that exist are not changed.")
	assert.NotContains(t, printed, "the first setup")
}
