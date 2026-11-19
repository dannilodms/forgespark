// Package domain holds the e-mail record, the templates of the e-mails the
// system sends and their texts in every supported language. The system sends
// e-mail and receives none (E-mail in CODING_STANDARDS.md). The address of a
// recipient is in the record and nowhere else: never in a job, a log or a
// span.
package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// EmailID identifies an e-mail record.
type EmailID int64

// NewEmailID reads an identifier that came from the sequence of e-mails.
func NewEmailID(value int64) (EmailID, error) {
	if value <= 0 {
		return 0, errors.New("the identifier of an e-mail is a positive number")
	}
	return EmailID(value), nil
}

// Recipient is the address an e-mail goes to: one address and nothing around
// it, no name and no angle brackets.
type Recipient struct {
	address string
}

// NewRecipient reads the address of a recipient, as it is written, by the
// rule of an e-mail address of kernel.IsEmailAddress. Its error never
// repeats the address.
func NewRecipient(address string) (Recipient, error) {
	if !kernel.IsEmailAddress(address) {
		return Recipient{}, RecipientInvalidError{}
	}
	return Recipient{address: address}, nil
}

// String returns the address.
func (r Recipient) String() string {
	return r.address
}

// ReferenceKind tells what kind of record an e-mail refers to.
type ReferenceKind string

const (
	// LoginCodeRequestKind is a login code request: the e-mail carries its
	// code.
	LoginCodeRequestKind ReferenceKind = "login_code_request"
)

// Reference is what an e-mail refers to: the record of another area that
// asked for it, by its kind and its identifier. The area that holds the
// record gives the content of the e-mail from it.
type Reference struct {
	kind       ReferenceKind
	identifier int64
}

// LoginCodeRequest is the login code request with this identifier.
func LoginCodeRequest(identifier int64) Reference {
	return Reference{kind: LoginCodeRequestKind, identifier: identifier}
}

// RestoreReference rebuilds a reference from what an adapter stored.
func RestoreReference(kind ReferenceKind, identifier int64) Reference {
	return Reference{kind: kind, identifier: identifier}
}

// Kind tells what kind of record the e-mail refers to.
func (r Reference) Kind() ReferenceKind { return r.kind }

// Identifier returns the identifier of the record the e-mail refers to.
func (r Reference) Identifier() int64 { return r.identifier }

func (r Reference) valid() bool {
	return r.kind == LoginCodeRequestKind && r.identifier > 0
}

// Status is where an e-mail is in its life: queued until the job delivers
// it, then sent, or failed when the attempts run out. A bounce its sender
// reports for a sent e-mail makes it bounced.
type Status string

// The statuses of an e-mail.
const (
	Queued  Status = "queued"
	Sent    Status = "sent"
	Failed  Status = "failed"
	Bounced Status = "bounced"
)

// Operation is what can happen to an e-mail record.
type Operation string

// The operations on an e-mail record.
const (
	// MarkSentOperation records that the sender took the e-mail.
	MarkSentOperation Operation = "mark_sent"
	// MarkFailedOperation records that the last attempt to deliver the e-mail
	// failed.
	MarkFailedOperation Operation = "mark_failed"
	// MarkBouncedOperation records that the server of the recipient refused
	// for good an e-mail the sender took, as the sender reported.
	MarkBouncedOperation Operation = "mark_bounced"
)

// transitions is the transition table of an e-mail (ADR-0022): for each
// operation, the statuses it starts from and the status it leads to.
var transitions = map[Operation]struct {
	from []Status
	to   Status
}{
	MarkSentOperation:    {from: []Status{Queued}, to: Sent},
	MarkFailedOperation:  {from: []Status{Queued}, to: Failed},
	MarkBouncedOperation: {from: []Status{Sent}, to: Bounced},
}

func (s Status) after(operation Operation) (Status, error) {
	transition, known := transitions[operation]
	if !known || !slices.Contains(transition.from, s) {
		return "", OperationNotAllowedError{Operation: operation, Status: s}
	}
	return transition.to, nil
}

// FailureCode is the stable code of why an e-mail failed.
type FailureCode string

// The codes of a failed e-mail.
const (
	// ContentUnavailable is an e-mail whose content the area that asked for
	// it could not give.
	ContentUnavailable FailureCode = "content_unavailable"
	// SendingFailed is an e-mail the sender did not take: the e-mail server
	// refused it or could not be reached.
	SendingFailed FailureCode = "sending_failed"
	// DeliveryFailed is an e-mail that failed for any other reason.
	DeliveryFailed FailureCode = "delivery_failed"
)

// Email is the record of one e-mail the system sends: to whom, its subject,
// its template and language, what it refers to and where it is in its life.
// It is created queued, with the job that delivers it.
type Email struct {
	id        EmailID
	recipient Recipient
	subject   string
	template  Template
	language  kernel.Language
	reference Reference
	status    Status
	queuedAt  time.Time
	messageID string
	sentAt    *time.Time
	failedAt  *time.Time
	failure   FailureCode
	bouncedAt *time.Time
}

// Queue creates the record of an e-mail about to be delivered, in the
// language of its recipient. Its subject is written now, from the template.
func Queue(id EmailID, recipient Recipient, template Template, language kernel.Language, reference Reference, at time.Time) (*Email, error) {
	if !template.known() {
		return nil, fmt.Errorf("no template of e-mail is named %q", template)
	}
	if language.String() == "" {
		return nil, errors.New("an e-mail is written in a supported language")
	}
	if !reference.valid() {
		return nil, fmt.Errorf("an e-mail refers to a record of a known kind, not %q %d", reference.kind, reference.identifier)
	}
	subject, err := template.subjectIn(language)
	if err != nil {
		return nil, err
	}
	return &Email{
		id:        id,
		recipient: recipient,
		subject:   subject,
		template:  template,
		language:  language,
		reference: reference,
		status:    Queued,
		queuedAt:  at,
	}, nil
}

// StoredEmail is an e-mail record as an adapter stored it.
type StoredEmail struct {
	ID        EmailID
	Recipient string
	Subject   string
	Template  Template
	Language  kernel.Language
	Reference Reference
	Status    Status
	QueuedAt  time.Time
	MessageID string
	SentAt    *time.Time
	FailedAt  *time.Time
	Failure   FailureCode
	BouncedAt *time.Time
}

// RestoreEmail rebuilds an e-mail from what an adapter stored. It checks
// nothing that was checked when the e-mail was saved.
func RestoreEmail(stored StoredEmail) *Email {
	return &Email{
		id:        stored.ID,
		recipient: Recipient{address: stored.Recipient},
		subject:   stored.Subject,
		template:  stored.Template,
		language:  stored.Language,
		reference: stored.Reference,
		status:    stored.Status,
		queuedAt:  stored.QueuedAt,
		messageID: stored.MessageID,
		sentAt:    stored.SentAt,
		failedAt:  stored.FailedAt,
		failure:   stored.Failure,
		bouncedAt: stored.BouncedAt,
	}
}

// IsQueued tells whether the e-mail still waits to be delivered. A job that
// runs again on an e-mail that is no longer queued delivers nothing.
func (e *Email) IsQueued() bool {
	return e.status == Queued
}

// Message writes the e-mail, ready to leave, in the language of its
// recipient, with the content of its template.
func (e *Email) Message(content Content) (Message, error) {
	if content.Template() != e.template {
		return Message{}, fmt.Errorf("the e-mail %d is a %s, and the content is of a %s", e.id, e.template, content.Template())
	}
	parts, err := e.template.write(e.language, content)
	if err != nil {
		return Message{}, err
	}
	return Message{EmailID: e.id, To: e.recipient.String(), Subject: e.subject, HTML: parts.html, Text: parts.text}, nil
}

// MarkSent records that the sender took the e-mail, with the identifier it
// gave the message.
func (e *Email) MarkSent(messageID string, at time.Time) error {
	next, err := e.status.after(MarkSentOperation)
	if err != nil {
		return err
	}
	e.status = next
	e.messageID = messageID
	e.sentAt = &at
	return nil
}

// MarkFailed records that the last attempt to deliver the e-mail failed, and
// why.
func (e *Email) MarkFailed(failure FailureCode, at time.Time) error {
	next, err := e.status.after(MarkFailedOperation)
	if err != nil {
		return err
	}
	e.status = next
	e.failure = failure
	e.failedAt = &at
	return nil
}

// MarkBounced records that the e-mail sent as this message bounced, as its
// sender reported. A report of another message, one the e-mail was not sent
// as, is refused with BounceOfAnotherMessageError.
func (e *Email) MarkBounced(messageID string, at time.Time) error {
	next, err := e.status.after(MarkBouncedOperation)
	if err != nil {
		return err
	}
	if messageID != e.messageID {
		return BounceOfAnotherMessageError{}
	}
	e.status = next
	e.bouncedAt = &at
	return nil
}

// ID returns the identifier of the e-mail.
func (e *Email) ID() EmailID { return e.id }

// Recipient returns the address the e-mail goes to.
func (e *Email) Recipient() Recipient { return e.recipient }

// Subject returns the subject of the e-mail.
func (e *Email) Subject() string { return e.subject }

// Template returns the template of the e-mail.
func (e *Email) Template() Template { return e.template }

// Language returns the language the e-mail is written in.
func (e *Email) Language() kernel.Language { return e.language }

// Reference returns what the e-mail refers to.
func (e *Email) Reference() Reference { return e.reference }

// Status returns where the e-mail is in its life.
func (e *Email) Status() Status { return e.status }

// QueuedAt returns when the e-mail was queued.
func (e *Email) QueuedAt() time.Time { return e.queuedAt }

// MessageID returns the identifier the sender gave the message, and false
// while it was not sent.
func (e *Email) MessageID() (string, bool) {
	return e.messageID, e.sentAt != nil
}

// SentAt returns when the sender took the e-mail, and false while it did not.
func (e *Email) SentAt() (time.Time, bool) {
	if e.sentAt == nil {
		return time.Time{}, false
	}
	return *e.sentAt, true
}

// Failure returns when the e-mail failed and why, and false while it did not.
func (e *Email) Failure() (FailureCode, time.Time, bool) {
	if e.failedAt == nil {
		return "", time.Time{}, false
	}
	return e.failure, *e.failedAt, true
}

// BouncedAt returns when the sender reported that the e-mail bounced, and
// false while it did not.
func (e *Email) BouncedAt() (time.Time, bool) {
	if e.bouncedAt == nil {
		return time.Time{}, false
	}
	return *e.bouncedAt, true
}

// Message is an e-mail ready to leave: its recipient, its subject and its two
// parts. EmailID names its record, for a sender that takes an idempotency
// key.
type Message struct {
	EmailID EmailID
	To      string
	Subject string
	HTML    string
	Text    string
}
