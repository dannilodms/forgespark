package application

import (
	"context"

	"github.com/dannilodms/forgespark/apps/api/internal/email/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// QueueEmail creates the record of an e-mail and saves the job that delivers
// it, both in the transaction of the caller: no e-mail is recorded and never
// sent, or sent and never recorded. Another area calls it from its own use
// case, inside its own transaction.
type QueueEmail struct {
	transactions Transactions
	emails       EmailRepository
	deliveries   DeliveryJobs
	clock        Clock
}

// NewQueueEmail builds the use case.
func NewQueueEmail(transactions Transactions, emails EmailRepository, deliveries DeliveryJobs, clock Clock) *QueueEmail {
	return &QueueEmail{transactions: transactions, emails: emails, deliveries: deliveries, clock: clock}
}

// QueueEmailInput is the e-mail to send: to whom, in which template and
// language, and what it refers to. The content comes later, when the e-mail
// is delivered, from the record it refers to.
type QueueEmailInput struct {
	Recipient string
	Template  Template
	// Language is the preferred language of the recipient, or the default
	// language when it is not known.
	Language  kernel.Language
	Reference Reference
}

// Execute queues the e-mail. It answers RecipientInvalidError when the
// recipient is not an address, and saves nothing then.
func (u *QueueEmail) Execute(ctx context.Context, input QueueEmailInput) error {
	recipient, err := domain.NewRecipient(input.Recipient)
	if err != nil {
		return err
	}
	now := u.clock.Now()
	return u.transactions.Run(ctx, func(ctx context.Context) error {
		id, err := u.emails.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		email, err := domain.Queue(id, recipient, input.Template, input.Language, input.Reference, now)
		if err != nil {
			return err
		}
		if err := u.emails.Save(ctx, email); err != nil {
			return err
		}
		return u.deliveries.Save(ctx, id)
	})
}
