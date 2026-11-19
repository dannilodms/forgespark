package applicationtest

import "context"

// Store is an in-memory implementation that can take part in a transaction:
// it gives a function that puts it back as it is now. The in-memory e-mail
// records and jobs of the e-mail area are stores too.
type Store interface {
	Snapshot() (restore func())
}

// Transactions stands in for the transactions of the database. An operation
// that fails leaves every store as it was before the operation started, as a
// rollback does. An operation inside another one, such as queueing an e-mail
// inside asking for a login code, joins it, as the transactions of the
// database do: nothing is undone until the outermost one fails.
type Transactions struct {
	stores []Store
	open   bool
}

// NewTransactions returns transactions over these stores: every in-memory
// implementation the use case under test writes to, of any area.
func NewTransactions(stores ...Store) *Transactions {
	return &Transactions{stores: stores}
}

// Run calls the operation and, when it fails, undoes what it wrote.
func (t *Transactions) Run(ctx context.Context, operation func(ctx context.Context) error) error {
	if t.open {
		return operation(ctx)
	}
	restores := make([]func(), 0, len(t.stores))
	for _, store := range t.stores {
		restores = append(restores, store.Snapshot())
	}

	t.open = true
	err := operation(ctx)
	t.open = false
	if err == nil {
		return nil
	}
	for _, restore := range restores {
		restore()
	}
	return err
}
