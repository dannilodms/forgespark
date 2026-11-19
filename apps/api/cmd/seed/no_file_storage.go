package main

import (
	"context"
	"errors"
	"time"

	filesdomain "github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
)

// errNoFileStorage is what every call to the storage of the seed answers. The
// seed chooses no image, so the use case that confirms an image never calls
// it: the error names the mistake if that ever changes.
var errNoFileStorage = errors.New("the seed has no file storage: it chooses no image")

// noFileStorage stands where the files area expects the file storage. The
// seed requires nothing of the storage in its configuration, and it reaches
// none.
type noFileStorage struct{}

func (noFileStorage) SignUpload(context.Context, filesdomain.ObjectKey, filesdomain.DeclaredImage, time.Duration) (string, error) {
	return "", errNoFileStorage
}

func (noFileStorage) Inspect(context.Context, filesdomain.ObjectKey) (filesdomain.StoredObject, bool, error) {
	return filesdomain.StoredObject{}, false, errNoFileStorage
}

func (noFileStorage) Copy(context.Context, filesdomain.ObjectKey, filesdomain.ObjectKey) error {
	return errNoFileStorage
}

func (noFileStorage) Remove(context.Context, filesdomain.ObjectKey) error {
	return errNoFileStorage
}

func (noFileStorage) List(context.Context, func([]filesdomain.ListedObject) error) error {
	return errNoFileStorage
}
