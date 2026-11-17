package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
)

// StartSetupUpload starts the upload of an image chosen in the setup: the logo
// of the consultancy or the photo of its administrator. The setup code allows
// it, and the files area keeps the file. Concluding the setup confirms it.
type StartSetupUpload struct {
	setupCodes  SetupCodeRepository
	startUpload *filesapplication.StartUpload
	clock       Clock
}

// NewStartSetupUpload builds the use case.
func NewStartSetupUpload(setupCodes SetupCodeRepository, startUpload *filesapplication.StartUpload, clock Clock) *StartSetupUpload {
	return &StartSetupUpload{setupCodes: setupCodes, startUpload: startUpload, clock: clock}
}

// StartSetupUploadInput carries the setup code as it was typed, what the
// image is for, and what the sender declares of it.
type StartSetupUploadInput struct {
	Code      string
	Purpose   FilePurpose
	MediaType string
	Size      int64
}

// SetupUpload is the file of an image and where the browser sends it.
type SetupUpload struct {
	FileID domain.FileID
	// URL is where the browser sends the bytes, with a PUT that carries the
	// declared type as its Content-Type and the declared size.
	URL       string
	ExpiresAt time.Time
}

// Execute checks the setup code first, so that a code that opens nothing
// starts no upload, then starts it. The declared type and size are answered
// with a kernel.InvalidFieldsError that names `media_type` and `size`.
func (u *StartSetupUpload) Execute(ctx context.Context, input StartSetupUploadInput) (SetupUpload, error) {
	if _, err := findWorkingSetupCode(ctx, u.setupCodes, input.Code, u.clock.Now()); err != nil {
		return SetupUpload{}, err
	}

	started, err := u.startUpload.Execute(ctx, filesapplication.StartUploadInput{
		Purpose:   input.Purpose,
		MediaType: input.MediaType,
		Size:      input.Size,
	})
	if err != nil {
		return SetupUpload{}, err
	}
	return SetupUpload{FileID: domain.FileID(started.FileID), URL: started.URL, ExpiresAt: started.ExpiresAt}, nil
}
