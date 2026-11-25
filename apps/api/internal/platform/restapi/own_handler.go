package restapi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
)

// A handler of its own is one Huma does not serve: it is registered on the
// mux under the prefix, outside the spec and outside every middleware of the
// REST adapter, such as the bounce notices of Resend, whose body is read raw
// to verify its signature. These functions give it the limits of a body and
// the error format every operation has.

// ReadBody reads the body of the request within the limits: the size, and
// the time the caller has to send it. A body over either is answered in the
// error format, 413 or 408, and ReadBody returns false: the handler answers
// nothing more.
func ReadBody(w http.ResponseWriter, r *http.Request, limits Limits) ([]byte, bool) {
	// A writer that cannot set a deadline, such as the recorder of a test,
	// reads with no time limit.
	_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(limits.BodyReadTimeout))
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limits.BodyBytes))
	if err == nil {
		return body, true
	}
	var tooLarge *http.MaxBytesError
	status := http.StatusBadRequest
	if errors.As(err, &tooLarge) {
		status = http.StatusRequestEntityTooLarge
	}
	if errors.Is(err, os.ErrDeadlineExceeded) {
		status = http.StatusRequestTimeout
	}
	WriteRefusal(w, r, status, "")
	return nil, false
}

// WriteRefusal answers a request the handler refuses, in the error format:
// the status, the code named after it (`unauthorized`, `bad_request`), the
// detail in English and the trace identifier. A refusal writes no line.
func WriteRefusal(w http.ResponseWriter, r *http.Request, status int, detail string) {
	writeProblem(w, newProblem(r.Context(), status, codeOfStatus(status), detail))
}

// WriteFailure answers 500 in the error format. The raw error goes to the
// line of the request, and nothing of it to the answer.
func WriteFailure(w http.ResponseWriter, r *http.Request, err error) {
	telemetry.RecordFailure(r.Context(), err)
	writeProblem(w, newProblem(r.Context(), http.StatusInternalServerError, codeOfStatus(http.StatusInternalServerError), ""))
}
