package applicationtest

import (
	"context"
	"fmt"

	"github.com/dannilodms/forgespark/apps/api/internal/email/domain"
)

// ContentSource gives a fixed content for each reference it was told about,
// as the area that asked for the e-mails would.
type ContentSource struct {
	contents map[domain.Reference]domain.Content
	failure  error
}

// NewContentSource returns a source that knows no reference.
func NewContentSource() *ContentSource {
	return &ContentSource{contents: map[domain.Reference]domain.Content{}}
}

// Give makes the source give this content for the reference.
func (s *ContentSource) Give(reference domain.Reference, content domain.Content) {
	s.contents[reference] = content
}

// FailWith makes every later ContentOf fail with this error.
func (s *ContentSource) FailWith(failure error) {
	s.failure = failure
}

// ContentOf gives the content of the reference.
func (s *ContentSource) ContentOf(_ context.Context, reference domain.Reference) (domain.Content, error) {
	if s.failure != nil {
		return nil, s.failure
	}
	content, known := s.contents[reference]
	if !known {
		return nil, fmt.Errorf("no %s %d", reference.Kind(), reference.Identifier())
	}
	return content, nil
}
