package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
)

func TestASetupCodeThatIsWrongExpiredOrEndedConcludesNothing(t *testing.T) {
	cases := map[string]func(t *testing.T, installed installation) (typed string){
		"wrong":            func(*testing.T, installation) string { return "ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ" },
		"not a setup code": func(*testing.T, installation) string { return "open sesame" },
		"expired": func(_ *testing.T, installed installation) string {
			// Issued at noon, it stops working at one, and the clock is at
			// ten past twelve.
			installed.clock.Advance(50 * time.Minute)
			return firstCode
		},
		"ended by a new one": func(t *testing.T, installed installation) string {
			_, err := installed.issueSetupCode.Execute(context.Background())
			require.NoError(t, err)
			return firstCode
		},
	}
	for name, arrange := range cases {
		t.Run(name, func(t *testing.T) {
			installed := issued(t)
			typed := arrange(t, installed)
			before := installed.setupCodes.All()

			err := installed.concludeSetup.Execute(context.Background(), typedSetup(typed))

			requireSetupCodeInvalid(t, err)
			assert.Empty(t, installed.consultancies.All(), "a consultancy was saved")
			assert.Empty(t, installed.users.All(), "a user was saved")
			assert.Empty(t, installed.roles.All(), "a role was saved")
			assert.Equal(t, before, installed.setupCodes.All(), "a setup code was changed")
		})
	}
}

func TestConcludingTwiceWithTheSameCodeSavesOnceAndTheSecondAttemptIsToldTheCodeIsInvalid(t *testing.T) {
	installed := issued(t)
	require.NoError(t, installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode)))
	again := typedSetup(firstCode)
	again.Administrator.Email = "bruno.lima@fabrica.example"

	err := installed.concludeSetup.Execute(context.Background(), again)

	requireSetupCodeInvalid(t, err)
	assert.Len(t, installed.consultancies.All(), 1)
	assert.Len(t, installed.users.All(), 1)
	assert.Len(t, installed.roles.All(), 2)
	assert.Len(t, installed.auditLog.Entries(), 2)
}

func TestConcludingTheFirstSetupWithACodeIssuedAfterItWasConcludedIsToldItIsAlreadyConcluded(t *testing.T) {
	installed := issued(t)
	require.NoError(t, installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode)))
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	before := installed.setupCodes.All()

	err = installed.concludeSetup.Execute(context.Background(), typedSetup(secondCode))

	var concluded domain.SetupAlreadyConcludedError
	require.ErrorAs(t, err, &concluded)
	assert.Equal(t, "setup_already_concluded", concluded.Code())
	assert.Len(t, installed.consultancies.All(), 1)
	assert.Len(t, installed.users.All(), 1)
	assert.Len(t, installed.roles.All(), 2)
	assert.Equal(t, before, installed.setupCodes.All(), "the setup code was ended")
}

func TestAFailureAtAnyPointOfConcludingLeavesNothingSavedAndTheSetupCodeWorking(t *testing.T) {
	failing := errors.New("the database refuses the write")
	points := map[string]func(installed installation){
		"saving a role":            func(installed installation) { installed.roles.FailWith(failing) },
		"saving the consultancy":   func(installed installation) { installed.consultancies.FailWith(failing) },
		"saving the administrator": func(installed installation) { installed.users.FailWith(failing) },
		"ending the setup code":    func(installed installation) { installed.setupCodes.FailWith(failing) },
		"recording the conclusion": func(installed installation) { installed.auditLog.FailWith(failing) },
	}
	for point, fail := range points {
		t.Run(point, func(t *testing.T) {
			installed := issued(t)
			before := installed.setupCodes.All()
			fail(installed)

			err := installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode))

			require.ErrorIs(t, err, failing)
			requireNothingSaved(t, installed)
			assert.Equal(t, before, installed.setupCodes.All(), "the setup code was ended")
		})
	}
}

func TestASetupCodeIssuedAfterTheSetupWasConcludedOpensTheRegistrationOfANewAdministrator(t *testing.T) {
	installed := issued(t)
	require.NoError(t, installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode)))
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)

	opened, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: secondCode})

	require.NoError(t, err)
	assert.Equal(t, "new_administrator", string(opened.Opens))
}

func TestTheStateOfTheSetupFollowsItsSetupCodesAndItsConclusion(t *testing.T) {
	installed := newInstallation()
	state := func() application.SetupState {
		t.Helper()
		told, err := installed.tellSetupState.Execute(context.Background())
		require.NoError(t, err)
		return told
	}

	assert.Equal(t, application.SetupState{Concluded: false, SetupCodeWaiting: false}, state(), "a new install")
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	assert.Equal(t, application.SetupState{Concluded: false, SetupCodeWaiting: true}, state(), "a setup code issued")
	installed.clock.Advance(time.Hour)
	assert.Equal(t, application.SetupState{Concluded: false, SetupCodeWaiting: false}, state(), "the setup code expired")
	_, err = installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	require.NoError(t, installed.concludeSetup.Execute(context.Background(), typedSetup(secondCode)))
	assert.Equal(t, application.SetupState{Concluded: true, SetupCodeWaiting: false}, state(), "the setup concluded")
	_, err = installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	assert.Equal(t, application.SetupState{Concluded: true, SetupCodeWaiting: true}, state(), "a setup code issued after the setup")
}
