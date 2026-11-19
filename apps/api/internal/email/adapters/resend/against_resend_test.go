package resend_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/resend"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// The tests against Resend itself. They call a service outside the
// repository, so they run in the scheduled workflow `resend.yml` only, never
// on a pull request, with a key of their own in RESEND_TEST_API_KEY, which no
// other workflow and no .env.example value sets: without it they skip.
// `task api:test:resend` runs them.
//
// They send to Resend's test addresses, which deliver nothing and act out a
// delivery or a bounce, from Resend's own address for an account with no
// domain of its own.

const (
	testFromAddress = "onboarding@resend.dev"
	deliveredTo     = "delivered@resend.dev"
	bouncedTo       = "bounced@resend.dev"
)

func testKey(t *testing.T) config.Secret {
	t.Helper()
	key := os.Getenv("RESEND_TEST_API_KEY")
	if key == "" {
		t.Skip("RESEND_TEST_API_KEY is not set: the tests against Resend run in the scheduled workflow only")
	}
	return config.NewSecret(key)
}

func realSender(key config.Secret) *resend.Sender {
	return resend.NewSender(resend.Settings{
		URL:            resend.URL,
		APIKey:         key,
		FromAddress:    testFromAddress,
		FromName:       "Fábrica Exemplo",
		TracerProvider: nil,
	})
}

// aRecord is the identifier of an e-mail record no other run used in the 24
// hours Resend keeps an idempotency key.
func aRecord(offset int64) domain.EmailID {
	return domain.EmailID(time.Now().UnixMilli()*10 + offset)
}

func messageTo(id domain.EmailID, to string) domain.Message {
	return domain.Message{
		EmailID: id,
		To:      to,
		Subject: "Seu código de login",
		HTML:    `<!DOCTYPE html><html lang="pt-BR"><body><p>Digite este código:</p><p><b>482915</b></p></body></html>`,
		Text:    "Digite este código:\n\n    482915\n",
	}
}

// retrieved is what Resend tells of an e-mail it took.
type retrieved struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	LastEvent string `json:"last_event"`
	Tags      []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"tags"`
}

func retrieve(t *testing.T, key config.Secret, id string) retrieved {
	t.Helper()
	email, err := read(key, id)
	require.NoError(t, err)
	return email
}

// read asks Resend what it tells of the e-mail.
func read(key config.Secret, id string) (retrieved, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, resend.URL+"/emails/"+id, nil)
	if err != nil {
		return retrieved{}, err
	}
	request.Header.Set("Authorization", "Bearer "+key.Reveal())
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return retrieved{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return retrieved{}, fmt.Errorf("the API of Resend answered %d", response.StatusCode)
	}
	var email retrieved
	err = json.NewDecoder(response.Body).Decode(&email)
	return email, err
}

func TestAgainstResendAnEmailIsTakenWithItsTagAndSentOnceForItsRecord(t *testing.T) {
	key := testKey(t)
	sender := realSender(key)
	message := messageTo(aRecord(1), deliveredTo)

	first, err := sender.Send(context.Background(), message)
	require.NoError(t, err)
	again, err := sender.Send(context.Background(), message)
	require.NoError(t, err)

	assert.NotEmpty(t, first)
	assert.Equal(t, first, again, "the idempotency key makes the second sending answer the first one")
	email := retrieve(t, key, first)
	assert.Equal(t, first, email.ID)
	require.Len(t, email.Tags, 1)
	assert.Equal(t, resend.EmailTag, email.Tags[0].Name)
	assert.Equal(t, message.EmailID, mustParse(t, email.Tags[0].Value))
	assert.Contains(t, email.From, testFromAddress)
	t.Logf("Resend shows the sender as %q", email.From)
}

func TestAgainstResendAnEmailToTheBouncingAddressBounces(t *testing.T) {
	key := testKey(t)

	id, err := realSender(key).Send(context.Background(), messageTo(aRecord(2), bouncedTo))

	require.NoError(t, err)
	require.Eventually(t, func() bool {
		email, err := read(key, id)
		return err == nil && email.LastEvent == "bounced"
	}, 2*time.Minute, 5*time.Second, "Resend reported no bounce of the e-mail %s", id)
}

func TestAgainstResendAWrongKeyIsRefusedWithItsStatusAndName(t *testing.T) {
	testKey(t)

	_, err := realSender(config.NewSecret("re_not_a_key_of_any_account")).Send(context.Background(), messageTo(aRecord(3), deliveredTo))

	require.Error(t, err)
	assert.Regexp(t, `^the API of Resend did not take the e-mail: answered (400|401|403) [a-z_]+$`, err.Error())
	assert.NotContains(t, err.Error(), deliveredTo)
}

func mustParse(t *testing.T, value string) domain.EmailID {
	t.Helper()
	var id int64
	require.NoError(t, json.Unmarshal([]byte(value), &id))
	return domain.EmailID(id)
}
