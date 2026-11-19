// Package oidctest drives the login page of the Dex of development over
// HTTP, as a person in a browser does, for the tests of the login through the
// local login provider: the tests of the `oidc` adapter, and the tests of the
// binary.
package oidctest

import (
	"errors"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// Person is someone the Dex of development knows, with the password they log
// in with there. The configuration of Dex in compose.yaml lists them.
type Person struct {
	Email    string
	Password string
}

var (
	// Administrator is the administrator of the seed, at the default mailbox.
	Administrator = Person{Email: "seed+admin@fabrica.example", Password: "local-provider-admin"}
	// Developer is the developer of the seed, who is not a user yet.
	Developer = Person{Email: "seed+developer@fabrica.example", Password: "local-provider-developer"}
)

// formAction is where the login page of Dex sends its form.
var formAction = regexp.MustCompile(`<form method="post" action="([^"]+)"`)

// LogIn opens the address in the browser, which leads to the login page of
// Dex, logs the person in there, and returns the address Dex sends the
// browser back to: the callback, which starts with the given address, with
// the code and the state in its query. The browser is not sent there. The
// cookies of every answer on the way, the session of the API included, go
// to the jar of the client.
func LogIn(t *testing.T, client *http.Client, address string, person Person, callback string) *url.URL {
	t.Helper()
	browser := *client
	browser.CheckRedirect = func(next *http.Request, _ []*http.Request) error {
		if strings.HasPrefix(next.URL.String(), callback) {
			return http.ErrUseLastResponse
		}
		return nil
	}

	page, err := browser.Get(address)
	if err != nil {
		t.Fatalf("open the login page of Dex: %v", err)
	}
	action := actionOf(t, page)
	answer, err := browser.PostForm(action.String(), url.Values{"login": {person.Email}, "password": {person.Password}})
	if err != nil {
		t.Fatalf("log in at Dex: %v", err)
	}
	defer answer.Body.Close()
	location, err := answer.Location()
	if err != nil || !strings.HasPrefix(location.String(), callback) {
		t.Fatalf("Dex did not send the browser back to %s: it answered %d", callback, answer.StatusCode)
	}
	return location
}

// actionOf reads where the login form of the page is sent, against the
// address of the page.
func actionOf(t *testing.T, page *http.Response) *url.URL {
	t.Helper()
	defer page.Body.Close()
	body, err := io.ReadAll(page.Body)
	if err != nil {
		t.Fatalf("read the login page of Dex: %v", err)
	}
	found := formAction.FindSubmatch(body)
	if page.StatusCode != http.StatusOK || found == nil {
		t.Fatalf("the browser did not reach the login page of Dex: %d at %s", page.StatusCode, page.Request.URL.Redacted())
	}
	action, err := page.Request.URL.Parse(html.UnescapeString(string(found[1])))
	if err != nil {
		t.Fatalf("read the form of the login page of Dex: %v", errors.Unwrap(err))
	}
	return action
}
