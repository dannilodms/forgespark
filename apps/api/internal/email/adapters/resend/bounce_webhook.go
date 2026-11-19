package resend

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// BounceWebhookPath is where Resend delivers the notices of a bounce: the
// address of the webhook to register in the account of Resend, under the
// production address, with the event `email.bounced`.
const BounceWebhookPath = restapi.Prefix + "/resend/bounces"

// signatureTolerance is how far from now a notice may say it was sent. A
// notice recorded by someone and sent again later is refused. Svix's own
// libraries allow five minutes.
const signatureTolerance = 5 * time.Minute

// errSignature is the refusal of a notice that does not prove it comes from
// Resend: a signature that is missing, wrong or of another time.
var errSignature = errors.New("the signature of the notice is not valid")

// BounceWebhook takes the notices of a bounce that Resend sends, signed with
// Svix: it verifies the signature on the raw body and the time it was sent,
// and runs MarkEmailBounced for each `email.bounced` event that names an
// e-mail record. It is a handler of its own, outside Huma: the OpenAPI spec
// leaves it out and no middleware of the REST adapter runs on it, the check
// of the origin included. It reads no cookie.
//
// A notice is answered 204 when it is taken, and when it is about something
// else, so Resend stops delivering it. A wrong signature is answered 401, and
// a failure 500, after which Resend delivers the notice again.
type BounceWebhook struct {
	secret      config.WebhookSecret
	limits      restapi.Limits
	markBounced *application.MarkEmailBounced
	clock       application.Clock
}

// NewBounceWebhook returns the webhook that verifies notices with the secret
// of the webhook in the account of Resend.
func NewBounceWebhook(secret config.WebhookSecret, limits restapi.Limits, markBounced *application.MarkEmailBounced, clock application.Clock) *BounceWebhook {
	return &BounceWebhook{secret: secret, limits: limits, markBounced: markBounced, clock: clock}
}

// Mount registers the webhook on the mux, at BounceWebhookPath, for POST.
func (b *BounceWebhook) Mount(mux *http.ServeMux) {
	mux.Handle(http.MethodPost+" "+BounceWebhookPath, b)
}

// ServeHTTP takes one notice.
func (b *BounceWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, read := restapi.ReadBody(w, r, b.limits)
	if !read {
		return
	}
	notice := r.Header.Get("svix-id")
	if !b.signed(notice, r.Header.Get("svix-timestamp"), r.Header.Get("svix-signature"), body) {
		restapi.WriteRefusal(w, r, http.StatusUnauthorized, errSignature.Error())
		return
	}
	var event event
	if err := json.Unmarshal(body, &event); err != nil {
		restapi.WriteRefusal(w, r, http.StatusBadRequest, "the notice is not JSON")
		return
	}
	emailID, isBounce := event.bouncedEmail()
	if !isBounce {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	err := b.markBounced.Execute(r.Context(), application.MarkEmailBouncedInput{
		Notice:    notice,
		EmailID:   emailID,
		MessageID: event.Data.EmailID,
	})
	if err != nil {
		restapi.WriteFailure(w, r, fmt.Errorf("take the bounce notice: %w", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// signed tells whether the notice carries a signature of the secret, made
// within the tolerance of now, as Svix signs: HMAC-SHA256 over the
// identifier of the notice, its timestamp and its raw body, joined by full
// stops. The header lists one signature or more, each after the version of
// its scheme (`v1,<base64>`), separated by spaces.
func (b *BounceWebhook) signed(notice, timestamp, signatures string, body []byte) bool {
	if notice == "" || !b.recent(timestamp) {
		return false
	}
	mac := hmac.New(sha256.New, b.secret.Key())
	mac.Write([]byte(notice + "." + timestamp + "."))
	mac.Write(body)
	expected := mac.Sum(nil)
	return slices.ContainsFunc(strings.Fields(signatures), func(each string) bool {
		version, signature, _ := strings.Cut(each, ",")
		decoded, err := base64.StdEncoding.DecodeString(signature)
		return version == "v1" && err == nil && hmac.Equal(decoded, expected)
	})
}

// recent tells whether the timestamp, in seconds since 1970, is within the
// tolerance of now.
func (b *BounceWebhook) recent(timestamp string) bool {
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	distance := b.clock.Now().Sub(time.Unix(seconds, 0))
	return distance.Abs() <= signatureTolerance
}

// event is what the webhook reads of an event of Resend. The address of the
// recipient, the subject and the answer of the server that refused the
// e-mail are left unread.
type event struct {
	Type string    `json:"type"`
	Data eventData `json:"data"`
}

type eventData struct {
	// EmailID is the identifier Resend gave the e-mail.
	EmailID string `json:"email_id"`
	// Tags are the tags the e-mail was sent with.
	Tags tags `json:"tags"`
}

// bouncedEmail gives the e-mail record a bounce names by its tag, and false
// for another event, or for an e-mail sent with no record of this system.
func (e event) bouncedEmail() (domain.EmailID, bool) {
	if e.Type != "email.bounced" {
		return 0, false
	}
	value, err := strconv.ParseInt(e.Data.Tags[EmailTag], 10, 64)
	if err != nil {
		return 0, false
	}
	id, err := domain.NewEmailID(value)
	return id, err == nil
}

// tags are the tags of an e-mail, by name. Resend's documentation shows them
// on an event as an object; a sending takes them as a list of names and
// values. Both are read.
type tags map[string]string

func (t *tags) UnmarshalJSON(raw []byte) error {
	var named map[string]string
	if json.Unmarshal(raw, &named) == nil {
		*t = named
		return nil
	}
	var listed []tag
	if err := json.Unmarshal(raw, &listed); err != nil {
		return err
	}
	*t = tags{}
	for _, each := range listed {
		(*t)[each.Name] = each.Value
	}
	return nil
}
