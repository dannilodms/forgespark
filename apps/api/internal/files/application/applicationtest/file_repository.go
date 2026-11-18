package applicationtest

import (
	"context"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
)

// FileRepository keeps the files in memory. It stores a copy of what it is
// given and hands out copies, as the database does.
type FileRepository struct {
	lastIdentifier int64
	files          []domain.File
	failure        error
}

// NewFileRepository returns a repository with no file.
func NewFileRepository() *FileRepository {
	return &FileRepository{}
}

// FailWith makes every later Save fail with this error, as a database that
// refuses the write.
func (r *FileRepository) FailWith(failure error) {
	r.failure = failure
}

// NextIdentifier reserves the next identifier, starting at 1.
func (r *FileRepository) NextIdentifier(context.Context) (domain.FileID, error) {
	r.lastIdentifier++
	return domain.NewFileID(r.lastIdentifier)
}

// Load returns a copy of the file with this identifier.
func (r *FileRepository) Load(_ context.Context, id domain.FileID) (*domain.File, bool, error) {
	position := slices.IndexFunc(r.files, func(stored domain.File) bool { return stored.ID() == id })
	if position < 0 {
		return nil, false, nil
	}
	file := r.files[position]
	return &file, true, nil
}

// Save stores the file, in place of the one with the same identifier when
// there is one.
func (r *FileRepository) Save(_ context.Context, file *domain.File) error {
	if r.failure != nil {
		return r.failure
	}
	position := slices.IndexFunc(r.files, func(stored domain.File) bool { return stored.ID() == file.ID() })
	if position < 0 {
		r.files = append(r.files, *file)
		return nil
	}
	r.files[position] = *file
	return nil
}

// Snapshot lets the repository take part in the in-memory transactions.
func (r *FileRepository) Snapshot() (restore func()) {
	saved := slices.Clone(r.files)
	return func() { r.files = saved }
}

// All returns every file stored, in the order they were first saved.
func (r *FileRepository) All() []domain.File {
	return slices.Clone(r.files)
}

// Find returns the file with this identifier, and false when there is none.
func (r *FileRepository) Find(id domain.FileID) (domain.File, bool) {
	position := slices.IndexFunc(r.files, func(stored domain.File) bool { return stored.ID() == id })
	if position < 0 {
		return domain.File{}, false
	}
	return r.files[position], true
}
