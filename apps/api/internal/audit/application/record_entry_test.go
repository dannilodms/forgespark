package application_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application/applicationtest"
)

var noon = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

func TestRecordingAnEntryAddsItToTheAuditLog(t *testing.T) {
	auditLog := applicationtest.NewAuditLog()
	recordEntry := application.NewRecordEntry(auditLog)

	err := recordEntry.Execute(context.Background(), application.RecordEntryInput{
		OccurredAt: noon,
		Operation:  application.SetupCodeIssued,
		Actor:      application.Installer(),
		Subject:    application.SetupCode(7),
	})

	require.NoError(t, err)
	entries := auditLog.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, noon, entries[0].OccurredAt())
	assert.Equal(t, "setup_code_issued", string(entries[0].Operation()))
	assert.Equal(t, "installer", string(entries[0].Actor().Kind()))
	_, identified := entries[0].Actor().Identifier()
	assert.False(t, identified, "whoever installs the system has no identifier")
	assert.Equal(t, "setup_code", string(entries[0].Subject().Kind()))
	assert.Equal(t, int64(7), entries[0].Subject().Identifier())
}

func TestTheConclusionOfASetupIsRecordedWithTheSetupCodeUsedAndTheUserCreated(t *testing.T) {
	auditLog := applicationtest.NewAuditLog()
	recordEntry := application.NewRecordEntry(auditLog)

	err := recordEntry.Execute(context.Background(), application.RecordEntryInput{
		OccurredAt: noon,
		Operation:  application.SetupConcluded,
		Actor:      application.SetupCodeUsed(7),
		Subject:    application.User(3),
	})

	require.NoError(t, err)
	entries := auditLog.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "setup_concluded", string(entries[0].Operation()))
	assert.Equal(t, "setup_code", string(entries[0].Actor().Kind()))
	actorID, identified := entries[0].Actor().Identifier()
	assert.True(t, identified)
	assert.Equal(t, int64(7), actorID)
	assert.Equal(t, "user", string(entries[0].Subject().Kind()))
	assert.Equal(t, int64(3), entries[0].Subject().Identifier())
}

func TestTheRegistrationOfAnAdministratorIsRecordedWithTheSetupCodeUsedAndTheUserCreated(t *testing.T) {
	auditLog := applicationtest.NewAuditLog()
	recordEntry := application.NewRecordEntry(auditLog)

	err := recordEntry.Execute(context.Background(), application.RecordEntryInput{
		OccurredAt: noon,
		Operation:  application.AdministratorRegistered,
		Actor:      application.SetupCodeUsed(9),
		Subject:    application.User(4),
	})

	require.NoError(t, err)
	entries := auditLog.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "administrator_registered", string(entries[0].Operation()))
	assert.Equal(t, application.SetupCodeUsed(9), entries[0].Actor())
	assert.Equal(t, application.User(4), entries[0].Subject())
}

func TestALoginIsRecordedWithTheUserWhoLoggedInAndTheWayIn(t *testing.T) {
	for _, way := range []application.WayIn{application.LoginCodeWayIn, "local"} {
		t.Run(string(way), func(t *testing.T) {
			auditLog := applicationtest.NewAuditLog()
			recordEntry := application.NewRecordEntry(auditLog)

			err := recordEntry.Execute(context.Background(), application.RecordEntryInput{
				OccurredAt: noon,
				Operation:  application.LoggedIn,
				Actor:      application.ActingUser(5),
				Subject:    application.User(5),
				WayIn:      way,
			})

			require.NoError(t, err)
			entries := auditLog.Entries()
			require.Len(t, entries, 1)
			assert.Equal(t, "logged_in", string(entries[0].Operation()))
			assert.Equal(t, "user", string(entries[0].Actor().Kind()))
			actorID, identified := entries[0].Actor().Identifier()
			assert.True(t, identified)
			assert.Equal(t, int64(5), actorID)
			assert.Equal(t, application.User(5), entries[0].Subject())
			wayIn, has := entries[0].WayIn()
			assert.True(t, has)
			assert.Equal(t, way, wayIn)
		})
	}
	assert.Equal(t, "login_code", string(application.LoginCodeWayIn))
}

func TestAnOperationOtherThanALoginHasNoWayIn(t *testing.T) {
	auditLog := applicationtest.NewAuditLog()
	recordEntry := application.NewRecordEntry(auditLog)

	err := recordEntry.Execute(context.Background(), application.RecordEntryInput{
		OccurredAt: noon,
		Operation:  application.SetupCodeIssued,
		Actor:      application.Installer(),
		Subject:    application.SetupCode(7),
	})

	require.NoError(t, err)
	_, has := auditLog.Entries()[0].WayIn()
	assert.False(t, has)
}

func TestALoginWhoseEntryDoesNotFitIsRefusedAndNotRecorded(t *testing.T) {
	fitting := application.RecordEntryInput{
		OccurredAt: noon,
		Operation:  application.LoggedIn,
		Actor:      application.ActingUser(5),
		Subject:    application.User(5),
		WayIn:      application.LoginCodeWayIn,
	}
	cases := map[string]func(input *application.RecordEntryInput){
		"no way in": func(input *application.RecordEntryInput) {
			input.WayIn = ""
		},
		"a way in that is not a key": func(input *application.RecordEntryInput) {
			input.WayIn = "Login Code"
		},
		"a way in longer than a key": func(input *application.RecordEntryInput) {
			input.WayIn = application.WayIn("provider_" + strings.Repeat("a", 64))
		},
		"a user with no identifier": func(input *application.RecordEntryInput) {
			input.Actor = application.ActingUser(0)
		},
		"another user than the one who logged in": func(input *application.RecordEntryInput) {
			input.Subject = application.User(6)
		},
		"a way in for another operation": func(input *application.RecordEntryInput) {
			input.Operation = application.AdministratorRegistered
			input.Actor = application.SetupCodeUsed(9)
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			auditLog := applicationtest.NewAuditLog()
			recordEntry := application.NewRecordEntry(auditLog)
			input := fitting
			change(&input)

			err := recordEntry.Execute(context.Background(), input)

			assert.Error(t, err)
			assert.Empty(t, auditLog.Entries())
		})
	}
}

func TestAnEntryWhoseActorIsASetupCodeWithNoIdentifierIsRefused(t *testing.T) {
	auditLog := applicationtest.NewAuditLog()
	recordEntry := application.NewRecordEntry(auditLog)

	err := recordEntry.Execute(context.Background(), application.RecordEntryInput{
		OccurredAt: noon,
		Operation:  application.SetupConcluded,
		Actor:      application.SetupCodeUsed(0),
		Subject:    application.User(3),
	})

	assert.Error(t, err)
	assert.Empty(t, auditLog.Entries())
}

func TestAnEntryThatDoesNotFitItsOperationIsRefusedAndNotRecorded(t *testing.T) {
	fitting := application.RecordEntryInput{
		OccurredAt: noon,
		Operation:  application.SetupCodeIssued,
		Actor:      application.Installer(),
		Subject:    application.SetupCode(7),
	}
	cases := map[string]func(input *application.RecordEntryInput){
		"an operation the audit log does not know": func(input *application.RecordEntryInput) {
			input.Operation = "setup_code_guessed"
		},
		"no actor": func(input *application.RecordEntryInput) {
			input.Actor = application.Actor{}
		},
		"no subject": func(input *application.RecordEntryInput) {
			input.Subject = application.Subject{}
		},
		"a subject with no identifier": func(input *application.RecordEntryInput) {
			input.Subject = application.SetupCode(0)
		},
		"no time": func(input *application.RecordEntryInput) {
			input.OccurredAt = time.Time{}
		},
		"the actor of another operation": func(input *application.RecordEntryInput) {
			input.Actor = application.SetupCodeUsed(7)
		},
		"the subject of another operation": func(input *application.RecordEntryInput) {
			input.Subject = application.User(3)
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			auditLog := applicationtest.NewAuditLog()
			recordEntry := application.NewRecordEntry(auditLog)
			input := fitting
			change(&input)

			err := recordEntry.Execute(context.Background(), input)

			assert.Error(t, err)
			assert.Empty(t, auditLog.Entries())
		})
	}
}
