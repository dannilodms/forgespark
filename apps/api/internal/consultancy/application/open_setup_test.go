package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
)

func TestAValidSetupCodeOpensTheFirstSetupAndTellsWhenItExpires(t *testing.T) {
	installed := newInstallation()
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	installed.clock.Advance(59 * time.Minute)

	opened, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: firstCode})

	require.NoError(t, err)
	assert.Equal(t, "first_setup", string(opened.Opens))
	assert.Equal(t, noon.Add(time.Hour), opened.ExpiresAt)
}

// requireSetupCodeInvalid checks the one answer a setup code that opens
// nothing gets, whatever the reason.
func requireSetupCodeInvalid(t *testing.T, err error) {
	t.Helper()
	var invalid domain.SetupCodeInvalidError
	require.ErrorAs(t, err, &invalid)
	assert.Equal(t, "setup_code_invalid", invalid.Code())
}

func TestAWrongSetupCodeOpensNothingAndLeavesTheSetupCodeWorking(t *testing.T) {
	installed := newInstallation()
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	before := installed.setupCodes.All()

	opened, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: "ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ"})

	requireSetupCodeInvalid(t, err)
	assert.Zero(t, opened)
	assert.Equal(t, before, installed.setupCodes.All(), "a wrong code changed the stored setup code")
	_, err = installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: firstCode})
	assert.NoError(t, err, "the setup code stopped working after a wrong one was typed")
}

func TestAnExpiredSetupCodeGetsTheAnswerOfAWrongOne(t *testing.T) {
	installed := newInstallation()
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	// A setup code lasts one hour: at the hour it no longer works.
	installed.clock.Advance(time.Hour)
	before := installed.setupCodes.All()

	opened, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: firstCode})

	requireSetupCodeInvalid(t, err)
	assert.Zero(t, opened)
	assert.Equal(t, before, installed.setupCodes.All(), "checking an expired code changed it")
}

func TestASetupCodeEndedByANewOneGetsTheAnswerOfAWrongOne(t *testing.T) {
	installed := newInstallation()
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	_, err = installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	before := installed.setupCodes.All()

	opened, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: firstCode})

	requireSetupCodeInvalid(t, err)
	assert.Zero(t, opened)
	assert.Equal(t, before, installed.setupCodes.All(), "checking an ended code changed it")
	_, err = installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: secondCode})
	assert.NoError(t, err, "the new setup code does not work")
}

func TestASetupCodeIsAcceptedHoweverItIsTyped(t *testing.T) {
	// The code issued is 01234-56789-ABCDE-FGHJK.
	typed := map[string]string{
		"as printed":                   "01234-56789-ABCDE-FGHJK",
		"in lower case":                "01234-56789-abcde-fghjk",
		"without hyphens":              "0123456789ABCDEFGHJK",
		"with spaces":                  "  01234 56789  ABCDE FGHJK ",
		"with O and I for 0 and 1":     "OI234-56789-ABCDE-FGHJK",
		"with L for 1":                 "0L234-56789-ABCDE-FGHJK",
		"in lower case with o and l":   "ol234 56789 abcde fghjk",
		"with a hyphen after each one": "0-1-2-3-4-5-6-7-8-9-A-B-C-D-E-F-G-H-J-K",
	}
	for name, code := range typed {
		t.Run(name, func(t *testing.T) {
			installed := newInstallation()
			_, err := installed.issueSetupCode.Execute(context.Background())
			require.NoError(t, err)

			opened, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: code})

			require.NoError(t, err)
			assert.Equal(t, noon.Add(time.Hour), opened.ExpiresAt)
		})
	}
}

func TestATextThatIsNoSetupCodeGetsTheAnswerOfAWrongOne(t *testing.T) {
	typed := map[string]string{
		"nothing":                 "",
		"one character short":     "01234-56789-ABCDE-FGHJ",
		"one character too many":  "01234-56789-ABCDE-FGHJKM",
		"a letter never printed":  "U1234-56789-ABCDE-FGHJK",
		"a character of no group": "01234_56789_ABCDE_FGHJK",
	}
	for name, code := range typed {
		t.Run(name, func(t *testing.T) {
			installed := newInstallation()
			_, err := installed.issueSetupCode.Execute(context.Background())
			require.NoError(t, err)

			opened, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: code})

			requireSetupCodeInvalid(t, err)
			assert.Zero(t, opened)
		})
	}
}
