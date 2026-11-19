package application

import (
	"context"
	"time"
)

// downloadValidity is how long a signed download URL works: whoever holds it
// can read the file until then, so it lasts minutes. A read that answers one
// is read again sooner than that, and gets a new one.
const downloadValidity = 15 * time.Minute

// SignDownload signs the URL from which the browser reads the bytes of a
// confirmed file, straight from the storage: an image shows from there, and
// never passes through the API (ADR-0017). The read of another area that
// shows the file calls it, after it checked whether the file may be read.
type SignDownload struct {
	files   FileRepository
	storage Storage
}

// NewSignDownload builds the use case.
func NewSignDownload(files FileRepository, storage Storage) *SignDownload {
	return &SignDownload{files: files, storage: storage}
}

// Execute returns a URL that reads the file, valid for fifteen minutes, and
// false when there is no such file or it was never confirmed: the object of
// an upload not confirmed holds whatever its sender chose to send. It locks
// nothing and saves nothing.
func (u *SignDownload) Execute(ctx context.Context, id FileID) (string, bool, error) {
	key, found, err := u.files.ConfirmedKey(ctx, id)
	if err != nil || !found {
		return "", false, err
	}
	url, err := u.storage.SignDownload(ctx, key, downloadValidity)
	if err != nil {
		return "", false, err
	}
	return url, true, nil
}

// URLOf signs the download of a file that may be absent, as the photo of a
// user or the logo of the consultancy: the URL of Execute, and no URL, with
// nothing signed, when there is no file, nil, or it was never confirmed.
func (u *SignDownload) URLOf(ctx context.Context, id *FileID) (string, error) {
	if id == nil {
		return "", nil
	}
	url, _, err := u.Execute(ctx, *id)
	return url, err
}
