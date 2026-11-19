package application

import (
	"context"
	"slices"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
)

// CleanUpStorage removes what an upload never confirmed leaves behind: the
// files whose upload started more than a day ago and was never confirmed,
// and the objects of the storage more than a day old that no confirmed file
// names, which a failed removal or a confirmation that did not commit leaves
// too (ADR-0017). A confirmed file and its object are never touched. A
// periodic job runs it, so it can run twice over the same state, or miss a
// run: a second run changes nothing, and a later run removes what the missed
// one would have.
type CleanUpStorage struct {
	transactions Transactions
	files        FileRepository
	storage      Storage
	clock        Clock
}

// NewCleanUpStorage builds the use case.
func NewCleanUpStorage(transactions Transactions, files FileRepository, storage Storage, clock Clock) *CleanUpStorage {
	return &CleanUpStorage{transactions: transactions, files: files, storage: storage, clock: clock}
}

// Execute removes the files of the uploads abandoned, in a transaction, and
// then lists the storage and removes each object left behind. It stops at
// the first failure, and the next run starts over.
//
// The files go first. A confirmation that holds one of them ends before the
// file is removed, or the file is not removed, because it is confirmed. So
// by the time the storage is listed, no confirmation still reads an object of
// an upload abandoned, and the copy a confirmation writes is younger than a
// day: an object it names is never removed.
func (u *CleanUpStorage) Execute(ctx context.Context) error {
	abandonedBefore := domain.AbandonedBefore(u.clock.Now())
	err := u.transactions.Run(ctx, func(ctx context.Context) error {
		return u.files.RemoveAbandoned(ctx, abandonedBefore)
	})
	if err != nil {
		return err
	}
	return u.storage.List(ctx, func(objects []domain.ListedObject) error {
		return u.removeLeftBehind(ctx, objects, abandonedBefore)
	})
}

// removeLeftBehind removes, of one page of the listing, each object old
// enough to be left behind that no confirmed file names.
func (u *CleanUpStorage) removeLeftBehind(ctx context.Context, objects []domain.ListedObject, abandonedBefore time.Time) error {
	old := []domain.ObjectKey{}
	for _, object := range objects {
		old = append(old, keyIfLeftBehind(object, abandonedBefore)...)
	}
	if len(old) == 0 {
		return nil
	}
	named, err := u.files.ConfirmedKeys(ctx, old)
	if err != nil {
		return err
	}
	unnamed := slices.DeleteFunc(old, func(key domain.ObjectKey) bool { return slices.Contains(named, key) })
	for i := 0; err == nil && i < len(unnamed); i++ {
		err = u.storage.Remove(ctx, unnamed[i])
	}
	return err
}

func keyIfLeftBehind(object domain.ListedObject, abandonedBefore time.Time) []domain.ObjectKey {
	if !object.IsLeftBehind(abandonedBefore) {
		return nil
	}
	return []domain.ObjectKey{object.Key}
}
