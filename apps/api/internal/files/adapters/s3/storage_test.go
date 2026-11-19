package s3_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/adapters/s3"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// The s3 adapter against the Garage of development, which `task services`
// starts. Each test writes objects under keys of its own and removes them.

// settingsOf reads where the storage is from the variables `task api:test`
// loads from .env.example.
func settingsOf(t *testing.T) s3.Settings {
	t.Helper()
	read := func(variable string) string {
		value, found := os.LookupEnv(variable)
		require.True(t, found, "%s is not set: run the tests with `task api:test`, which reads .env.example", variable)
		return value
	}
	return s3.Settings{
		URL:             read("FILE_STORAGE_URL"),
		PublicURL:       read("FILE_STORAGE_PUBLIC_URL"),
		Region:          read("FILE_STORAGE_REGION"),
		Bucket:          read("FILE_STORAGE_BUCKET"),
		AccessKeyID:     read("FILE_STORAGE_ACCESS_KEY_ID"),
		SecretAccessKey: config.NewSecret(read("FILE_STORAGE_SECRET_ACCESS_KEY")),
		// Tests run with no telemetry.
		TracerProvider: nil,
	}
}

// aKey is the key of an object no other test uses. It is removed when the
// test ends.
func aKey(t *testing.T, storage *s3.Storage) domain.ObjectKey {
	t.Helper()
	key, err := domain.NewObjectKey(rand.Reader)
	require.NoError(t, err)
	t.Cleanup(func() { _ = storage.Remove(context.Background(), key) })
	return key
}

func declared(t *testing.T, mediaType string, size int64) domain.DeclaredImage {
	t.Helper()
	image, err := domain.DeclareImage(mediaType, size)
	require.NoError(t, err)
	return image
}

// browser sends requests as a browser on the machine does to the public host
// name of the storage, and as Caddy forwards them: to Garage, with the Host
// header the browser wrote, which the signature covers.
func browser(t *testing.T, settings s3.Settings) *http.Client {
	t.Helper()
	storage, err := url.Parse(settings.URL)
	require.NoError(t, err)
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, storage.Host)
			},
		},
	}
}

// put sends the bytes to the URL with a PUT, as the browser does.
func put(t *testing.T, client *http.Client, signedURL, mediaType string, content []byte) (int, string) {
	t.Helper()
	request, err := http.NewRequest(http.MethodPut, signedURL, bytes.NewReader(content))
	require.NoError(t, err)
	request.Header.Set("Content-Type", mediaType)
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return response.StatusCode, string(body)
}

// storedPast stores bytes under the key with the key of the API, past any
// signed URL, as a sender stores what no URL allowed: the storage holds
// whatever reaches it.
func storedPast(t *testing.T, settings s3.Settings, key domain.ObjectKey, content []byte) {
	t.Helper()
	_, err := sdkClient(settings).PutObject(context.Background(), &awss3.PutObjectInput{
		Bucket: aws.String(settings.Bucket),
		Key:    aws.String(string(key)),
		Body:   bytes.NewReader(content),
	})
	require.NoError(t, err)
}

// contentOf reads the whole object under the key with the key of the API.
func contentOf(t *testing.T, settings s3.Settings, key domain.ObjectKey) []byte {
	t.Helper()
	object, err := sdkClient(settings).GetObject(context.Background(), &awss3.GetObjectInput{
		Bucket: aws.String(settings.Bucket),
		Key:    aws.String(string(key)),
	})
	require.NoError(t, err)
	defer object.Body.Close()
	content, err := io.ReadAll(object.Body)
	require.NoError(t, err)
	return content
}

// sdkClient reaches the storage with the key of the API, past the adapter.
func sdkClient(settings s3.Settings) *awss3.Client {
	return awss3.New(awss3.Options{
		BaseEndpoint: aws.String(settings.URL),
		UsePathStyle: true,
		Region:       settings.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(settings.AccessKeyID, settings.SecretAccessKey.Reveal(), ""),
	})
}

// aPNG is the content of a PNG of this size: its signature, then filler.
func aPNG(size int) []byte {
	return aPNGFilledWith(size, 0x42)
}

// aPNGFilledWith is the content of a PNG of this size: its signature, then
// the filler given.
func aPNGFilledWith(size int, filler byte) []byte {
	content := bytes.Repeat([]byte{filler}, size)
	copy(content, "\x89PNG\r\n\x1a\n\x00\x00\x00\x0dIHDR")
	return content
}

func TestTheURLIsSignedForThePublicHostNameOfTheStorage(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)

	signed, err := storage.SignUpload(context.Background(), aKey(t, storage), declared(t, "image/png", 1024), 5*time.Minute)

	require.NoError(t, err)
	parsed, err := url.Parse(signed)
	require.NoError(t, err)
	public, err := url.Parse(settings.PublicURL)
	require.NoError(t, err)
	assert.Equal(t, public.Scheme, parsed.Scheme)
	assert.Equal(t, public.Host, parsed.Host)
	assert.True(t, strings.HasPrefix(parsed.Path, "/"+settings.Bucket+"/"), "the bucket is not in the path: %s", parsed.Path)
	assert.Equal(t, "300", parsed.Query().Get("X-Amz-Expires"), "the URL does not expire in five minutes")
}

func TestAnImageSentToTheSignedURLIsStoredAsDeclared(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	key := aKey(t, storage)
	signed, err := storage.SignUpload(context.Background(), key, declared(t, "image/png", 1024), 5*time.Minute)
	require.NoError(t, err)

	status, body := put(t, browser(t, settings), signed, "image/png", aPNG(1024))

	require.Equal(t, http.StatusOK, status, body)
	object, found, err := storage.Inspect(context.Background(), key)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, int64(1024), object.Size)
	assert.Equal(t, aPNG(1024)[:domain.FirstBytesRead], object.FirstBytes)
}

// The signature covers the type and the size declared, so the storage takes
// nothing else through the URL.
func TestTheSignedURLRefusesAnotherTypeOrAnotherSize(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	cases := map[string]struct {
		mediaType string
		content   []byte
	}{
		"a larger image":  {"image/png", aPNG(4096)},
		"a smaller image": {"image/png", aPNG(512)},
		"another type":    {"image/jpeg", aPNG(1024)},
	}
	for name, sent := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			key := aKey(t, storage)
			signed, err := storage.SignUpload(context.Background(), key, declared(t, "image/png", 1024), 5*time.Minute)
			require.NoError(t, err)

			status, body := put(t, browser(t, settings), signed, sent.mediaType, sent.content)

			assert.Equal(t, http.StatusForbidden, status, body)
			_, found, err := storage.Inspect(context.Background(), key)
			require.NoError(t, err)
			assert.False(t, found, "the storage kept what the URL was not signed for")
		})
	}
}

func TestASignedURLThatExpiredIsRefused(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	key := aKey(t, storage)
	signed, err := storage.SignUpload(context.Background(), key, declared(t, "image/png", 1024), time.Second)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	status, body := put(t, browser(t, settings), signed, "image/png", aPNG(1024))

	// Garage answers 400, "Date is too old", where AWS answers 403.
	assert.True(t, status >= 400 && status < 500, "an expired URL was answered %d: %s", status, body)
	_, found, err := storage.Inspect(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, found, "the storage kept what was sent to an expired URL")
}

// preflight asks the storage, as a browser does before it sends a PUT to
// another origin, whether a page of this origin may send one.
func preflight(t *testing.T, client *http.Client, signedURL, origin string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodOptions, signedURL, nil)
	require.NoError(t, err)
	request.Header.Set("Origin", origin)
	request.Header.Set("Access-Control-Request-Method", http.MethodPut)
	request.Header.Set("Access-Control-Request-Headers", "content-type")
	response, err := client.Do(request)
	require.NoError(t, err)
	_ = response.Body.Close()
	return response
}

// The bucket has the cross-origin rule the step of compose.yaml gives it. A
// browser asks before it sends a file from a page of another origin, and
// sends nothing when the storage refuses.
func TestTheStorageLetsABrowserSendAFileFromTheOriginOfTheAppOnly(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	appOrigin, found := os.LookupEnv("APP_ORIGIN")
	require.True(t, found, "APP_ORIGIN is not set: run the tests with `task api:test`")
	storage := s3.NewStorage(settings)
	key := aKey(t, storage)
	signed, err := storage.SignUpload(context.Background(), key, declared(t, "image/png", 1024), 5*time.Minute)
	require.NoError(t, err)
	client := browser(t, settings)

	fromTheApp := preflight(t, client, signed, appOrigin)
	fromElsewhere := preflight(t, client, signed, "http://elsewhere.localhost")

	require.Equal(t, http.StatusOK, fromTheApp.StatusCode)
	assert.Equal(t, appOrigin, fromTheApp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, fromTheApp.Header.Get("Access-Control-Allow-Methods"), http.MethodPut)
	assert.Contains(t, strings.ToLower(fromTheApp.Header.Get("Access-Control-Allow-Headers")), "content-type")
	assert.Equal(t, http.StatusForbidden, fromElsewhere.StatusCode)
}

func TestInspectingAnObjectReadsItsSizeAndItsFirstBytesOnly(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	key := aKey(t, storage)
	// Larger than any image: the storage holds whatever was sent.
	large := aPNG(3 * 1024 * 1024)
	storedPast(t, settings, key, large)

	object, found, err := storage.Inspect(context.Background(), key)

	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, int64(len(large)), object.Size)
	assert.Equal(t, large[:domain.FirstBytesRead], object.FirstBytes)
}

func TestInspectingAShortObjectReadsAllOfIt(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	key := aKey(t, storage)
	storedPast(t, settings, key, []byte("GIF8"))

	object, found, err := storage.Inspect(context.Background(), key)

	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, int64(4), object.Size)
	assert.Equal(t, []byte("GIF8"), object.FirstBytes)
}

func TestInspectingAnEmptyObjectReadsNoByte(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	key := aKey(t, storage)
	storedPast(t, settings, key, []byte{})

	object, found, err := storage.Inspect(context.Background(), key)

	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, int64(0), object.Size)
	assert.Empty(t, object.FirstBytes)
}

func TestInspectingAKeyWithNoObjectFindsNothing(t *testing.T) {
	t.Parallel()
	storage := s3.NewStorage(settingsOf(t))

	_, found, err := storage.Inspect(context.Background(), aKey(t, storage))

	require.NoError(t, err)
	assert.False(t, found)
}

func TestCopyingAnObjectKeepsItsBytesUnderTheOtherKeyToo(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	from := aKey(t, storage)
	to := aKey(t, storage)
	storedPast(t, settings, from, aPNG(2048))

	err := storage.Copy(context.Background(), from, to)

	require.NoError(t, err)
	assert.Equal(t, aPNG(2048), contentOf(t, settings, to))
	assert.Equal(t, aPNG(2048), contentOf(t, settings, from))
}

// Confirming an upload copies its object to a key of its own and removes the
// object the upload was sent to. The URL signed for the upload still works
// until it expires, and a PUT to it then writes under the key of the upload
// only: the copy keeps the bytes that were confirmed.
func TestAPutToTheSignedURLAfterItsObjectWasCopiedAndRemovedLeavesTheCopyAsItWas(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	uploadKey := aKey(t, storage)
	kept := aKey(t, storage)
	signed, err := storage.SignUpload(context.Background(), uploadKey, declared(t, "image/png", 1024), 5*time.Minute)
	require.NoError(t, err)
	client := browser(t, settings)
	status, body := put(t, client, signed, "image/png", aPNG(1024))
	require.Equal(t, http.StatusOK, status, body)
	require.NoError(t, storage.Copy(context.Background(), uploadKey, kept))
	require.NoError(t, storage.Remove(context.Background(), uploadKey))

	status, body = put(t, client, signed, "image/png", aPNGFilledWith(1024, 0x43))

	require.Equal(t, http.StatusOK, status, "the URL no longer works, so this test proves nothing: %s", body)
	assert.Equal(t, aPNG(1024), contentOf(t, settings, kept), "a PUT to the signed URL changed the copy")
	assert.Equal(t, aPNGFilledWith(1024, 0x43), contentOf(t, settings, uploadKey))
}

func TestRemovingAnObjectLeavesNothingUnderItsKey(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	key := aKey(t, storage)
	storedPast(t, settings, key, aPNG(2048))

	err := storage.Remove(context.Background(), key)

	require.NoError(t, err)
	_, found, err := storage.Inspect(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestRemovingAKeyWithNoObjectIsNoFailure(t *testing.T) {
	t.Parallel()
	storage := s3.NewStorage(settingsOf(t))

	err := storage.Remove(context.Background(), aKey(t, storage))

	assert.NoError(t, err)
}

// listed lists the whole bucket and returns the objects of these keys, as
// the listing gave them, and how many pages it took. The bucket is shared by
// every test and every checkout, so it may hold objects of others: the test
// looks at its own keys only.
func listed(t *testing.T, storage *s3.Storage, keys ...domain.ObjectKey) (map[domain.ObjectKey][]domain.ListedObject, int) {
	t.Helper()
	found := map[domain.ObjectKey][]domain.ListedObject{}
	pages := 0
	err := storage.List(context.Background(), func(objects []domain.ListedObject) error {
		pages++
		for _, object := range objects {
			found[object.Key] = append(found[object.Key], object)
		}
		return nil
	})
	require.NoError(t, err)
	mine := map[domain.ObjectKey][]domain.ListedObject{}
	for _, key := range keys {
		mine[key] = found[key]
	}
	return mine, pages
}

// The listing goes through every page of the bucket, and gives each object
// once, with the time it was last written.
func TestTheListingGivesEachObjectOnceWithTheTimeItWasLastWritten(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	// Two keys a page, so three objects take two pages at least.
	s3.ListInPagesOf(storage, 2)
	keys := []domain.ObjectKey{aKey(t, storage), aKey(t, storage), aKey(t, storage)}
	// The clock of the storage may differ a little from the clock of the
	// machine.
	before := time.Now().Add(-5 * time.Second)
	for _, key := range keys {
		storedPast(t, settings, key, aPNG(64))
	}
	after := time.Now().Add(5 * time.Second)

	found, pages := listed(t, storage, keys...)

	assert.GreaterOrEqual(t, pages, 2, "the listing stopped at the first page")
	for _, key := range keys {
		require.Len(t, found[key], 1, "an object is not listed once")
		assert.Equal(t, key, found[key][0].Key)
		assert.WithinRange(t, found[key][0].LastModified, before, after)
		assert.Equal(t, time.UTC, found[key][0].LastModified.Location())
	}
}

// Confirming a file copies its object to the key the file keeps it under.
// The copy is written when it is made, not when the upload was sent, so the
// cleanup of the storage, which removes the objects no confirmed file names
// once they are a day old, finds it young while the confirmation commits.
func TestACopyIsListedWithTheTimeItWasMade(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	from := aKey(t, storage)
	to := aKey(t, storage)
	storedPast(t, settings, from, aPNG(64))
	time.Sleep(1500 * time.Millisecond)

	require.NoError(t, storage.Copy(context.Background(), from, to))

	found, _ := listed(t, storage, from, to)
	require.Len(t, found[from], 1)
	require.Len(t, found[to], 1)
	assert.GreaterOrEqual(t, found[to][0].LastModified.Sub(found[from][0].LastModified), time.Second,
		"the copy is listed with the time of the object it was copied from")
}

func TestAnObjectRemovedIsListedNoMore(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	storage := s3.NewStorage(settings)
	key := aKey(t, storage)
	storedPast(t, settings, key, aPNG(64))
	require.NoError(t, storage.Remove(context.Background(), key))

	found, _ := listed(t, storage, key)

	assert.Empty(t, found[key])
}

// The secret key keeps its secret type until the adapter hands it to the SDK,
// so settings printed by mistake show no key.
func TestTheSettingsOfTheStoragePrintNoSecretKey(t *testing.T) {
	settings := s3.Settings{
		URL: "http://localhost:3900", PublicURL: "http://files.localhost", Region: "garage", Bucket: "files",
		AccessKeyID: "GK0123", SecretAccessKey: config.NewSecret("a-secret-key-to-hide"), TracerProvider: nil,
	}

	for _, printed := range []string{fmt.Sprint(settings), fmt.Sprintf("%+v", settings), fmt.Sprintf("%#v", settings)} {
		assert.NotContains(t, printed, "a-secret-key-to-hide")
	}
}

// standIn is a storage that answers as the handler says, in place of Garage,
// for what Garage never answers on demand.
func standIn(t *testing.T, handler http.HandlerFunc) *s3.Storage {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return s3.NewStorage(s3.Settings{
		URL: server.URL, PublicURL: server.URL, Region: "garage", Bucket: "files",
		AccessKeyID: "GK0123", SecretAccessKey: config.NewSecret("a-secret-key"), TracerProvider: nil,
	})
}

// A failed call names the storage it could not reach and no key of an
// object: its error goes into the logs of the job or of the request.
func TestACallThatFailsNamesTheStorageAndNoKeyOfAnObject(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	// Port 1 of the machine has nothing behind it.
	settings.URL = "http://127.0.0.1:1"
	storage := s3.NewStorage(settings)
	key := domain.ObjectKey("0123456789abcdef0123456789abcdef")
	other := domain.ObjectKey("fedcba9876543210fedcba9876543210")
	calls := map[string]func() error{
		"Inspect": func() error { _, _, err := storage.Inspect(context.Background(), key); return err },
		"Copy":    func() error { return storage.Copy(context.Background(), key, other) },
		"Remove":  func() error { return storage.Remove(context.Background(), key) },
		"List": func() error {
			return storage.List(context.Background(), func([]domain.ListedObject) error { return nil })
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := call()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "127.0.0.1:1")
			assert.NotContains(t, err.Error(), string(key))
			assert.NotContains(t, err.Error(), string(other))
		})
	}
}

// Each page of the listing after the first asks for the keys after a token,
// which the storage makes from the last key it gave (Garage encodes it in
// base64). A page that fails names neither.
func TestAListingThatFailsOnALaterPageNamesNoKey(t *testing.T) {
	t.Parallel()
	key := "0123456789abcdef0123456789abcdef"
	token := "]" + base64.StdEncoding.EncodeToString([]byte(key))
	storage := standIn(t, func(answer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("continuation-token") != "" {
			// The connection is cut, as a storage that goes away does.
			cutConnection(answer)
			return
		}
		answer.Header().Set("Content-Type", "application/xml")
		_, _ = fmt.Fprintf(answer, `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>files</Name><KeyCount>1</KeyCount>
<MaxKeys>1000</MaxKeys><IsTruncated>true</IsTruncated>
<Contents><Key>%s</Key><LastModified>2026-11-16T12:00:00.000Z</LastModified><Size>64</Size></Contents>
<NextContinuationToken>%s</NextContinuationToken></ListBucketResult>`, key, token)
	})
	var first []domain.ListedObject

	err := storage.List(context.Background(), func(objects []domain.ListedObject) error {
		first = append(first, objects...)
		return nil
	})

	require.Error(t, err)
	assert.Equal(t, []domain.ListedObject{{Key: domain.ObjectKey(key), LastModified: time.Date(2026, time.November, 16, 12, 0, 0, 0, time.UTC)}}, first)
	assert.NotContains(t, err.Error(), key)
	assert.NotContains(t, err.Error(), url.QueryEscape(token))
	assert.NotContains(t, err.Error(), strings.TrimPrefix(token, "]"))
}

// cutConnection closes the connection of the request without an answer.
func cutConnection(answer http.ResponseWriter) {
	connection, _, err := http.NewResponseController(answer).Hijack()
	if err != nil {
		return
	}
	_ = connection.Close()
}

// What the storage answers is part of the error, except a key it repeats.
func TestAnAnswerOfTheStorageThatNamesTheKeyIsNotRepeated(t *testing.T) {
	t.Parallel()
	key := "0123456789abcdef0123456789abcdef"
	storage := standIn(t, refusingWithTheKey(key))

	err := storage.Remove(context.Background(), domain.ObjectKey(key))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "AccessDenied")
	assert.NotContains(t, err.Error(), key)
}

// refusingWithTheKey answers every request with a refusal that repeats the
// key in its message.
func refusingWithTheKey(key string) http.HandlerFunc {
	return func(answer http.ResponseWriter, _ *http.Request) {
		answer.Header().Set("Content-Type", "application/xml")
		answer.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprintf(answer, `<?xml version="1.0" encoding="UTF-8"?>
<Error><Code>AccessDenied</Code><Message>no right to remove %s</Message><Key>%s</Key></Error>`, key, key)
	}
}

// The SDK records on the span of a call its raw error, with the URL of the
// request, and what the storage answered. Neither carries a key of an object.
func TestTheSpansOfACallThatFailsCarryNoKeyOfAnObject(t *testing.T) {
	t.Parallel()
	key := domain.ObjectKey("0123456789abcdef0123456789abcdef")
	repeating := httptest.NewServer(refusingWithTheKey(string(key)))
	t.Cleanup(repeating.Close)
	cases := map[string]string{
		"a storage that does not answer": "http://127.0.0.1:1",
		"a storage that repeats the key": repeating.URL,
	}
	for name, address := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			spans := tracetest.NewInMemoryExporter()
			settings := s3.Settings{
				URL: address, PublicURL: address, Region: "garage", Bucket: "files",
				AccessKeyID: "GK0123", SecretAccessKey: config.NewSecret("a-secret-key"),
				TracerProvider: sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans)),
			}

			err := s3.NewStorage(settings).Remove(context.Background(), key)

			require.Error(t, err)
			recorded := fmt.Sprint(spans.GetSpans().Snapshots())
			require.Contains(t, recorded, "exception.message", "the span recorded no error")
			assert.NotContains(t, recorded, string(key))
		})
	}
}

func TestAStorageThatDoesNotAnswerFailsWithinItsTimeout(t *testing.T) {
	t.Parallel()
	settings := settingsOf(t)
	// Port 1 of the machine has nothing behind it.
	settings.URL = "http://127.0.0.1:1"
	storage := s3.NewStorage(settings)

	_, _, err := storage.Inspect(context.Background(), domain.ObjectKey("absent"))

	require.Error(t, err)
	assert.NotContains(t, err.Error(), settings.SecretAccessKey.Reveal())
}
