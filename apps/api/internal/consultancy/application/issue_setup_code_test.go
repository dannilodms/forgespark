package application_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	auditapplicationtest "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application/applicationtest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application/applicationtest"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	filesapplicationtest "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application/applicationtest"
)

var noon = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

// countingBytes stands in for the random source: it gives 0, 1, 2 and so on.
// A setup code takes the five low bits of each byte, so the first code issued
// is the 20 first characters of the alphabet and the second starts at the
// twenty-first.
type countingBytes struct {
	next byte
}

func (c *countingBytes) Read(buffer []byte) (int, error) {
	for i := range buffer {
		buffer[i] = c.next
		c.next++
	}
	return len(buffer), nil
}

const (
	firstCode  = "01234-56789-ABCDE-FGHJK"
	secondCode = "MNPQR-STVWX-YZ012-34567"
)

// installation is a system with nothing in it, built on the in-memory
// implementations of this area and of the files area, with the clock stopped
// at noon.
type installation struct {
	clock            *applicationtest.Clock
	setupCodes       *applicationtest.SetupCodeRepository
	consultancies    *applicationtest.ConsultancyRepository
	users            *applicationtest.UserRepository
	roles            *applicationtest.RoleRepository
	auditLog         *auditapplicationtest.AuditLog
	files            *filesapplicationtest.FileRepository
	objects          *filesapplicationtest.Storage
	transactions     *applicationtest.Transactions
	issueSetupCode   *application.IssueSetupCode
	openSetup        *application.OpenSetup
	startSetupUpload *application.StartSetupUpload
	concludeSetup    *application.ConcludeSetup
	tellSetupState   *application.TellSetupState
}

func newInstallation() installation {
	clock := applicationtest.NewClock(noon)
	setupCodes := applicationtest.NewSetupCodeRepository()
	consultancies := applicationtest.NewConsultancyRepository()
	users := applicationtest.NewUserRepository()
	roles := applicationtest.NewRoleRepository()
	auditLog := auditapplicationtest.NewAuditLog()
	files := filesapplicationtest.NewFileRepository()
	objects := filesapplicationtest.NewStorage()
	transactions := applicationtest.NewTransactions(setupCodes, consultancies, users, roles, auditLog, files)
	recordEntry := auditapplication.NewRecordEntry(auditLog)
	startUpload := filesapplication.NewStartUpload(transactions, files, objects, clock, &countingBytes{})
	confirmUpload := filesapplication.NewConfirmUpload(transactions, files, objects, clock)
	return installation{
		clock:            clock,
		setupCodes:       setupCodes,
		consultancies:    consultancies,
		users:            users,
		roles:            roles,
		auditLog:         auditLog,
		files:            files,
		objects:          objects,
		transactions:     transactions,
		issueSetupCode:   application.NewIssueSetupCode(transactions, setupCodes, consultancies, recordEntry, clock, &countingBytes{}),
		openSetup:        application.NewOpenSetup(setupCodes, consultancies, clock),
		startSetupUpload: application.NewStartSetupUpload(setupCodes, startUpload, clock),
		concludeSetup: application.NewConcludeSetup(
			transactions, setupCodes, consultancies, users, roles, confirmUpload, recordEntry, clock,
		),
		tellSetupState: application.NewTellSetupState(applicationtest.NewSetupStateReader(consultancies, setupCodes), clock),
	}
}

func TestIssuingASetupCodeReturnsTheCodeItsExpiryAndWhatItOpens(t *testing.T) {
	installed := newInstallation()

	issued, err := installed.issueSetupCode.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, firstCode, issued.Code.Reveal())
	assert.Equal(t, noon.Add(time.Hour), issued.ExpiresAt)
	assert.Equal(t, "first_setup", string(issued.Opens))
}

// The way back in: once the consultancy exists, a setup code registers a new
// administrator.
func TestASetupCodeIssuedOnceTheSetupIsConcludedOpensTheRegistrationOfANewAdministrator(t *testing.T) {
	installed := concluded(t)

	issued, err := installed.issueSetupCode.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, secondCode, issued.Code.Reveal())
	assert.Equal(t, "new_administrator", string(issued.Opens))
}

func TestTheSetupCodeIsStoredAsItsHashAndNeverAsTheCode(t *testing.T) {
	installed := newInstallation()

	_, err := installed.issueSetupCode.Execute(context.Background())

	require.NoError(t, err)
	stored := installed.setupCodes.All()
	require.Len(t, stored, 1)
	hash := stored[0].Hash()
	// The SHA-256 hash of 0123456789ABCDEFGHJK: the first code, normalised.
	assert.Equal(t, "a6b239e79989fa88083024f6fa3518b4f917667d90dc8408053fcf0cfc8cfb8a", hex.EncodeToString(hash[:]))
	assert.Equal(t, noon, stored[0].IssuedAt())
	assert.Equal(t, noon.Add(time.Hour), stored[0].ExpiresAt())
	_, ended := stored[0].EndedAt()
	assert.False(t, ended, "a setup code just issued was ended")
}

func TestANewSetupCodeEndsTheOneBefore(t *testing.T) {
	installed := newInstallation()
	first, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	installed.clock.Advance(10 * time.Minute)

	second, err := installed.issueSetupCode.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, firstCode, first.Code.Reveal())
	assert.Equal(t, secondCode, second.Code.Reveal())
	stored := installed.setupCodes.All()
	require.Len(t, stored, 2)
	endedAt, ended := stored[0].EndedAt()
	assert.True(t, ended, "the first setup code still works")
	assert.Equal(t, noon.Add(10*time.Minute), endedAt)
	_, ended = stored[1].EndedAt()
	assert.False(t, ended, "the second setup code was ended")
}

func TestEachSetupCodeIssuedLeavesOneEntryInTheAuditLog(t *testing.T) {
	installed := newInstallation()
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	installed.clock.Advance(10 * time.Minute)

	_, err = installed.issueSetupCode.Execute(context.Background())

	require.NoError(t, err)
	stored := installed.setupCodes.All()
	require.Len(t, stored, 2)
	entries := installed.auditLog.Entries()
	require.Len(t, entries, 2)
	assert.Equal(t, auditapplication.SetupCodeIssued, entries[0].Operation())
	assert.Equal(t, auditapplication.Installer(), entries[0].Actor())
	assert.Equal(t, auditapplication.SetupCode(int64(stored[0].ID())), entries[0].Subject())
	assert.Equal(t, noon, entries[0].OccurredAt())
	assert.Equal(t, auditapplication.SetupCodeIssued, entries[1].Operation())
	assert.Equal(t, auditapplication.Installer(), entries[1].Actor())
	assert.Equal(t, auditapplication.SetupCode(int64(stored[1].ID())), entries[1].Subject())
	assert.Equal(t, noon.Add(10*time.Minute), entries[1].OccurredAt())
	assert.NotEqual(t, stored[0].ID(), stored[1].ID())
}

func TestAFailureToWriteTheAuditEntryLeavesNoSetupCodeStoredAndTheEarlierOneWorking(t *testing.T) {
	installed := newInstallation()
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	installed.clock.Advance(10 * time.Minute)
	auditLogIsDown := errors.New("the audit log is down")
	installed.auditLog.FailWith(auditLogIsDown)

	issued, err := installed.issueSetupCode.Execute(context.Background())

	require.ErrorIs(t, err, auditLogIsDown)
	assert.Empty(t, issued.Code.Reveal(), "a setup code that was not stored was returned")
	stored := installed.setupCodes.All()
	require.Len(t, stored, 1, "the setup code of the failed operation was stored")
	_, ended := stored[0].EndedAt()
	assert.False(t, ended, "the failed operation ended the earlier setup code")
	assert.Len(t, installed.auditLog.Entries(), 1)
}

func TestAnIssuedSetupCodeShowsItsCharactersOnlyWhenRevealed(t *testing.T) {
	installed := newInstallation()
	issued, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	encoded, err := json.Marshal(issued)
	require.NoError(t, err)
	var logged bytes.Buffer
	slog.New(slog.NewJSONHandler(&logged, nil)).Info("issued", "result", issued, "code", issued.Code)
	slog.New(slog.NewTextHandler(&logged, nil)).Info("issued", "result", issued, "code", issued.Code)

	printed := fmt.Sprintf("%v %+v %#v %s %q", issued, issued, issued, issued.Code, issued.Code) +
		string(encoded) + logged.String()

	assert.NotContains(t, printed, "01234")
	assert.NotContains(t, printed, "FGHJK")
	assert.Contains(t, printed, "[setup code]")
}
