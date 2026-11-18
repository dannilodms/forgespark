package application

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
)

// ConfirmUpload checks the object a browser sent to the storage and makes its
// file usable, with the bytes kept under a new random key, which no signed URL
// reaches. An object over the size limit, or one that is not an image of the
// three types, is removed from the storage, and its file stays unconfirmed.
// The caller checks whether the confirmation is allowed.
type ConfirmUpload struct {
	transactions Transactions
	files        FileRepository
	storage      Storage
	clock        Clock
	random       io.Reader
}

// NewConfirmUpload builds the use case. The random source makes the key the
// bytes of a confirmed file are kept under: the command passes crypto/rand.
func NewConfirmUpload(transactions Transactions, files FileRepository, storage Storage, clock Clock, random io.Reader) *ConfirmUpload {
	return &ConfirmUpload{transactions: transactions, files: files, storage: storage, clock: clock, random: random}
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
// confirmed stays as it is, and the storage is not reached: no signed URL
// reaches its object, which cannot have changed.
//
// The file is checked in a transaction, which locks it, so two confirmations
// of one file run one after the other and the second finds it confirmed. The
// object of a file confirmed is copied to its new key in that transaction,
// before the file names that key. Removing the object the upload was sent to
// is an effect on the storage, so it runs once that transaction is committed
// (Writes in CODING_STANDARDS.md).
func (u *ConfirmUpload) Execute(ctx context.Context, input ConfirmUploadInput) error {
	id, err := domain.NewFileID(int64(input.FileID))
	if err != nil {
		return domain.FileNotFoundError{}
	}
	now := u.clock.Now()
	kept, err := domain.NewObjectKey(u.random)
	if err != nil {
		return err
	}

	// Whether the file was confirmed before; if not, the object its upload
	// was sent to, which nothing reads once the file is confirmed or refused,
	// and the rule the object broke, if any.
	var confirmedBefore bool
	var uploaded domain.ObjectKey
	var broken error
	err = u.transactions.Run(ctx, func(ctx context.Context) error {
		file, found, err := u.files.Load(ctx, id)
		if err != nil {
			return err
		}
		if !found || !file.IsFor(input.Purpose) {
			return domain.FileNotFoundError{}
		}
		confirmedBefore = file.IsConfirmed()
		if confirmedBefore {
			return nil
		}
		uploaded = file.ObjectKey()
		broken, err = u.confirm(ctx, file, kept, now)
		if err != nil || broken != nil {
			// Nothing of a refused file changed: it stays unconfirmed.
			return err
		}
		return u.files.Save(ctx, file)
	})
	if err != nil || confirmedBefore {
		return err
	}
	if broken == nil {
		// The file is confirmed, under a key no signed URL reaches. When the
		// object of its upload cannot be removed, the confirmation stands and
		// that object is left behind: no file names it, so nothing reads it,
		// and its URL stops working minutes after the upload started
		// (ADR-0017). The failed call shows in the span of the storage.
		_ = u.storage.Remove(ctx, uploaded)
		return nil
	}
	return u.remove(ctx, uploaded, broken)
}

// confirm reads the object of the file, tells the file what it holds and,
// when the file takes it, copies the object to the key the file keeps it
// under. It returns the rule the object breaks, when it breaks one, apart
// from a failure of the system.
func (u *ConfirmUpload) confirm(ctx context.Context, file *domain.File, kept domain.ObjectKey, now time.Time) (broken, failed error) {
	uploaded := file.ObjectKey()
	object, found, err := u.storage.Inspect(ctx, uploaded)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.FileNotUploadedError{}
	}
	if broken := file.Confirm(object, kept, now); broken != nil {
		return broken, nil
	}
	return nil, u.storage.Copy(ctx, uploaded, kept)
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
