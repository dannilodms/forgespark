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
	stores        []Store
	commitFailure error
}

// NewTransactions returns transactions over these stores: every in-memory
// implementation the use case under test writes to.
func NewTransactions(stores ...Store) *Transactions {
	return &Transactions{stores: stores}
}

// FailCommitWith makes every later commit fail with this error, after the
// operation succeeded, as a database that loses its connection at the end:
// what the operation wrote is undone.
func (t *Transactions) FailCommitWith(failure error) {
	t.commitFailure = failure
}

// Run calls the operation and, when it fails or cannot be committed, undoes
// what it wrote.
func (t *Transactions) Run(ctx context.Context, operation func(ctx context.Context) error) error {
	restores := make([]func(), 0, len(t.stores))
	for _, store := range t.stores {
		restores = append(restores, store.Snapshot())
	}

	err := operation(ctx)
	if err == nil {
		err = t.commitFailure
	}
	if err == nil {
		return nil
	}
	for _, restore := range restores {
		restore()
	}
	return err
}
