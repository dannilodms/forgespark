// Package river runs the jobs of the access area: the daily cleanup of what
// the login keeps only for a while. A job worker reads the arguments of its
// job and calls one use case (Jobs in CODING_STANDARDS.md).
package river

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/dannilodms/forgespark/apps/api/internal/access/application"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/jobs"
)

// CleanUpOldRecordsArgs are the arguments of the job of the daily cleanup.
// They name nothing: the job removes every record old enough.
type CleanUpOldRecordsArgs struct{}

// Kind names the job in River.
func (CleanUpOldRecordsArgs) Kind() string { return "clean_up_old_records" }

// InsertOpts puts the job in the light queue.
func (CleanUpOldRecordsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: jobs.QueueLight}
}

// CleanUpOldRecordsWorker runs the daily cleanup. The `worker` command
// schedules it once a day.
type CleanUpOldRecordsWorker struct {
	river.WorkerDefaults[CleanUpOldRecordsArgs]

	cleanUpOldRecords *application.CleanUpOldRecords
}

// NewCleanUpOldRecordsWorker builds the job worker on its use case.
func NewCleanUpOldRecordsWorker(cleanUpOldRecords *application.CleanUpOldRecords) *CleanUpOldRecordsWorker {
	return &CleanUpOldRecordsWorker{WorkerDefaults: river.WorkerDefaults[CleanUpOldRecordsArgs]{}, cleanUpOldRecords: cleanUpOldRecords}
}

// Work runs the cleanup once.
func (w *CleanUpOldRecordsWorker) Work(ctx context.Context, _ *river.Job[CleanUpOldRecordsArgs]) error {
	return w.cleanUpOldRecords.Execute(ctx)
}
