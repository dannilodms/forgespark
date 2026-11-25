package applicationtest

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/files/domain"
)

// listPageSize is how many objects a page of the listing holds: few, so a
// use case that lists meets more than one page.
const listPageSize = 2

// Storage keeps objects in memory, as the file storage does under their keys,
// each with the time it was written by the clock it was given, and remembers
// each URL it signed, to send an object or to read one.
type Storage struct {
	clock     interface{ Now() time.Time }
	objects   map[domain.ObjectKey]storedObject
	signed    map[domain.ObjectKey]SignedUpload
	downloads map[domain.ObjectKey]SignedDownload
	failure   error
	// Failures of one kind of call only.
	copyFailure   error
	removeFailure error
	listFailure   error
}

// storedObject is an object and when it was written.
type storedObject struct {
	content []byte
	written time.Time
}

// SignedUpload is a URL the storage signed, and what it accepts.
type SignedUpload struct {
	URL      string
	Declared domain.DeclaredImage
	Validity time.Duration
}

// SignedDownload is a URL the storage signed to read an object.
type SignedDownload struct {
	URL      string
	Validity time.Duration
}

// NewStorage returns a storage with no object, which writes each object at
// the time of this clock.
func NewStorage(clock interface{ Now() time.Time }) *Storage {
	return &Storage{
		clock:     clock,
		objects:   map[domain.ObjectKey]storedObject{},
		signed:    map[domain.ObjectKey]SignedUpload{},
		downloads: map[domain.ObjectKey]SignedDownload{},
	}
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

// FailListingWith makes every later listing fail with this error, and
// nothing else.
func (s *Storage) FailListingWith(failure error) {
	s.listFailure = failure
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

// SignDownload signs a URL that reads the object under this key.
func (s *Storage) SignDownload(_ context.Context, key domain.ObjectKey, validity time.Duration) (string, error) {
	if s.failure != nil {
		return "", s.failure
	}
	url := "https://files.example/bucket/" + string(key) + "?signature=in-memory-download"
	s.downloads[key] = SignedDownload{URL: url, Validity: validity}
	return url, nil
}

// SignedDownload returns the last URL signed to read the object under the
// key.
func (s *Storage) SignedDownload(key domain.ObjectKey) (SignedDownload, bool) {
	signed, found := s.downloads[key]
	return signed, found
}

// SignedDownloadKeys lists the keys a URL to read was signed for.
func (s *Storage) SignedDownloadKeys() []domain.ObjectKey {
	return slices.Collect(maps.Keys(s.downloads))
}

// Put stores the bytes under the key, now, as the browser does with the
// signed URL: whatever it chooses to send.
func (s *Storage) Put(key domain.ObjectKey, content []byte) {
	s.objects[key] = storedObject{content: slices.Clone(content), written: s.clock.Now()}
}

// Inspect reads the size of the object and its first bytes.
func (s *Storage) Inspect(_ context.Context, key domain.ObjectKey) (domain.StoredObject, bool, error) {
	if s.failure != nil {
		return domain.StoredObject{}, false, s.failure
	}
	object, found := s.objects[key]
	if !found {
		return domain.StoredObject{}, false, nil
	}
	first := object.content[:min(len(object.content), domain.FirstBytesRead)]
	return domain.StoredObject{Size: int64(len(object.content)), FirstBytes: slices.Clone(first)}, true, nil
}

// Copy stores a copy of the object under one key under the other, written
// now. The URL signed for the first key reaches the first key only.
func (s *Storage) Copy(_ context.Context, from, to domain.ObjectKey) error {
	if err := cmp.Or(s.failure, s.copyFailure); err != nil {
		return err
	}
	object, found := s.objects[from]
	if !found {
		return fmt.Errorf("copy %s: no such object", from)
	}
	s.objects[to] = storedObject{content: slices.Clone(object.content), written: s.clock.Now()}
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

// List calls each with the objects stored when the listing started, in the
// order of their keys, listPageSize at a time.
func (s *Storage) List(_ context.Context, each func(objects []domain.ListedObject) error) error {
	if err := cmp.Or(s.failure, s.listFailure); err != nil {
		return err
	}
	listed := make([]domain.ListedObject, 0, len(s.objects))
	for _, key := range slices.Sorted(maps.Keys(s.objects)) {
		listed = append(listed, domain.ListedObject{Key: key, LastModified: s.objects[key].written})
	}
	pages := slices.Collect(slices.Chunk(listed, listPageSize))
	var err error
	for i := 0; err == nil && i < len(pages); i++ {
		err = each(pages[i])
	}
	return err
}

// Holds tells whether an object is stored under the key.
func (s *Storage) Holds(key domain.ObjectKey) bool {
	_, found := s.objects[key]
	return found
}

// Content returns the bytes stored under the key, and false when there are
// none.
func (s *Storage) Content(key domain.ObjectKey) ([]byte, bool) {
	object, found := s.objects[key]
	return slices.Clone(object.content), found
}

// SignedUpload returns the URL signed for the object under the key.
func (s *Storage) SignedUpload(key domain.ObjectKey) (SignedUpload, bool) {
	signed, found := s.signed[key]
	return signed, found
}

// Keys lists the keys an object is stored under, in their order.
func (s *Storage) Keys() []domain.ObjectKey {
	return slices.Sorted(maps.Keys(s.objects))
}

// SignedKeys lists the keys a URL was signed for.
func (s *Storage) SignedKeys() []domain.ObjectKey {
	return slices.Collect(maps.Keys(s.signed))
}
