package application

import (
	"context"
	"io"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
)

// uploadValidity is how long a signed upload URL works: whoever holds it can
// send to it until then, so it lasts minutes.
const uploadValidity = 5 * time.Minute

// StartUpload creates the record of a file that a browser is about to send,
// under a random object key, and signs the URL the browser sends its bytes
// to. The file is not usable until its upload is confirmed. The caller checks
// whether the upload is allowed.
type StartUpload struct {
	transactions Transactions
	files        FileRepository
	storage      Storage
	clock        Clock
	random       io.Reader
}

// NewStartUpload builds the use case. The random source makes the object
// keys: the command passes crypto/rand.
func NewStartUpload(transactions Transactions, files FileRepository, storage Storage, clock Clock, random io.Reader) *StartUpload {
	return &StartUpload{transactions: transactions, files: files, storage: storage, clock: clock, random: random}
}

// StartUploadInput is what the file is for, and what the sender declares of
// it before sending a byte.
type StartUploadInput struct {
	Purpose   Purpose
	MediaType string
	Size      int64
}

// StartedUpload is the file created and where to send its bytes.
type StartedUpload struct {
	FileID FileID
	// URL is where the browser sends the bytes, with a PUT that carries the
	// declared type as its Content-Type.
	URL       string
	ExpiresAt time.Time
}

// Execute checks the declared type and size, answering with a
// kernel.InvalidFieldsError that names `media_type` and `size`, saves the
// file and signs its URL.
func (u *StartUpload) Execute(ctx context.Context, input StartUploadInput) (StartedUpload, error) {
	declared, err := domain.DeclareImage(input.MediaType, input.Size)
	if err != nil {
		return StartedUpload{}, err
	}
	now := u.clock.Now()
	key, err := domain.NewObjectKey(u.random)
	if err != nil {
		return StartedUpload{}, err
	}

	var file *domain.File
	err = u.transactions.Run(ctx, func(ctx context.Context) error {
		id, err := u.files.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		file, err = domain.StartUpload(id, key, input.Purpose, declared, now)
		if err != nil {
			return err
		}
		return u.files.Save(ctx, file)
	})
	if err != nil {
		return StartedUpload{}, err
	}

	url, err := u.storage.SignUpload(ctx, key, declared, uploadValidity)
	if err != nil {
		return StartedUpload{}, err
	}
	return StartedUpload{FileID: file.ID(), URL: url, ExpiresAt: now.Add(uploadValidity)}, nil
}
