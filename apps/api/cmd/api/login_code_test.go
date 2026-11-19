package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	accessdomain "github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// The path where a browser asks for a login code.
const loginCodePath = "/api/session/login_code"

// asked is the answer to asking for a login code, with its cookies.
type asked struct {
	received
	header  http.Header
	cookies []*http.Cookie
}

// askForLoginCode asks the API for a login code for the e-mail, from a browser
// that keeps its cookies in the jar.
func askForLoginCode(t *testing.T, client *http.Client, api, email string) asked {
	t.Helper()
	body, err := json.Marshal(map[string]string{"email": email})
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, api+loginCodePath, strings.NewReader(string(body)))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	answer := asked{
		received: received{status: response.StatusCode, mediaType: response.Header.Get("Content-Type"), body: string(raw)},
		header:   response.Header,
		cookies:  response.Cookies(),
	}
	_ = json.Unmarshal(raw, &answer.fields)
	return answer
}

// aBrowser is a client that keeps the cookies it receives.
func aBrowser(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &http.Client{Jar: jar, Timeout: 30 * time.Second}
}

// withAUser concludes the setup on the temporary database, through the API,
// with an administrator of this e-mail in pt-BR.
func withAUser(t *testing.T, temporary databasetest.Database, api, email string) {
	t.Helper()
	_, code := issueSetupCodeIn(t, temporary)
	body := setupBody(code)
	object(body, "administrator")["email"] = email
	concluded := concludeSetup(t, api, body)
	require.Equal(t, http.StatusNoContent, concluded.status, concluded.body)
}

// anAddress is an e-mail no other test uses, so the tests share Mailpit.
func anAddress() string {
	return "login+" + strings.ToLower(rand.Text()[:12]) + "@fabrica.example"
}

func sessionCookie(t *testing.T, answer asked) *http.Cookie {
	t.Helper()
	position := slices.IndexFunc(answer.cookies, func(cookie *http.Cookie) bool { return cookie.Name == "session" })
	require.GreaterOrEqual(t, position, 0, "the answer sets no cookie of the session: %v", answer.header.Values("Set-Cookie"))
	return answer.cookies[position]
}

// The answer is the same for an active user, an e-mail of no user and a
// deactivated user: 202, no body, and the cookie of the session, which
// remembers the browser.
func TestAskingForALoginCodeAnswersTheSameForEveryEmail(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	withAUser(t, temporary, api, "Ana.Souza@Fabrica.example")
	owner := asOwner(t, temporary)
	_, err := owner.Exec(context.Background(), `INSERT INTO users (id, name, email, country, national_id, preferred_language, deactivated)
		VALUES (nextval('users_id_seq'), 'Carla Dias', 'carla.dias@fabrica.example', 'BR', '1', 'es', true)`)
	require.NoError(t, err)

	answers := map[string]asked{}
	for _, email := range []string{"ANA.souza@fabrica.example", "bruno.lima@fabrica.example", "carla.dias@fabrica.example"} {
		answers[email] = askForLoginCode(t, aBrowser(t), api, email)
	}

	for email, answer := range answers {
		assert.Equal(t, http.StatusAccepted, answer.status, "%s: %s", email, answer.body)
		assert.Empty(t, answer.body, email)
		cookie := sessionCookie(t, answer)
		assert.True(t, cookie.HttpOnly, email)
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite, email)
		assert.Equal(t, "/", cookie.Path, email)
		assert.False(t, cookie.Secure, "the app of development is served over HTTP: %s", email)
		assert.InDelta(t, (7 * 24 * time.Hour).Seconds(), float64(cookie.MaxAge), 60, email)
	}
	// Only the active user has a request, an e-mail and its job.
	assert.Equal(t, 1, countOf(t, owner, "login_code_requests"))
	var recipient, language string
	require.NoError(t, owner.QueryRow(context.Background(), "SELECT recipient, language FROM emails").Scan(&recipient, &language))
	assert.Equal(t, "ana.souza@fabrica.example", recipient)
	assert.Equal(t, "pt-BR", language)
	var arguments string
	require.NoError(t, owner.QueryRow(context.Background(), "SELECT args::text FROM river_job WHERE kind = 'deliver_email'").Scan(&arguments))
	assert.NotContains(t, arguments, "ana.souza")
}

// The limit counts the asks for the typed address, so the e-mail of a user
// and an e-mail of no user meet it the same way (ADR-0011).
func TestASecondLoginCodeWithinAMinuteAnswersTheRuleOfTheLimitForEveryEmail(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	withAUser(t, temporary, api, "ana.souza@fabrica.example")

	for _, email := range []string{"ana.souza@fabrica.example", "bruno.lima@fabrica.example"} {
		browser := aBrowser(t)
		require.Equal(t, http.StatusAccepted, askForLoginCode(t, browser, api, email).status)

		again := askForLoginCode(t, browser, api, strings.ToUpper(email))

		require.Equal(t, http.StatusUnprocessableEntity, again.status, "%s: %s", email, again.body)
		assert.Equal(t, "application/problem+json", again.mediaType)
		assert.Equal(t, "login_code_limit_reached", again.fields["code"], email)
	}
	owner := asOwner(t, temporary)
	assert.Equal(t, 1, countOf(t, owner, "login_code_requests"))
	assert.Equal(t, 2, countOf(t, owner, "login_code_asks"), "a refused ask was counted")
}

// Two asks for one address at the same moment are counted one after the
// other, so one gets the code and the other the limit: never a failure of the
// unique index of the open requests (ticket 03).
func TestTwoAsksForOneAddressAtTheSameMomentAnswerTheCodeAndTheLimit(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	withAUser(t, temporary, api, "ana.souza@fabrica.example")
	statuses := make(chan int, 2)
	ask := func() int {
		response, err := http.Post(api+loginCodePath, "application/json", strings.NewReader(`{"email": "ana.souza@fabrica.example"}`))
		if err != nil {
			return 0
		}
		defer response.Body.Close()
		return response.StatusCode
	}

	for range 2 {
		go func() { statuses <- ask() }()
	}

	assert.ElementsMatch(t, []int{http.StatusAccepted, http.StatusUnprocessableEntity}, []int{<-statuses, <-statuses})
	assert.Equal(t, 1, countOf(t, asOwner(t, temporary), "login_code_requests"))
}

func TestAnEmailThatIsNotAnAddressIsRefusedOnItsField(t *testing.T) {
	t.Parallel()
	api := startAPI(t, environmentOf(databasetest.Migrated(t)))

	for typed, code := range map[string]string{
		"ana.souza":                       "invalid",
		"":                                "required",
		"Ana <ana.souza@fabrica.example>": "invalid",
		"ana.souza@fabrica.example, bruno@fabrica.example": "invalid",
	} {
		answer := askForLoginCode(t, aBrowser(t), api, typed)

		require.Equal(t, http.StatusUnprocessableEntity, answer.status, "%q: %s", typed, answer.body)
		assert.Equal(t, "validation_failed", answer.fields["code"], typed)
		errors, _ := answer.fields["errors"].([]any)
		require.Len(t, errors, 1, typed)
		field, _ := errors[0].(map[string]any)
		assert.Equal(t, "body.email", field["location"], typed)
		assert.Equal(t, code, field["code"], typed)
		assert.NotContains(t, answer.body, "ana.souza", "the answer repeats the e-mail")
	}
}

// The cookie goes over HTTPS only where the app is served over HTTPS.
func TestTheCookieOfTheSessionIsSecureWhenTheAppIsServedOverHTTPS(t *testing.T) {
	t.Parallel()
	environment := environmentOf(databasetest.Migrated(t))
	environment["APP_ORIGIN"] = "https://factory.example.com"
	api := startAPI(t, environment)

	answer := askForLoginCode(t, aBrowser(t), api, "bruno.lima@fabrica.example")

	require.Equal(t, http.StatusAccepted, answer.status, answer.body)
	assert.True(t, sessionCookie(t, answer).Secure)
}

// sixDigits finds the code in the plain-text part of the e-mail.
var sixDigits = regexp.MustCompile(`(?m)^\s+(\d{6})\s*$`)

// The whole way of a login code on the machine: the API saves the request and
// the job, the worker computes the code and sends the e-mail to Mailpit, in
// the preferred language of the user. The code in the e-mail is the code of
// the request, which only the key and the request give.
func TestALoginCodeAskedForReachesTheMailboxOfTheUserInTheirLanguage(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	address := anAddress()
	withAUser(t, temporary, api, address)
	require.Equal(t, http.StatusAccepted, askForLoginCode(t, aBrowser(t), api, address).status)

	worker := launchWorker(t, workerEnvironmentOf(t, temporary))
	message := readFromMailpit(t, address)
	worker.signal()
	require.Equal(t, 0, <-worker.status, "%s", worker.logs)

	assert.Equal(t, "Seu código de login", message.Subject)
	assert.Equal(t, address, message.To[0].Address)
	found := sixDigits.FindStringSubmatch(message.Text)
	require.NotNil(t, found, "no code in the e-mail: %s", message.Text)
	assert.Equal(t, codeOfTheRequest(t, temporary), found[1])
	assert.Contains(t, message.HTML, found[1])
	assert.Contains(t, message.Text, "10 minutos")
	var status string
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(), "SELECT status FROM emails").Scan(&status))
	assert.Equal(t, "sent", status)
}

// codeOfTheRequest computes the code of the one login code request of the
// database with the key of the tests.
func codeOfTheRequest(t *testing.T, temporary databasetest.Database) string {
	t.Helper()
	var id int64
	var random []byte
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(),
		"SELECT id, random_value FROM login_code_requests").Scan(&id, &random))
	raw, err := hex.DecodeString(theLoginCodeKey)
	require.NoError(t, err)
	key, err := accessdomain.NewLoginCodeKey(raw)
	require.NoError(t, err)
	request := accessdomain.RestoreLoginCodeRequest(accessdomain.StoredLoginCodeRequest{
		ID: accessdomain.LoginCodeRequestID(id), RandomValue: random,
	})
	return request.Code(key).Reveal()
}

// mailpitMessage is what the API of Mailpit answers about one e-mail.
type mailpitMessage struct {
	To      []struct{ Address string }
	Subject string
	Text    string
	HTML    string
}

// readFromMailpit waits for the one e-mail to the address in the Mailpit of
// development, at MAILPIT_PORT, and reads it.
func readFromMailpit(t *testing.T, address string) mailpitMessage {
	t.Helper()
	mailpit := "http://localhost:" + requiredVariable(t, "MAILPIT_PORT") + "/api/v1"
	var found struct {
		Messages []struct{ ID string }
	}
	require.Eventually(t, func() bool {
		getMailpit(t, mailpit+"/search?query="+url.QueryEscape("to:"+address), &found)
		return len(found.Messages) == 1
	}, 30*time.Second, 100*time.Millisecond, "Mailpit has no e-mail to the address")
	var message mailpitMessage
	getMailpit(t, mailpit+"/message/"+found.Messages[0].ID, &message)
	return message
}

// requiredVariable is a variable `task api:test` reads from .env.example.
func requiredVariable(t *testing.T, variable string) string {
	t.Helper()
	value := os.Getenv(variable)
	require.NotEmpty(t, value, "%s is not set: run the tests with `task api:test`, which reads .env.example", variable)
	return value
}

func getMailpit(t *testing.T, address string, into any) {
	t.Helper()
	response, err := http.Get(address)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode, address)
	require.NoError(t, json.NewDecoder(response.Body).Decode(into))
}
