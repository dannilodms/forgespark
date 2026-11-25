package application

import (
	filesapplication "github.com/dannilodms/forgespark/apps/api/internal/files/application"
)

// The setup uploads images through the files area. An adapter of this area
// reaches another area only through this package, so the words of the files
// area its operations answer with are named here.
type (
	// FilePurpose is what an image of the setup is for.
	FilePurpose = filesapplication.Purpose
	// FileTypeUnsupportedError is the answer to an image that is not a PNG,
	// a JPEG or a WebP.
	FileTypeUnsupportedError = filesapplication.FileTypeUnsupportedError
	// FileTooLargeError is the answer to an image larger than MaxImageSize.
	FileTooLargeError = filesapplication.FileTooLargeError
	// FileEmptyError is the answer to an upload declared with no byte.
	FileEmptyError = filesapplication.FileEmptyError
	// FileNotFoundError is the answer to an image that does not exist, or is
	// for something else.
	FileNotFoundError = filesapplication.FileNotFoundError
	// FileNotUploadedError is the answer to an image whose bytes are not in
	// the storage.
	FileNotUploadedError = filesapplication.FileNotUploadedError
)

// The images of the setup: the logo of the consultancy and the photo of its
// administrator, who is a user.
const (
	ConsultancyLogo = filesapplication.ConsultancyLogo
	UserPhoto       = filesapplication.UserPhoto
)

// MaxImageSize is the largest image, in bytes.
const MaxImageSize = filesapplication.MaxImageSize

// ImageTypes lists the types an image may be, as HTTP names them.
func ImageTypes() []string {
	types := filesapplication.ImageTypes()
	names := make([]string, 0, len(types))
	for _, each := range types {
		names = append(names, string(each))
	}
	return names
}
