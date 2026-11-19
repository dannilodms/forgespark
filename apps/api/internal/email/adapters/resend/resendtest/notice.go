// Package resendtest writes and signs the notices of a bounce as Resend does,
// for the tests of the webhook that takes them: the ones of the adapter and
// the ones of the binary.
package resendtest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"
)

// Secret is a signing secret as Resend shows one, `whsec_` followed by the
// key in base64. It is made up, and signs nothing outside the tests.
var Secret = "whsec_" + base64.StdEncoding.EncodeToString([]byte("a key of the tests, made up here"))

// Headers returns the headers Resend sends a notice with, as Svix writes
// them: the identifier of the notice, the time it was sent and its signature
// with the secret. They alternate names and values.
func Headers(secret, notice string, sentAt time.Time, body string) []string {
	timestamp := strconv.FormatInt(sentAt.Unix(), 10)
	return []string{
		"svix-id", notice,
		"svix-timestamp", timestamp,
		"svix-signature", "v1," + Signature(secret, notice, timestamp, body),
		"Content-Type", "application/json",
	}
}

// Signature signs the notice with the secret, as Svix does: HMAC-SHA256 with
// the key the secret carries in base64, over the identifier, the timestamp
// and the body, joined by full stops, in base64.
func Signature(secret, notice, timestamp, body string) string {
	key, err := base64.StdEncoding.DecodeString(secret[len("whsec_"):])
	if err != nil {
		panic("resendtest: the secret is not whsec_ followed by base64")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(notice + "." + timestamp + "." + body))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// Bounce returns the body of the `email.bounced` event of an e-mail, as the
// documentation of Resend shows it: the identifier Resend gave the e-mail,
// the tags it was sent with, the identifier of its record among them, and
// the address of its recipient.
func Bounce(resendID string, emailID int64, recipient string) string {
	return Event("email.bounced", resendID, map[string]string{"email_id": strconv.FormatInt(emailID, 10)}, recipient)
}

// Event returns the body of an event of this type about an e-mail, with
// these tags.
func Event(eventType, resendID string, tags map[string]string, recipient string) string {
	body, err := json.Marshal(map[string]any{
		"type":       eventType,
		"created_at": "2026-11-18T12:00:01.126Z",
		"data": map[string]any{
			"created_at": "2026-11-18T12:00:00.894Z",
			"email_id":   resendID,
			"message_id": "<111-222-333@email.fabrica.example>",
			"from":       "Fábrica Exemplo <no-reply@fabrica.example>",
			"to":         []string{recipient},
			"subject":    "Your login code",
			"bounce": map[string]any{
				"message": "The recipient's email address " + recipient + " does not exist.",
				"subType": "General",
				"type":    "Permanent",
			},
			"tags": tags,
		},
	})
	if err != nil {
		panic("resendtest: " + err.Error())
	}
	return string(body)
}
