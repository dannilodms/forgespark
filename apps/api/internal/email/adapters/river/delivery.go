// Package river delivers e-mails as River jobs. Outbound, it saves the job of
// an e-mail in the transaction that records it; inbound, its job worker runs
// the delivery. A job carries the identifier of its e-mail and nothing else,
// because the River panel and the telemetry show the arguments of a job.
package river

import (
	"context"

	riverqueue "github.com/riverqueue/river"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/jobs"
)

// DeliveryArgs are the arguments of the job that delivers an e-mail.
type DeliveryArgs struct {
	EmailID int64 `json:"email_id"`
}

// Kind names the job kind.
func (DeliveryArgs) Kind() string { return "deliver_email" }

// InsertOpts puts the job in the light queue, with the attempts and the
// timeout River's client gives every job.
func (DeliveryArgs) InsertOpts() riverqueue.InsertOpts {
	return riverqueue.InsertOpts{Queue: jobs.QueueLight}
}

// DeliveryJobs saves the job of an e-mail, for QueueEmail.
type DeliveryJobs struct {
	saver *jobs.Saver
}

// NewDeliveryJobs returns the jobs saved by this saver.
func NewDeliveryJobs(saver *jobs.Saver) *DeliveryJobs {
	return &DeliveryJobs{saver: saver}
}

// Save saves the job of the e-mail in the transaction the context carries.
func (j *DeliveryJobs) Save(ctx context.Context, id domain.EmailID) error {
	return j.saver.Save(ctx, DeliveryArgs{EmailID: int64(id)}, nil)
}

// DeliveryWorker runs the job of an e-mail: it calls DeliverEmail, telling it
// whether this attempt is the last one.
type DeliveryWorker struct {
	riverqueue.WorkerDefaults[DeliveryArgs]

	deliver *application.DeliverEmail
}

// NewDeliveryWorker returns the job worker of the delivery.
func NewDeliveryWorker(deliver *application.DeliverEmail) *DeliveryWorker {
	return &DeliveryWorker{WorkerDefaults: riverqueue.WorkerDefaults[DeliveryArgs]{}, deliver: deliver}
}

// Work delivers the e-mail of the job.
func (w *DeliveryWorker) Work(ctx context.Context, job *riverqueue.Job[DeliveryArgs]) error {
	return w.deliver.Execute(ctx, application.DeliverEmailInput{
		EmailID:     domain.EmailID(job.Args.EmailID),
		LastAttempt: job.Attempt >= job.MaxAttempts,
	})
}
