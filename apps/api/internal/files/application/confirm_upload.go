package application

import (
	"context"
	"fmt"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
)

// ConfirmUpload checks the object a browser sent to the storage and makes its
// file usable. An object over the size limit, or one that is not an image of
// the three types, is removed from the storage, and its file stays
// unconfirmed. The caller checks whether the confirmation is allowed.
type ConfirmUpload struct {
	transactions Transactions
	files        FileRepository
	storage      Storage
	clock        Clock
}

// NewConfirmUpload builds the use case.
func NewConfirmUpload(transactions Transactions, files FileRepository, storage Storage, clock Clock) *ConfirmUpload {
	return &ConfirmUpload{transactions: transactions, files: files, storage: storage, clock: clock}
}

// ConfirmUploadInput names the file and what it is given for.
type ConfirmUploadInput struct {
	FileID  FileID
	Purpose Purpose
}

// Execute confirms the file. It answers FileNotFoundError when no file of the
// purpose has the identifier, FileNotUploadedError when its object is not in
// the storage, and FileTooLargeError or FileTypeUnsupportedError when the
// object breaks a rule of an image, after removing it. A file already
// confirmed is checked again, and stays confirmed.
//
// The file is checked in a transaction, which locks it. Removing the object
// of a refused file is an effect on the storage, so it runs once that
// transaction is committed (Writes in CODING_STANDARDS.md).
func (u *ConfirmUpload) Execute(ctx context.Context, input ConfirmUploadInput) error {
	id, err := domain.NewFileID(int64(input.FileID))
	if err != nil {
		return domain.FileNotFoundError{}
	}
	now := u.clock.Now()

	// The object of a refused file, and the rule it broke.
	var refused domain.ObjectKey
	var broken error
	err = u.transactions.Run(ctx, func(ctx context.Context) error {
		file, found, err := u.files.Load(ctx, id)
		if err != nil {
			return err
		}
		if !found || !file.IsFor(input.Purpose) {
			return domain.FileNotFoundError{}
		}
		broken, err = u.confirm(ctx, file, now)
		if err != nil {
			return err
		}
		if broken != nil {
			// Nothing of the file changed: it stays unconfirmed.
			refused = file.ObjectKey()
			return nil
		}
		return u.files.Save(ctx, file)
	})
	if err != nil {
		return err
	}
	if broken == nil {
		return nil
	}
	return u.remove(ctx, refused, broken)
}

// confirm reads the object of the file and tells the file what it holds. It
// returns the rule the object breaks, when it breaks one, apart from a failure
// of the system.
func (u *ConfirmUpload) confirm(ctx context.Context, file *domain.File, now time.Time) (broken, failed error) {
	object, found, err := u.storage.Inspect(ctx, file.ObjectKey())
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.FileNotUploadedError{}
	}
	return file.Confirm(object, now), nil
}

// remove removes the object of a file that was refused, which nothing will
// ever use, and answers the rule it broke.
func (u *ConfirmUpload) remove(ctx context.Context, key domain.ObjectKey, broken error) error {
	if err := u.storage.Remove(ctx, key); err != nil {
		// A failure of the system, not a rule: the object is still there.
		return fmt.Errorf("remove the object of a file that was refused (%v): %w", broken, err)
	}
	return broken
}
