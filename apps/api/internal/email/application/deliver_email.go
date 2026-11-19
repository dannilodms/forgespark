package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
)

// DeliverEmail writes a queued e-mail in the language of its recipient, with
// the content the area that asked for it gives, hands it to the sender and
// marks it sent. The job saved with the e-mail runs it, and runs it again
// when it fails, until the attempts run out: the last failure marks the
// e-mail failed. An e-mail that is no longer queued is not sent again.
type DeliverEmail struct {
	transactions Transactions
	emails       EmailRepository
	contents     ContentSources
	sender       Sender
	clock        Clock
}

// NewDeliverEmail builds the use case. The contents give the content of each
// template, and the sender is the one the configuration chose.
func NewDeliverEmail(transactions Transactions, emails EmailRepository, contents ContentSources, sender Sender, clock Clock) *DeliverEmail {
	return &DeliverEmail{transactions: transactions, emails: emails, contents: contents, sender: sender, clock: clock}
}

// DeliverEmailInput names the e-mail and tells whether this attempt is the
// last one.
type DeliverEmailInput struct {
	EmailID EmailID
	// LastAttempt tells that no attempt follows this one: if it fails, the
	// e-mail is marked failed.
	LastAttempt bool
}

// Execute delivers the e-mail, and returns why it could not. The e-mail is
// loaded and locked, so two runs of its job never both send it, and it is
// sent inside that transaction: when the sender took it but the transaction
// could not be committed, the next attempt sends it again, which a sender
// with an idempotency key turns into one e-mail.
//
// On the last attempt a failure marks the e-mail failed, with the code of
// what failed, and is still returned, so the job ends as failed too. A run
// cut by the shutdown is not a failure: the job runs again later.
func (u *DeliverEmail) Execute(ctx context.Context, input DeliverEmailInput) error {
	err := u.transactions.Run(ctx, func(ctx context.Context) error {
		return u.deliver(ctx, input.EmailID)
	})
	if err == nil || !input.LastAttempt || ctx.Err() != nil {
		return err
	}
	return errors.Join(err, u.markFailed(ctx, input.EmailID, failureCodeOf(err)))
}

func (u *DeliverEmail) deliver(ctx context.Context, id domain.EmailID) error {
	email, found, err := u.emails.Load(ctx, id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no e-mail %d exists", id)
	}
	if !email.IsQueued() {
		return nil
	}
	content, err := u.contentOf(ctx, email)
	if err != nil {
		return failure{code: domain.ContentUnavailable, err: fmt.Errorf("give the content of the e-mail %d: %w", id, err)}
	}
	message, err := email.Message(content)
	if err != nil {
		return err
	}
	messageID, err := u.sender.Send(ctx, message)
	if err != nil {
		return failure{code: domain.SendingFailed, err: fmt.Errorf("send the e-mail %d: %w", id, err)}
	}
	if err := email.MarkSent(messageID, u.clock.Now()); err != nil {
		return err
	}
	return u.emails.Save(ctx, email)
}

func (u *DeliverEmail) contentOf(ctx context.Context, email *domain.Email) (domain.Content, error) {
	source, known := u.contents[email.Template()]
	if !known {
		return nil, fmt.Errorf("nothing gives the content of the template %s", email.Template())
	}
	return source.ContentOf(ctx, email.Reference())
}

// markFailed marks the e-mail failed in a transaction of its own, since the
// one of the delivery was rolled back. An e-mail that is no longer queued
// stays as it is.
func (u *DeliverEmail) markFailed(ctx context.Context, id domain.EmailID, code domain.FailureCode) error {
	return u.transactions.Run(ctx, func(ctx context.Context) error {
		email, found, err := u.emails.Load(ctx, id)
		if err != nil || !found || !email.IsQueued() {
			return err
		}
		if err := email.MarkFailed(code, u.clock.Now()); err != nil {
			return err
		}
		return u.emails.Save(ctx, email)
	})
}

// failure is what failed in a delivery, with the code a failed e-mail
// records for it.
type failure struct {
	code domain.FailureCode
	err  error
}

func (f failure) Error() string { return f.err.Error() }

func (f failure) Unwrap() error { return f.err }

func failureCodeOf(err error) domain.FailureCode {
	var known failure
	if errors.As(err, &known) {
		return known.code
	}
	return domain.DeliveryFailed
}
