package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
)

// The shell of the app: who is logged in, with their photo, and the
// consultancy, with its logo. The browser reads both images from the file
// storage, through URLs the API signs.

// consultancyPath is the consultancy the shell shows.
const consultancyPath = "/api/consultancy"

// loggedInWithImages is a browser where Ana logged in with a login code,
// after a setup concluded with a logo for the consultancy and a photo for her.
func loggedInWithImages(t *testing.T, temporary databasetest.Database, api string) *http.Client {
	t.Helper()
	_, code := issueSetupCodeIn(t, temporary)
	body := setupBody(code)
	object(body, "consultancy")["logo_file_id"] = sentImage(t, api, code, "consultancy_logo")
	object(body, "administrator")["photo_file_id"] = sentImage(t, api, code, "user_photo")
	concluded := concludeSetup(t, api, body)
	require.Equal(t, http.StatusNoContent, concluded.status, concluded.body)
	browser := aBrowser(t)
	require.Equal(t, http.StatusAccepted, askForLoginCode(t, browser, api, "ana.souza@fabrica.example").status)
	checked := checkCode(t, browser, api, codeOfTheRequest(t, temporary))
	require.Equal(t, http.StatusNoContent, checked.status, checked.body)
	return browser
}

// readFromTheStorage reads a signed URL with a GET, as the browser does for
// an image. The URL names the public host name of the storage; the request
// reaches Garage with that Host, as Caddy forwards it.
func readFromTheStorage(t *testing.T, signedURL string) (int, []byte) {
	t.Helper()
	storage, err := url.Parse(os.Getenv("FILE_STORAGE_URL"))
	require.NoError(t, err)
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	client := &http.Client{Timeout: 30 * time.Second, Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, storage.Host)
		},
	}}
	response, err := client.Get(signedURL)
	require.NoError(t, err)
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return response.StatusCode, content
}

// requireAnImageOfTheStorage checks that a URL the API answered is signed for
// the public host name of the storage, for fifteen minutes, and reads the
// image that was sent.
func requireAnImageOfTheStorage(t *testing.T, signedURL any) {
	t.Helper()
	address, isText := signedURL.(string)
	require.True(t, isText, "no URL was answered: %v", signedURL)
	assert.True(t, strings.HasPrefix(address, os.Getenv("FILE_STORAGE_PUBLIC_URL")+"/"+os.Getenv("FILE_STORAGE_BUCKET")+"/"),
		"the URL is not signed for the public host name of the storage: %s", address)
	parsed, err := url.Parse(address)
	require.NoError(t, err)
	assert.Equal(t, "900", parsed.Query().Get("X-Amz-Expires"), "the URL does not expire in fifteen minutes")
	status, content := readFromTheStorage(t, address)
	require.Equal(t, http.StatusOK, status, string(content))
	assert.Equal(t, aPNG(2048), content)
}

func TestWhoIsLoggedInCarriesAURLThatReadsTheirPhoto(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	browser := loggedInWithImages(t, temporary, api)

	who := whoIsLoggedIn(t, browser, api)

	require.Equal(t, http.StatusOK, who.status, who.body)
	assert.ElementsMatch(t, []string{"name", "email", "photo_url", "preferred_language", "role_keys"}, keysOf(who.fields))
	requireAnImageOfTheStorage(t, who.fields["photo_url"])
}

func TestTheConsultancyIsItsNameAndAURLThatReadsItsLogo(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	browser := loggedInWithImages(t, temporary, api)

	consultancy := sendFrom(t, browser, http.MethodGet, api+consultancyPath, "", nil)

	require.Equal(t, http.StatusOK, consultancy.status, consultancy.body)
	assert.Equal(t, "application/json", consultancy.mediaType)
	assert.ElementsMatch(t, []string{"name", "logo_url"}, keysOf(consultancy.fields))
	assert.Equal(t, "Fábrica de Software", consultancy.fields["name"])
	requireAnImageOfTheStorage(t, consultancy.fields["logo_url"])
}

// The setup of loggedIn chose no image: the answers leave the URLs out.
func TestAConsultancyAndAUserWithNoImageAreAnsweredWithNoURL(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	browser, _ := loggedIn(t, temporary, api)

	consultancy := sendFrom(t, browser, http.MethodGet, api+consultancyPath, "", nil)
	who := whoIsLoggedIn(t, browser, api)

	require.Equal(t, http.StatusOK, consultancy.status, consultancy.body)
	assert.JSONEq(t, `{"name": "Fábrica de Software"}`, consultancy.body)
	require.Equal(t, http.StatusOK, who.status, who.body)
	assert.NotContains(t, who.fields, "photo_url")
}

func TestTheConsultancyIsReadByALoggedInUserOnly(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	withAUser(t, temporary, api, "Ana.Souza@Fabrica.example")

	requireNotLoggedIn(t, sendFrom(t, aBrowser(t), http.MethodGet, api+consultancyPath, "", nil))
}
