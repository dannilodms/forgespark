package application

import (
	"context"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
	filesapplication "github.com/dannilodms/forgespark/apps/api/internal/files/application"
)

// StartSetupUpload starts the upload of an image chosen in the setup: the logo
// of the consultancy or the photo of its administrator. The setup code allows
// it, and the files area keeps the file. Concluding the setup confirms it.
type StartSetupUpload struct {
	setupCodes    SetupCodeRepository
	consultancies ConsultancyRepository
	startUpload   *filesapplication.StartUpload
	clock         Clock
}

// NewStartSetupUpload builds the use case.
func NewStartSetupUpload(
	setupCodes SetupCodeRepository,
	consultancies ConsultancyRepository,
	startUpload *filesapplication.StartUpload,
	clock Clock,
) *StartSetupUpload {
	return &StartSetupUpload{setupCodes: setupCodes, consultancies: consultancies, startUpload: startUpload, clock: clock}
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
// starts no upload, then starts it. The photo of an administrator is sent
// in the first setup and in the way back in; the logo of the consultancy
// only in the first setup, and a code that opens the registration of a new
// administrator answers SetupAlreadyConcludedError to it. The declared type
// and size are answered with a kernel.InvalidFieldsError that names
// `media_type` and `size`.
func (u *StartSetupUpload) Execute(ctx context.Context, input StartSetupUploadInput) (SetupUpload, error) {
	now := u.clock.Now()
	setupCode, err := findWorkingSetupCode(ctx, u.setupCodes, input.Code, now)
	if err != nil {
		return SetupUpload{}, err
	}
	if err := u.allows(ctx, setupCode, input.Purpose, now); err != nil {
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

// allows checks that what the setup code opens takes an image for this
// purpose: the logo belongs to the first setup alone.
func (u *StartSetupUpload) allows(ctx context.Context, setupCode *domain.SetupCode, purpose FilePurpose, now time.Time) error {
	if purpose != ConsultancyLogo {
		return nil
	}
	return opensWithWorkingSetupCode(ctx, u.setupCodes, u.consultancies, setupCode, domain.OpensFirstSetup, now)
}
