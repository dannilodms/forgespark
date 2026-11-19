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

// Send keeps the message.
func (s *Sender) Send(_ context.Context, message domain.Message) (string, error) {
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
