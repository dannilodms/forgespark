package application

import (
	"context"
	"errors"
	"io"
	"time"

	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
)

// IssueSetupCode issues the setup code that opens the setup. Only the
// `issue-setup-code` command of the binary runs it: it has no REST operation,
// because whoever runs it has the server and no session.
type IssueSetupCode struct {
	transactions  Transactions
	setupCodes    SetupCodeRepository
	consultancies ConsultancyRepository
	recordEntry   *auditapplication.RecordEntry
	clock         Clock
	random        io.Reader
}

// NewIssueSetupCode builds the use case. The random source is the
// cryptographic one of the standard library everywhere but in a test.
func NewIssueSetupCode(
	transactions Transactions,
	setupCodes SetupCodeRepository,
	consultancies ConsultancyRepository,
	recordEntry *auditapplication.RecordEntry,
	clock Clock,
	random io.Reader,
) *IssueSetupCode {
	return &IssueSetupCode{
		transactions:  transactions,
		setupCodes:    setupCodes,
		consultancies: consultancies,
		recordEntry:   recordEntry,
		clock:         clock,
		random:        random,
	}
}

// IssuedSetupCode is what whoever installs the system is told. Code is the
// only copy of the setup code there is: the system keeps its hash.
type IssuedSetupCode struct {
	Code      domain.PlainSetupCode
	ExpiresAt time.Time
	Opens     domain.SetupCodeOpens
}

// Execute ends every setup code issued before, stores the hash of a new one
// and records the operation in the audit log, all in one transaction. The
// code opens the first setup while no consultancy exists, and the
// registration of a new administrator once one does.
func (u *IssueSetupCode) Execute(ctx context.Context) (IssuedSetupCode, error) {
	var issued IssuedSetupCode
	err := u.transactions.Run(ctx, func(ctx context.Context) error {
		var err error
		issued, err = u.issue(ctx)
		return err
	})
	if err != nil {
		return IssuedSetupCode{}, err
	}
	return issued, nil
}

func (u *IssueSetupCode) issue(ctx context.Context) (IssuedSetupCode, error) {
	now := u.clock.Now()

	if err := u.endEarlierSetupCodes(ctx, now); err != nil {
		return IssuedSetupCode{}, err
	}
	id, err := u.setupCodes.NextIdentifier(ctx)
	if err != nil {
		return IssuedSetupCode{}, err
	}
	setupCode, plain, err := domain.IssueSetupCode(id, now, u.random)
	if err != nil {
		return IssuedSetupCode{}, err
	}
	if err := u.setupCodes.Save(ctx, setupCode); err != nil {
		return IssuedSetupCode{}, err
	}
	// Recorded inside the transaction: no setup code is issued without its
	// entry in the audit log.
	err = u.recordEntry.Execute(ctx, auditapplication.RecordEntryInput{
		OccurredAt: now,
		Operation:  auditapplication.SetupCodeIssued,
		Actor:      auditapplication.Installer(),
		Subject:    auditapplication.SetupCode(int64(id)),
	})
	if err != nil {
		return IssuedSetupCode{}, err
	}
	concluded, err := u.consultancies.Exists(ctx)
	if err != nil {
		return IssuedSetupCode{}, err
	}

	return IssuedSetupCode{
		Code:      plain,
		ExpiresAt: setupCode.ExpiresAt(),
		Opens:     domain.SetupCodeOpensWhen(concluded),
	}, nil
}

// endEarlierSetupCodes ends every setup code that was not ended yet, so a
// code that whoever installs the system lost sight of opens nothing.
func (u *IssueSetupCode) endEarlierSetupCodes(ctx context.Context, now time.Time) error {
	earlier, err := u.setupCodes.LoadNotEnded(ctx)
	if err != nil {
		return err
	}
	for _, setupCode := range earlier {
		setupCode.End(now)
		err = errors.Join(err, u.setupCodes.Save(ctx, setupCode))
	}
	return err
}
