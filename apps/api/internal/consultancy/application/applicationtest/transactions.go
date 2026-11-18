package applicationtest

import "context"

// Store is an in-memory implementation that can take part in a transaction:
// it gives a function that puts it back as it is now.
type Store interface {
	Snapshot() (restore func())
}

// Transactions stands in for the transactions of the database. An operation
// that fails leaves every store as it was before the operation started, as a
// rollback does.
type Transactions struct {
	stores []Store
}

// NewTransactions returns transactions over these stores: every in-memory
// implementation the use case under test writes to, the ones of other areas
// included.
func NewTransactions(stores ...Store) *Transactions {
	return &Transactions{stores: stores}
}

// Run calls the operation and, when it fails, undoes what it wrote.
func (t *Transactions) Run(ctx context.Context, operation func(ctx context.Context) error) error {
	restores := make([]func(), 0, len(t.stores))
	for _, store := range t.stores {
		restores = append(restores, store.Snapshot())
	}

	err := operation(ctx)
	if err == nil {
		return nil
	}
	for _, restore := range restores {
		restore()
	}
	return err
}
