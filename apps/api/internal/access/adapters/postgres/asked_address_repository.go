package postgres

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// askLockClass is the first number of the key of the advisory lock that holds
// the asks for an address; the second is taken from the digest of the
// address. Two addresses whose digests share their first four bytes wait for
// each other, which costs a moment and counts nothing wrong.
const askLockClass = 70_011

// AskedAddressRepository keeps the asks for login codes in the
// login_code_asks table, by the digest of each address.
type AskedAddressRepository struct {
	pool *pgxpool.Pool
}

// NewAskedAddressRepository returns the repository on this pool.
func NewAskedAddressRepository(pool *pgxpool.Pool) *AskedAddressRepository {
	return &AskedAddressRepository{pool: pool}
}

func (r *AskedAddressRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// Hold takes the advisory lock of the address until the transaction ends.
// Outside a transaction it would end with the statement, so it holds nothing
// there.
func (r *AskedAddressRepository) Hold(ctx context.Context, address domain.AskedAddress) error {
	err := r.queries(ctx).HoldLoginCodeAsks(ctx, HoldLoginCodeAsksParams{
		LockClass: askLockClass,
		LockKey:   int32(binary.BigEndian.Uint32(address.Digest())), //nolint:gosec // The four bytes are a key, and any value of them is one.
	})
	if err != nil {
		return fmt.Errorf("hold the asks for a login code of an address: %w", err)
	}
	return nil
}

// CountAskedSince counts the asks for the address after the moment.
func (r *AskedAddressRepository) CountAskedSince(ctx context.Context, address domain.AskedAddress, since time.Time) (int, error) {
	count, err := r.queries(ctx).CountLoginCodeAsksSince(ctx, CountLoginCodeAsksSinceParams{Address: address.Digest(), Since: since})
	if err != nil {
		return 0, fmt.Errorf("count the asks for a login code of an address: %w", err)
	}
	return int(count), nil
}

// Add keeps one ask for the address at the moment.
func (r *AskedAddressRepository) Add(ctx context.Context, address domain.AskedAddress, at time.Time) error {
	if err := r.queries(ctx).AddLoginCodeAsk(ctx, AddLoginCodeAskParams{Address: address.Digest(), AskedAt: at}); err != nil {
		return fmt.Errorf("keep an ask for a login code: %w", err)
	}
	return nil
}

// RemoveAskedBefore removes the asks made before the moment.
func (r *AskedAddressRepository) RemoveAskedBefore(ctx context.Context, before time.Time) error {
	if err := r.queries(ctx).RemoveLoginCodeAsksBefore(ctx, before); err != nil {
		return fmt.Errorf("remove the old asks for a login code: %w", err)
	}
	return nil
}
