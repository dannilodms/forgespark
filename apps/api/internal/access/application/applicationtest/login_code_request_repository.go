package applicationtest

import (
	"context"
	"slices"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
)

// LoginCodeRequestRepository keeps the login code requests in memory. It
// stores a copy of what it is given and hands out copies, as the database
// does.
type LoginCodeRequestRepository struct {
	lastIdentifier int64
	requests       []domain.LoginCodeRequest
	failure        error
}

// NewLoginCodeRequestRepository returns a repository with no request.
func NewLoginCodeRequestRepository() *LoginCodeRequestRepository {
	return &LoginCodeRequestRepository{}
}

// FailWith makes every later Save fail with this error, as a database that
// refuses the write.
func (r *LoginCodeRequestRepository) FailWith(failure error) {
	r.failure = failure
}

// NextIdentifier reserves the next identifier, starting at 1.
func (r *LoginCodeRequestRepository) NextIdentifier(context.Context) (domain.LoginCodeRequestID, error) {
	r.lastIdentifier++
	return domain.NewLoginCodeRequestID(r.lastIdentifier)
}

// LoadNotEndedOf returns a copy of each request of the user that was not
// ended.
func (r *LoginCodeRequestRepository) LoadNotEndedOf(_ context.Context, user domain.UserID) ([]*domain.LoginCodeRequest, error) {
	found := []*domain.LoginCodeRequest{}
	for _, stored := range r.ofUser(user) {
		_, ended := stored.EndedAt()
		found = appendIf(found, &stored, !ended)
	}
	return found, nil
}

// ofUser returns a copy of the requests of the user.
func (r *LoginCodeRequestRepository) ofUser(user domain.UserID) []domain.LoginCodeRequest {
	return slices.DeleteFunc(slices.Clone(r.requests), func(stored domain.LoginCodeRequest) bool { return stored.User() != user })
}

func appendIf(requests []*domain.LoginCodeRequest, request *domain.LoginCodeRequest, wanted bool) []*domain.LoginCodeRequest {
	if !wanted {
		return requests
	}
	return append(requests, request)
}

// Find returns a copy of the request with this identifier.
func (r *LoginCodeRequestRepository) Find(_ context.Context, id domain.LoginCodeRequestID) (*domain.LoginCodeRequest, bool, error) {
	position := slices.IndexFunc(r.requests, func(stored domain.LoginCodeRequest) bool { return stored.ID() == id })
	if position < 0 {
		return nil, false, nil
	}
	found := r.requests[position]
	return &found, true, nil
}

// Load returns a copy of the request with this identifier. Memory has no
// lock: the tests run one request at a time.
func (r *LoginCodeRequestRepository) Load(ctx context.Context, id domain.LoginCodeRequestID) (*domain.LoginCodeRequest, bool, error) {
	return r.Find(ctx, id)
}

// Save stores the request, in place of the one with the same identifier when
// there is one.
func (r *LoginCodeRequestRepository) Save(_ context.Context, request *domain.LoginCodeRequest) error {
	if r.failure != nil {
		return r.failure
	}
	position := slices.IndexFunc(r.requests, func(stored domain.LoginCodeRequest) bool { return stored.ID() == request.ID() })
	if position < 0 {
		r.requests = append(r.requests, *request)
		return nil
	}
	r.requests[position] = *request
	return nil
}

// RemoveEndedOrExpiredBefore removes the requests that ended, or expired,
// before the moment.
func (r *LoginCodeRequestRepository) RemoveEndedOrExpiredBefore(_ context.Context, before time.Time) error {
	r.requests = slices.DeleteFunc(r.requests, func(stored domain.LoginCodeRequest) bool {
		endedAt, ended := stored.EndedAt()
		return stored.ExpiresAt().Before(before) || (ended && endedAt.Before(before))
	})
	return nil
}

// Snapshot lets the repository take part in the in-memory transactions.
func (r *LoginCodeRequestRepository) Snapshot() (restore func()) {
	saved := slices.Clone(r.requests)
	return func() { r.requests = saved }
}

// All returns every request stored, in the order they were first saved.
func (r *LoginCodeRequestRepository) All() []domain.LoginCodeRequest {
	return slices.Clone(r.requests)
}
