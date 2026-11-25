// Package river runs the jobs of the files area: the periodic cleanup of the
// storage. A job worker reads the arguments of its job and calls one use case
// (Jobs in CODING_STANDARDS.md).
package river

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/dannilodms/forgespark/apps/api/internal/files/application"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/jobs"
)

// cleanUpTimeout is how long one cleanup may run. It lists the whole bucket
// and removes objects one by one, which takes longer than the minute River
// gives a job by default. Each call to the storage has its own timeout.
const cleanUpTimeout = 10 * time.Minute

// CleanUpStorageArgs are the arguments of the job that cleans up the storage.
// They name nothing: the job looks at every file and every object.
type CleanUpStorageArgs struct{}

// Kind names the job in River.
func (CleanUpStorageArgs) Kind() string { return "clean_up_storage" }

// InsertOpts puts the job in the light queue.
func (CleanUpStorageArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: jobs.QueueLight}
}

// CleanUpStorageWorker runs the cleanup of the storage. The `worker` command
// schedules it every hour.
type CleanUpStorageWorker struct {
	river.WorkerDefaults[CleanUpStorageArgs]
	cleanUpStorage *application.CleanUpStorage
}

// NewCleanUpStorageWorker builds the job worker on its use case.
func NewCleanUpStorageWorker(cleanUpStorage *application.CleanUpStorage) *CleanUpStorageWorker {
	return &CleanUpStorageWorker{WorkerDefaults: river.WorkerDefaults[CleanUpStorageArgs]{}, cleanUpStorage: cleanUpStorage}
}

// Work runs the cleanup once. Its error carries no key of an object, so the
// line River writes when a run fails carries none either.
func (w *CleanUpStorageWorker) Work(ctx context.Context, _ *river.Job[CleanUpStorageArgs]) error {
	return w.cleanUpStorage.Execute(ctx)
}

// Timeout gives the cleanup longer than River's default.
func (w *CleanUpStorageWorker) Timeout(*river.Job[CleanUpStorageArgs]) time.Duration {
	return cleanUpTimeout
}
