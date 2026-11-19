package application

import (
	"context"
	"errors"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	emailapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
)

// OldRecords are what the daily cleanup removes from: the login code
// requests, the asks for a login code, the sessions, and, through its use
// case, the bounce notices of the e-mail area; and the clock.
type OldRecords struct {
	Requests             LoginCodeRequestRepository
	Asks                 AskedAddressRepository
	EndedSessions        EndedSessions
	CleanUpBounceNotices *emailapplication.CleanUpBounceNotices
	Clock                Clock
}

// CleanUpOldRecords is the daily cleanup of the worker (Jobs in
// CODING_STANDARDS.md): it removes what the system keeps only for a while,
// and reads no longer. A login code request goes a day after it ended or
// expired, an ask for a login code once no limit counts it, a session once it
// ended, and a bounce notice after 30 days, which the e-mail area removes.
// It runs once a day, so the database is reached no more often than that.
type CleanUpOldRecords struct {
	from OldRecords
}

// NewCleanUpOldRecords builds the use case.
func NewCleanUpOldRecords(from OldRecords) *CleanUpOldRecords {
	return &CleanUpOldRecords{from: from}
}

// Execute removes the old records. Each removal is a statement of its own,
// and a failed one stops none of the others: the job fails, and runs again.
// A run more than it needed removes nothing more.
func (u *CleanUpOldRecords) Execute(ctx context.Context) error {
	now := u.from.Clock.Now()
	return errors.Join(
		u.from.Requests.RemoveEndedOrExpiredBefore(ctx, now.Add(-domain.LoginCodeRequestKeptFor)),
		u.from.Asks.RemoveAskedBefore(ctx, now.Add(-domain.AskKeptFor)),
		u.from.EndedSessions.Remove(ctx, now),
		u.from.CleanUpBounceNotices.Execute(ctx),
	)
}
