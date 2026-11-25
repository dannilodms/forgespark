package applicationtest

import (
	"context"
	"slices"

	"github.com/dannilodms/forgespark/apps/api/internal/email/domain"
)

// DeliveryJobs keeps the jobs that deliver e-mails in memory, as the
// identifiers of their e-mails. It takes part in the in-memory transactions,
// as River saves a job in the transaction of the use case.
type DeliveryJobs struct {
	saved   []domain.EmailID
	failure error
}

// NewDeliveryJobs returns no job.
func NewDeliveryJobs() *DeliveryJobs {
	return &DeliveryJobs{}
}

// FailWith makes every later Save fail with this error.
func (j *DeliveryJobs) FailWith(failure error) {
	j.failure = failure
}

// Save keeps the job of the e-mail.
func (j *DeliveryJobs) Save(_ context.Context, id domain.EmailID) error {
	if j.failure != nil {
		return j.failure
	}
	j.saved = append(j.saved, id)
	return nil
}

// Snapshot lets the jobs take part in the in-memory transactions.
func (j *DeliveryJobs) Snapshot() (restore func()) {
	saved := slices.Clone(j.saved)
	return func() { j.saved = saved }
}

// Saved returns the e-mail of each job saved, in order.
func (j *DeliveryJobs) Saved() []domain.EmailID {
	return slices.Clone(j.saved)
}
