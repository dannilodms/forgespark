// Package domain holds the file, the rules of an image and the errors of the
// files area. A file is a record of the database; its bytes are an object of
// the file storage, which the browser sends and fetches directly (ADR-0017).
package domain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"
)

// FileID identifies a file.
type FileID int64

// NewFileID reads an identifier that came from the sequence of files.
func NewFileID(value int64) (FileID, error) {
	if value <= 0 {
		return 0, errors.New("the identifier of a file is a positive number")
	}
	return FileID(value), nil
}

// Purpose is what a file is for. Each purpose takes an image.
type Purpose string

// The purposes of a file.
const (
	ConsultancyLogo Purpose = "consultancy_logo"
	UserPhoto       Purpose = "user_photo"
)

func (p Purpose) known() bool {
	return p == ConsultancyLogo || p == UserPhoto
}

// ObjectKey is where the bytes of a file are in the storage. It is random and
// holds neither the name of the file nor an identifier.
type ObjectKey string

// keyBytes is how many random bytes make a key: 128 bits.
const keyBytes = 16

// NewObjectKey makes a key from the random source.
func NewObjectKey(random io.Reader) (ObjectKey, error) {
	read := make([]byte, keyBytes)
	if _, err := io.ReadFull(random, read); err != nil {
		return "", fmt.Errorf("make the key of an object: %w", err)
	}
	return ObjectKey(hex.EncodeToString(read)), nil
}

// File is a file whose bytes are in the storage. Its upload starts with what
// the sender declares, and the file can be used once its upload is confirmed:
// once the object was checked.
type File struct {
	id          FileID
	objectKey   ObjectKey
	purpose     Purpose
	declared    DeclaredImage
	startedAt   time.Time
	confirmedAt *time.Time
}

// StartUpload creates the record of a file whose bytes are about to be sent.
func StartUpload(id FileID, key ObjectKey, purpose Purpose, declared DeclaredImage, at time.Time) (*File, error) {
	if !purpose.known() {
		return nil, fmt.Errorf("a file is for the logo of the consultancy or the photo of a user, not %q", purpose)
	}
	return &File{id: id, objectKey: key, purpose: purpose, declared: declared, startedAt: at}, nil
}

// RestoreFile rebuilds a file from what an adapter stored. It checks nothing
// that was checked when the file was saved.
func RestoreFile(
	id FileID, key ObjectKey, purpose Purpose, mediaType MediaType, size int64, startedAt time.Time, confirmedAt *time.Time,
) *File {
	return &File{
		id:          id,
		objectKey:   key,
		purpose:     purpose,
		declared:    DeclaredImage{mediaType: mediaType, size: size},
		startedAt:   startedAt,
		confirmedAt: confirmedAt,
	}
}

// Confirm checks the object the sender stored and makes the file usable,
// with its bytes kept under a key of their own, which no signed URL reaches:
// the caller copies the object there before the file is saved. An object
// over MaxImageSize, or one that is not a PNG, a JPEG or a WebP, is refused,
// and the file stays unconfirmed, under the key of its upload. A file
// confirmed again keeps its first confirmation: its time and its key.
func (f *File) Confirm(object StoredObject, kept ObjectKey, at time.Time) error {
	if err := checkStoredImage(object); err != nil {
		return err
	}
	if f.confirmedAt != nil {
		return nil
	}
	f.objectKey = kept
	f.confirmedAt = &at
	return nil
}

// IsConfirmed tells whether its upload was confirmed. The object of a
// confirmed file is under a key no signed URL reaches, so it cannot change.
func (f *File) IsConfirmed() bool {
	return f.confirmedAt != nil
}

// IsFor tells whether the file is for this purpose.
func (f *File) IsFor(purpose Purpose) bool {
	return f.purpose == purpose
}

// ID returns the identifier of the file.
func (f *File) ID() FileID { return f.id }

// ObjectKey returns where its bytes are in the storage: the key its upload
// was sent to until it is confirmed, and the key they are kept under after.
func (f *File) ObjectKey() ObjectKey { return f.objectKey }

// Purpose returns what the file is for.
func (f *File) Purpose() Purpose { return f.purpose }

// Declared returns the type and the size declared when its upload started.
func (f *File) Declared() DeclaredImage { return f.declared }

// StartedAt returns when its upload started.
func (f *File) StartedAt() time.Time { return f.startedAt }

// ConfirmedAt returns when its upload was confirmed, and false while it was
// not.
func (f *File) ConfirmedAt() (time.Time, bool) {
	if f.confirmedAt == nil {
		return time.Time{}, false
	}
	return *f.confirmedAt, true
}
