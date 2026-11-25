package application

import "github.com/dannilodms/forgespark/apps/api/internal/files/domain"

// Another domain area imports this package and no other of the files area, so
// the words it needs to start and to confirm an upload are named here.
type (
	// FileID identifies a file.
	FileID = domain.FileID
	// Purpose is what a file is for.
	Purpose = domain.Purpose
	// MediaType is the type of an image, as HTTP names it.
	MediaType = domain.MediaType
	// File is a file whose bytes are in the storage.
	File = domain.File

	// FileTypeUnsupportedError is the answer to a file that is not a PNG, a
	// JPEG or a WebP.
	FileTypeUnsupportedError = domain.FileTypeUnsupportedError
	// FileTooLargeError is the answer to an image larger than MaxImageSize.
	FileTooLargeError = domain.FileTooLargeError
	// FileEmptyError is the answer to an upload declared with no byte.
	FileEmptyError = domain.FileEmptyError
	// FileNotFoundError is the answer to a file that does not exist, or is
	// for something else.
	FileNotFoundError = domain.FileNotFoundError
	// FileNotUploadedError is the answer to a file whose bytes are not in the
	// storage.
	FileNotUploadedError = domain.FileNotUploadedError
)

// The purposes of a file.
const (
	ConsultancyLogo = domain.ConsultancyLogo
	UserPhoto       = domain.UserPhoto
)

// The types an image may be.
const (
	PNG  = domain.PNG
	JPEG = domain.JPEG
	WebP = domain.WebP
)

// MaxImageSize is the largest image, in bytes.
const MaxImageSize = domain.MaxImageSize

// ImageTypes lists the types an image may be.
func ImageTypes() []MediaType {
	return domain.ImageTypes()
}
