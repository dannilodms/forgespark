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
func (u *ConfirmUpload) Execute(ctx context.Context, input ConfirmUploadInput) error {
	id, err := domain.NewFileID(int64(input.FileID))
	if err != nil {
		return domain.FileNotFoundError{}
	}
	now := u.clock.Now()

	return u.transactions.Run(ctx, func(ctx context.Context) error {
		file, found, err := u.files.Load(ctx, id)
		if err != nil {
			return err
		}
		if !found || !file.IsFor(input.Purpose) {
			return domain.FileNotFoundError{}
		}
		if err := u.confirm(ctx, file, now); err != nil {
			return err
		}
		return u.files.Save(ctx, file)
	})
}

// confirm reads the object of the file and tells the file what it holds. An
// object the file refuses is removed: nothing else will ever use it.
func (u *ConfirmUpload) confirm(ctx context.Context, file *domain.File, now time.Time) error {
	object, found, err := u.storage.Inspect(ctx, file.ObjectKey())
	if err != nil {
		return err
	}
	if !found {
		return domain.FileNotUploadedError{}
	}
	refused := file.Confirm(object, now)
	if refused == nil {
		return nil
	}
	if err := u.storage.Remove(ctx, file.ObjectKey()); err != nil {
		// A failure of the system, not a rule: the object is still there.
		return fmt.Errorf("remove the object of a file that was refused (%v): %w", refused, err)
	}
	return refused
}
