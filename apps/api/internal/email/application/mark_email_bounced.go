package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
)

// MarkEmailBounced records a bounce the sender reported: the e-mail it took
// was refused for good by the server of the recipient. The inbound adapter of
// the sender runs it for each notice whose signature it verified.
//
// A notice is taken once: its identifier is kept in the transaction that
// marks the e-mail, and a notice delivered again changes nothing. A notice
// that names no e-mail of the system, a message the e-mail was not sent as,
// or an e-mail that failed or already bounced is taken and changes nothing
// either.
type MarkEmailBounced struct {
	transactions Transactions
	emails       EmailRepository
	notices      BounceNotices
	clock        Clock
}

// NewMarkEmailBounced builds the use case.
func NewMarkEmailBounced(transactions Transactions, emails EmailRepository, notices BounceNotices, clock Clock) *MarkEmailBounced {
	return &MarkEmailBounced{transactions: transactions, emails: emails, notices: notices, clock: clock}
}

// MarkEmailBouncedInput is one bounce notice of the sender.
type MarkEmailBouncedInput struct {
	// Notice identifies the notice, as its sender gave it. A notice the
	// sender delivers again keeps its identifier.
	Notice string
	// EmailID names the e-mail that bounced, as the sender was told when it
	// took it.
	EmailID EmailID
	// MessageID is the identifier the sender gave the message that bounced.
	MessageID string
}

// errNoNotice is a notice with no identifier, which could not be taken once.
var errNoNotice = errors.New("a bounce notice has an identifier")

// Execute takes the notice. It fails, and keeps nothing, when the e-mail is
// still queued: the notice came before the delivery marked the e-mail sent,
// and the sender delivers it again later.
func (u *MarkEmailBounced) Execute(ctx context.Context, input MarkEmailBouncedInput) error {
	if input.Notice == "" {
		return errNoNotice
	}
	return u.transactions.Run(ctx, func(ctx context.Context) error {
		return u.markBounced(ctx, input)
	})
}

func (u *MarkEmailBounced) markBounced(ctx context.Context, input MarkEmailBouncedInput) error {
	now := u.clock.Now()
	first, err := u.notices.Keep(ctx, input.Notice, now)
	if err != nil || !first {
		return err
	}
	email, found, err := u.emails.Load(ctx, input.EmailID)
	if err != nil || !found {
		return err
	}
	err = email.MarkBounced(input.MessageID, now)
	if stillQueued(err) {
		return fmt.Errorf("the e-mail %d is not marked sent yet: %w", input.EmailID, err)
	}
	if nothingToMark(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return u.emails.Save(ctx, email)
}

// stillQueued tells a bounce refused because the e-mail is still queued.
func stillQueued(err error) bool {
	var refused domain.OperationNotAllowedError
	return errors.As(err, &refused) && refused.Status == domain.Queued
}

// nothingToMark tells a bounce of an e-mail that failed or already bounced,
// or of a message the e-mail was not sent as.
func nothingToMark(err error) bool {
	var refused domain.OperationNotAllowedError
	var another domain.BounceOfAnotherMessageError
	return errors.As(err, &refused) || errors.As(err, &another)
}
