package domain

import (
	"bytes"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// MaxImageSize is the largest image, in bytes: 2 MiB. Nothing is resized or
// cropped, so a larger one is refused.
const MaxImageSize int64 = 2 * 1024 * 1024

// MediaType is the type of an image, as HTTP names it.
type MediaType string

// The three types an image may be.
const (
	PNG  MediaType = "image/png"
	JPEG MediaType = "image/jpeg"
	WebP MediaType = "image/webp"
)

// ImageTypes lists the types an image may be.
func ImageTypes() []MediaType {
	return []MediaType{PNG, JPEG, WebP}
}

// DeclaredImage is what the sender of an image says it is, before a byte of
// it was sent: its type and its size. Only DeclareImage builds one.
type DeclaredImage struct {
	mediaType MediaType
	size      int64
}

// DeclareImage checks what the sender declares: a PNG, a JPEG or a WebP, of
// one byte to 2 MiB. It returns a kernel.InvalidFieldsError that names each
// value refused, as `media_type` and `size`.
func DeclareImage(mediaType string, size int64) (DeclaredImage, error) {
	var invalid kernel.InvalidFieldsError
	declared := DeclaredImage{mediaType: MediaType(mediaType), size: size}
	invalid.AddField("media_type", checkImageType(declared.mediaType))
	invalid.AddField("size", checkImageSize(size))
	if err := invalid.Err(); err != nil {
		return DeclaredImage{}, err
	}
	return declared, nil
}

// MediaType returns the type declared.
func (d DeclaredImage) MediaType() MediaType { return d.mediaType }

// Size returns the size declared, in bytes.
func (d DeclaredImage) Size() int64 { return d.size }

func checkImageType(mediaType MediaType) error {
	if !slices.Contains(ImageTypes(), mediaType) {
		return FileTypeUnsupportedError{}
	}
	return nil
}

func checkImageSize(size int64) error {
	if size < 1 {
		return FileEmptyError{}
	}
	if size > MaxImageSize {
		return FileTooLargeError{}
	}
	return nil
}

// StoredObject is what the storage holds for a file: the size of its object
// and its first bytes, at most FirstBytesRead of them.
type StoredObject struct {
	Size       int64
	FirstBytes []byte
}

// FirstBytesRead is how many bytes of an object tell its type: the signature
// of a PNG, a JPEG or a WebP fits in them.
const FirstBytesRead = 12

// signatures tell, from its first bytes, whether an object is a PNG, a JPEG
// or a WebP. A WebP is a RIFF file whose form is WEBP: the four bytes between
// are its size.
var signatures = []func(first []byte) bool{
	func(first []byte) bool { return bytes.HasPrefix(first, []byte("\x89PNG\r\n\x1a\n")) },
	func(first []byte) bool { return bytes.HasPrefix(first, []byte{0xFF, 0xD8, 0xFF}) },
	func(first []byte) bool {
		return len(first) >= 12 && bytes.Equal(first[:4], []byte("RIFF")) && bytes.Equal(first[8:12], []byte("WEBP"))
	},
}

// checkStoredImage checks the object the sender stored: its real size, which
// no signed URL can bound, and that its first bytes are those of a PNG, a
// JPEG or a WebP.
func checkStoredImage(object StoredObject) error {
	if object.Size > MaxImageSize {
		return FileTooLargeError{}
	}
	isImage := slices.ContainsFunc(signatures, func(matches func([]byte) bool) bool { return matches(object.FirstBytes) })
	if !isImage {
		return FileTypeUnsupportedError{}
	}
	return nil
}
