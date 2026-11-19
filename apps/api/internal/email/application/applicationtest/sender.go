package applicationtest

import (
	"context"
	"fmt"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
)

// Sender keeps the messages it is handed, as a mailbox that receives them,
// and gives each the identifier `message-<n>`.
type Sender struct {
	sent    []domain.Message
	failure error
	hangs   bool
}

// NewSender returns a sender that took nothing yet.
func NewSender() *Sender {
	return &Sender{}
}

// FailWith makes every later Send fail with this error, as an e-mail server
// that cannot be reached, until it is called with nil.
func (s *Sender) FailWith(failure error) {
	s.failure = failure
}

// HangUntilTheContextEnds makes every later Send wait until its context ends
// and fail with its error, as an e-mail server that never answers, until the
// job runs out of its time or is cut.
func (s *Sender) HangUntilTheContextEnds() {
	s.hangs = true
}

// Send keeps the message.
func (s *Sender) Send(ctx context.Context, message domain.Message) (string, error) {
	if s.hangs {
		<-ctx.Done()
		return "", ctx.Err()
	}
	if s.failure != nil {
		return "", s.failure
	}
	s.sent = append(s.sent, message)
	return fmt.Sprintf("message-%d", len(s.sent)), nil
}

// Sent returns every message taken, in order.
func (s *Sender) Sent() []domain.Message {
	return slices.Clone(s.sent)
}
