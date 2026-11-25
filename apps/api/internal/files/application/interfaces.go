// Package application holds the use cases of the files area and the
// interfaces they need. A use case is a type built once, with what it needs,
// and run with Execute. The area has no REST operation of its own yet: the
// use cases of the setup call these, after they checked the setup code, and
// the reads that show an image sign its download.
package application

import (
	"context"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/files/domain"
)

// Transactions opens and closes the transaction of a use case that writes.
type Transactions interface {
	// Run calls the operation inside one transaction, and passes it the
	// context every repository must receive to take part. The transaction is
	// committed when the operation returns nil and rolled back, with nothing
	// saved, when it returns an error.
	Run(ctx context.Context, operation func(ctx context.Context) error) error
}

// Clock tells the current time. A use case never reads the time of the
// machine by itself, so a test decides what time it is.
type Clock interface {
	Now() time.Time
}

// FileRepository keeps the files.
type FileRepository interface {
	// NextIdentifier reserves the identifier of a file that does not exist
	// yet.
	NextIdentifier(ctx context.Context) (domain.FileID, error)
	// Load loads the file with this identifier, and false when there is none.
	// It locks the file until the transaction ends.
	Load(ctx context.Context, id domain.FileID) (*domain.File, bool, error)
	// Save stores a file whole, whether it is new or changed.
	Save(ctx context.Context, file *domain.File) error
	// RemoveAbandoned removes the files whose upload started before this
	// instant and was never confirmed. A file being confirmed is removed, or
	// not, once its confirmation ends.
	RemoveAbandoned(ctx context.Context, startedBefore time.Time) error
	// ConfirmedKeys returns those of the keys a confirmed file names.
	ConfirmedKeys(ctx context.Context, among []domain.ObjectKey) ([]domain.ObjectKey, error)
	// ConfirmedKey returns the key the bytes of the file with this identifier
	// are kept under, and false when there is no such file or it was never
	// confirmed. It locks nothing: it serves a read.
	ConfirmedKey(ctx context.Context, id domain.FileID) (domain.ObjectKey, bool, error)
}

// Storage is the file storage, which holds the bytes of each file under its
// object key, in one private bucket.
type Storage interface {
	// SignUpload signs a URL to which the browser sends the bytes of an
	// object with a PUT, valid for this long. The URL accepts what was
	// declared only: its type and its size.
	SignUpload(ctx context.Context, key domain.ObjectKey, declared domain.DeclaredImage, validity time.Duration) (string, error)
	// SignDownload signs a URL from which the browser reads the bytes of the
	// object under this key with a GET, valid for this long.
	SignDownload(ctx context.Context, key domain.ObjectKey, validity time.Duration) (string, error)
	// Inspect reads the size of the object under this key and its first
	// bytes, at most domain.FirstBytesRead of them, and false when there is no
	// object.
	Inspect(ctx context.Context, key domain.ObjectKey) (domain.StoredObject, bool, error)
	// Copy copies the object under one key to another key, inside the
	// storage: the bytes never pass through the API. No URL signed for the
	// first key reaches the copy.
	Copy(ctx context.Context, from, to domain.ObjectKey) error
	// Remove removes the object under this key. Removing one that does not
	// exist is no failure.
	Remove(ctx context.Context, key domain.ObjectKey) error
	// List lists every object of the bucket, with the time it was last
	// written, and calls each with them a page at a time. It stops at the
	// first failure, of the storage or of each.
	List(ctx context.Context, each func(objects []domain.ListedObject) error) error
}
