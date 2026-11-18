package applicationtest

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
)

// Storage keeps objects in memory, as the file storage does under their keys,
// and remembers each URL it signed.
type Storage struct {
	objects map[domain.ObjectKey][]byte
	signed  map[domain.ObjectKey]SignedUpload
	failure error
	// Failures of one kind of call only.
	copyFailure   error
	removeFailure error
}

// SignedUpload is a URL the storage signed, and what it accepts.
type SignedUpload struct {
	URL      string
	Declared domain.DeclaredImage
	Validity time.Duration
}

// NewStorage returns a storage with no object.
func NewStorage() *Storage {
	return &Storage{objects: map[domain.ObjectKey][]byte{}, signed: map[domain.ObjectKey]SignedUpload{}}
}

// FailWith makes every later call fail with this error, as a storage that
// does not answer.
func (s *Storage) FailWith(failure error) {
	s.failure = failure
}

// FailCopyingWith makes every later copy fail with this error, and nothing
// else.
func (s *Storage) FailCopyingWith(failure error) {
	s.copyFailure = failure
}

// FailRemovingWith makes every later removal fail with this error, and
// nothing else.
func (s *Storage) FailRemovingWith(failure error) {
	s.removeFailure = failure
}

// SignUpload signs a URL for the object under this key.
func (s *Storage) SignUpload(_ context.Context, key domain.ObjectKey, declared domain.DeclaredImage, validity time.Duration) (string, error) {
	if s.failure != nil {
		return "", s.failure
	}
	url := "https://files.example/bucket/" + string(key) + "?signature=in-memory"
	s.signed[key] = SignedUpload{URL: url, Declared: declared, Validity: validity}
	return url, nil
}

// Put stores the bytes under the key, as the browser does with the signed
// URL: whatever it chooses to send.
func (s *Storage) Put(key domain.ObjectKey, content []byte) {
	s.objects[key] = slices.Clone(content)
}

// Inspect reads the size of the object and its first bytes.
func (s *Storage) Inspect(_ context.Context, key domain.ObjectKey) (domain.StoredObject, bool, error) {
	if s.failure != nil {
		return domain.StoredObject{}, false, s.failure
	}
	content, found := s.objects[key]
	if !found {
		return domain.StoredObject{}, false, nil
	}
	first := content[:min(len(content), domain.FirstBytesRead)]
	return domain.StoredObject{Size: int64(len(content)), FirstBytes: slices.Clone(first)}, true, nil
}

// Copy stores a copy of the object under one key under the other. The URL
// signed for the first key reaches the first key only.
func (s *Storage) Copy(_ context.Context, from, to domain.ObjectKey) error {
	if err := cmp.Or(s.failure, s.copyFailure); err != nil {
		return err
	}
	content, found := s.objects[from]
	if !found {
		return fmt.Errorf("copy %s: no such object", from)
	}
	s.objects[to] = slices.Clone(content)
	return nil
}

// Remove removes the object under the key.
func (s *Storage) Remove(_ context.Context, key domain.ObjectKey) error {
	if err := cmp.Or(s.failure, s.removeFailure); err != nil {
		return err
	}
	delete(s.objects, key)
	return nil
}

// Holds tells whether an object is stored under the key.
func (s *Storage) Holds(key domain.ObjectKey) bool {
	_, found := s.objects[key]
	return found
}

// Content returns the bytes stored under the key, and false when there are
// none.
func (s *Storage) Content(key domain.ObjectKey) ([]byte, bool) {
	content, found := s.objects[key]
	return slices.Clone(content), found
}

// SignedUpload returns the URL signed for the object under the key.
func (s *Storage) SignedUpload(key domain.ObjectKey) (SignedUpload, bool) {
	signed, found := s.signed[key]
	return signed, found
}

// Keys lists the keys an object is stored under.
func (s *Storage) Keys() []domain.ObjectKey {
	return slices.Collect(maps.Keys(s.objects))
}

// SignedKeys lists the keys a URL was signed for.
func (s *Storage) SignedKeys() []domain.ObjectKey {
	return slices.Collect(maps.Keys(s.signed))
}
