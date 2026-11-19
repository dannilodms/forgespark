// Package application holds the use cases of the e-mail area and the
// interfaces they need. Another area queues an e-mail inside its own
// transaction with QueueEmail, and the job saved with it runs DeliverEmail,
// which writes the e-mail with the content that area gives and hands it to
// the sender the configuration chose. A bounce the sender reports runs
// MarkEmailBounced.
package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
)

// Transactions opens and closes the transaction of a use case that writes.
type Transactions interface {
	// Run calls the operation inside one transaction, and passes it the
	// context every repository must receive to take part. The transaction is
	// committed when the operation returns nil and rolled back, with nothing
	// saved, when it returns an error. Inside an operation of another area it
	// joins the transaction already open.
	Run(ctx context.Context, operation func(ctx context.Context) error) error
}

// Clock tells the current time. A use case never reads the time of the
// machine by itself, so a test decides what time it is.
type Clock interface {
	Now() time.Time
}

// EmailRepository keeps the e-mail records.
type EmailRepository interface {
	// NextIdentifier reserves the identifier of an e-mail that does not exist
	// yet.
	NextIdentifier(ctx context.Context) (domain.EmailID, error)
	// Load loads the e-mail with this identifier, and false when there is
	// none. It locks the e-mail until the transaction ends.
	Load(ctx context.Context, id domain.EmailID) (*domain.Email, bool, error)
	// Save stores an e-mail whole, whether it is new or changed.
	Save(ctx context.Context, email *domain.Email) error
}

// DeliveryJobs saves the job that delivers an e-mail.
type DeliveryJobs interface {
	// Save saves the job of this e-mail in the transaction the context
	// carries, so the job exists only if the e-mail does. Its arguments are
	// the identifier of the e-mail and nothing else.
	Save(ctx context.Context, id domain.EmailID) error
}

// Sender hands an e-mail to the e-mail server or service the configuration
// chose.
type Sender interface {
	// Send sends the message and returns the identifier the sender gave it.
	// Its error never carries the address of the recipient.
	Send(ctx context.Context, message domain.Message) (messageID string, err error)
}

// BounceNotices keeps the identifiers of the bounce notices taken, so a
// notice its sender delivers again changes nothing.
type BounceNotices interface {
	// Keep keeps the identifier of a notice in the transaction the context
	// carries, and tells false when it was kept before.
	Keep(ctx context.Context, notice string, receivedAt time.Time) (bool, error)
}

// ContentSource gives the content of the e-mails of one template, from the
// record each one refers to, when the e-mail is delivered. The area that asks
// for the e-mails of the template implements it.
type ContentSource interface {
	ContentOf(ctx context.Context, reference domain.Reference) (domain.Content, error)
}

// ContentSources gives, for each template, the source of its content.
type ContentSources map[domain.Template]ContentSource
