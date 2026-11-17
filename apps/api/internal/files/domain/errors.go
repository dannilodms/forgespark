package domain

import "strconv"

// FileTypeUnsupportedError is the answer to a file that is not an image this
// system takes: one declared as another type when its upload starts, or one
// whose first bytes are not those of a PNG, a JPEG or a WebP when it is
// confirmed.
type FileTypeUnsupportedError struct{}

func (FileTypeUnsupportedError) Error() string {
	return "the file is not a PNG, a JPEG or a WebP image"
}

// Code returns the stable code an inbound adapter answers with.
func (FileTypeUnsupportedError) Code() string {
	return "file_type_unsupported"
}

// FileTooLargeError is the answer to an image larger than MaxImageSize: one
// declared so when its upload starts, or one stored so when it is confirmed.
type FileTooLargeError struct{}

func (FileTooLargeError) Error() string {
	return "the file is larger than " + strconv.FormatInt(MaxImageSize, 10) + " bytes"
}

// Code returns the stable code an inbound adapter answers with.
func (FileTooLargeError) Code() string {
	return "file_too_large"
}

// Params returns the largest size, in bytes.
func (FileTooLargeError) Params() map[string]any {
	return map[string]any{"max_size": MaxImageSize}
}

// FileEmptyError is the answer to an upload declared with no byte.
type FileEmptyError struct{}

func (FileEmptyError) Error() string {
	return "the file has no byte"
}

// Code returns the stable code an inbound adapter answers with.
func (FileEmptyError) Code() string {
	return "file_empty"
}

// FileNotFoundError is the answer to the identifier of a file that does not
// exist, or that is for something else than what it was given for.
type FileNotFoundError struct{}

func (FileNotFoundError) Error() string {
	return "no file of this kind has this identifier"
}

// Code returns the stable code an inbound adapter answers with.
func (FileNotFoundError) Code() string {
	return "file_not_found"
}

// FileNotUploadedError is the answer to confirming a file whose bytes never
// reached the storage, or were removed because they broke a rule.
type FileNotUploadedError struct{}

func (FileNotUploadedError) Error() string {
	return "the file was never sent to the storage"
}

// Code returns the stable code an inbound adapter answers with.
func (FileNotUploadedError) Code() string {
	return "file_not_uploaded"
}
