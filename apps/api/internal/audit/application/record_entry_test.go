package application_test

import (
	"context"
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
