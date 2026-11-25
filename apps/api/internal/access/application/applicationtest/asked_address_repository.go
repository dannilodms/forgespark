package applicationtest

import (
	"bytes"
	"context"
	"slices"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
)

// Ask is one ask for a login code as the limits keep it: the digest of the
// address and the moment, never the address.
type Ask struct {
	Digest  []byte
	AskedAt time.Time
}

// AskedAddressRepository keeps the asks for login codes in memory.
type AskedAddressRepository struct {
	asks []Ask
}

// NewAskedAddressRepository returns a repository with no ask.
func NewAskedAddressRepository() *AskedAddressRepository {
	return &AskedAddressRepository{}
}

// Hold locks nothing: the tests run one request at a time.
func (r *AskedAddressRepository) Hold(context.Context, domain.AskedAddress) error {
	return nil
}

// CountAskedSince counts the asks for the address after the moment.
func (r *AskedAddressRepository) CountAskedSince(_ context.Context, address domain.AskedAddress, since time.Time) (int, error) {
	count := 0
	for _, ask := range r.asks {
		count += countIf(bytes.Equal(ask.Digest, address.Digest()) && ask.AskedAt.After(since))
	}
	return count, nil
}

func countIf(wanted bool) int {
	if wanted {
		return 1
	}
	return 0
}

// Add keeps one ask for the address at the moment.
func (r *AskedAddressRepository) Add(_ context.Context, address domain.AskedAddress, at time.Time) error {
	r.asks = append(r.asks, Ask{Digest: address.Digest(), AskedAt: at})
	return nil
}

// RemoveAskedBefore removes the asks made before the moment.
func (r *AskedAddressRepository) RemoveAskedBefore(_ context.Context, before time.Time) error {
	r.asks = slices.DeleteFunc(r.asks, func(ask Ask) bool { return ask.AskedAt.Before(before) })
	return nil
}

// Snapshot lets the repository take part in the in-memory transactions.
func (r *AskedAddressRepository) Snapshot() (restore func()) {
	saved := slices.Clone(r.asks)
	return func() { r.asks = saved }
}

// All returns every ask kept, in the order they were made.
func (r *AskedAddressRepository) All() []Ask {
	return slices.Clone(r.asks)
}
