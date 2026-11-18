package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/identifier"
)

// The variables of the file storage the API reaches: the Garage of
// development, as `task api:test` reads them from .env.example.
var fileStorageVariables = []string{
	"FILE_STORAGE_URL",
	"FILE_STORAGE_PUBLIC_URL",
	"FILE_STORAGE_REGION",
	"FILE_STORAGE_BUCKET",
	"FILE_STORAGE_ACCESS_KEY_ID",
	"FILE_STORAGE_SECRET_ACCESS_KEY",
}

// withFileStorage adds the variables of the file storage of development to
// an environment.
func withFileStorage(environment map[string]string) map[string]string {
	for _, variable := range fileStorageVariables {
		environment[variable] = os.Getenv(variable)
	}
	return environment
}

// uploadBody is the body of starting the upload of an image.
func uploadBody(code, purpose, mediaType string, size int) string {
	body, _ := json.Marshal(map[string]any{"code": code, "purpose": purpose, "media_type": mediaType, "size": size})
	return string(body)
}

// startUpload asks the API to start the upload of an image.
func startUpload(t *testing.T, api, body string) received {
	t.Helper()
	return sendJSON(t, http.MethodPost, api+"/api/setup/uploads", body)
}

// aPNG is the content of a PNG of this size: its signature, then filler.
func aPNG(size int) []byte {
	content := bytes.Repeat([]byte{0x42}, size)
	copy(content, "\x89PNG\r\n\x1a\n\x00\x00\x00\x0dIHDR")
	return content
}

// sendToTheStorage sends the bytes to a signed URL with a PUT, from the origin
// of the app, as the browser does. The URL names the public host name of the
// storage; the request reaches Garage with that Host, as Caddy forwards it.
func sendToTheStorage(t *testing.T, signedURL, mediaType string, content []byte) int {
	t.Helper()
	storage, err := url.Parse(os.Getenv("FILE_STORAGE_URL"))
	require.NoError(t, err)
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	client := &http.Client{Timeout: 30 * time.Second, Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, storage.Host)
		},
	}}
	request, err := http.NewRequest(http.MethodPut, signedURL, bytes.NewReader(content))
	require.NoError(t, err)
	request.Header.Set("Content-Type", mediaType)
	request.Header.Set("Origin", os.Getenv("APP_ORIGIN"))
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	return response.StatusCode
}

// sentImage starts the upload of a PNG through the API and sends it to the
// storage. It returns the identifier of the file, as the API answered it.
func sentImage(t *testing.T, api, code, purpose string) string {
	t.Helper()
	started := startUpload(t, api, uploadBody(code, purpose, "image/png", 2048))
	require.Equal(t, http.StatusCreated, started.status, started.body)
	signedURL, _ := started.fields["upload_url"].(string)
	require.Equal(t, http.StatusOK, sendToTheStorage(t, signedURL, "image/png", aPNG(2048)))
	fileID, _ := started.fields["file_id"].(string)
	return fileID
}

func TestStartingAnUploadAnswers201WithTheFileItsSignedURLAndWhenItExpires(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	before := time.Now()

	answer := startUpload(t, api, uploadBody(code, "consultancy_logo", "image/png", 2048))

	require.Equal(t, http.StatusCreated, answer.status, answer.body)
	assert.Equal(t, "application/json", answer.mediaType)
	assert.ElementsMatch(t, []string{"file_id", "upload_url", "expires_at"}, keysOf(answer.fields))
	fileID, _ := answer.fields["file_id"].(string)
	assert.Regexp(t, `^fil_[0-9A-Za-z]{8,}$`, fileID)
	uploadURL, _ := answer.fields["upload_url"].(string)
	assert.True(t, strings.HasPrefix(uploadURL, os.Getenv("FILE_STORAGE_PUBLIC_URL")+"/"+os.Getenv("FILE_STORAGE_BUCKET")+"/"),
		"the URL is not signed for the public host name of the storage: %s", uploadURL)
	expiresAt, err := time.Parse(time.RFC3339, answer.fields["expires_at"].(string))
	require.NoError(t, err)
	assert.WithinRange(t, expiresAt, before.Add(5*time.Minute-time.Second), time.Now().Add(5*time.Minute+time.Second))
	var stored int64
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(),
		"SELECT id FROM files WHERE purpose = 'consultancy_logo' AND confirmed_at IS NULL").Scan(&stored))
	decoded, err := identifier.File.Decode(fileID)
	require.NoError(t, err)
	assert.Equal(t, stored, decoded, "the identifier answered is not the one of the file saved")
	assert.NotContains(t, answer.body, `"file_id":`+strconv.FormatInt(stored, 10), "the integer left the backend")
}

func keysOf(fields map[string]any) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	return keys
}

// The bytes never pass through the API: the browser sends them to the URL,
// and concluding the setup confirms them.
func TestAnImageSentToTheSignedURLIsKeptWhenTheSetupIsConcluded(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	logo := sentImage(t, api, code, "consultancy_logo")
	photo := sentImage(t, api, code, "user_photo")
	body := setupBody(code)
	object(body, "consultancy")["logo_file_id"] = logo
	object(body, "administrator")["photo_file_id"] = photo

	answer := concludeSetup(t, api, body)

	require.Equal(t, http.StatusNoContent, answer.status, answer.body)
	var keptLogo, keptPhoto bool
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(), `
		SELECT
			EXISTS (SELECT FROM consultancy JOIN files ON files.id = consultancy.logo_file_id
				WHERE files.purpose = 'consultancy_logo' AND files.confirmed_at IS NOT NULL),
			EXISTS (SELECT FROM users JOIN files ON files.id = users.photo_file_id
				WHERE files.purpose = 'user_photo' AND files.confirmed_at IS NOT NULL)`).Scan(&keptLogo, &keptPhoto))
	assert.True(t, keptLogo, "the consultancy has no confirmed logo")
	assert.True(t, keptPhoto, "the administrator has no confirmed photo")
}

func TestACodeThatIsNotValidStartsNoUpload(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))

	answer := startUpload(t, api, uploadBody("ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ", "consultancy_logo", "image/png", 2048))

	requireSetupCodeInvalid(t, answer)
	assert.Equal(t, 0, countOf(t, asOwner(t, temporary), "files"))
}

func TestATypeOrASizeTheContractRefusesStartsNoUpload(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	refused := map[string]struct {
		body     string
		location string
	}{
		"a GIF":                {uploadBody(code, "consultancy_logo", "image/gif", 2048), "body.media_type"},
		"over two megabytes":   {uploadBody(code, "consultancy_logo", "image/png", 2*1024*1024+1), "body.size"},
		"no byte":              {uploadBody(code, "consultancy_logo", "image/png", 0), "body.size"},
		"for something else":   {uploadBody(code, "invoice", "image/png", 2048), "body.purpose"},
		"with no type":         {`{"code":"` + code + `","purpose":"user_photo","size":10}`, "body.media_type"},
		"with a name of file":  {`{"code":"` + code + `","purpose":"user_photo","media_type":"image/png","size":10,"name":"me.png"}`, "body.name"},
		"with a fraction size": {`{"code":"` + code + `","purpose":"user_photo","media_type":"image/png","size":1.5}`, "body.size"},
	}
	for name, sent := range refused {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			answer := startUpload(t, api, sent.body)

			require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
			assert.Equal(t, "validation_failed", answer.fields["code"])
			fields, _ := answer.fields["errors"].([]any)
			require.Len(t, fields, 1, answer.body)
			named, _ := fields[0].(map[string]any)
			assert.Equal(t, sent.location, named["location"])
		})
	}
	assert.Equal(t, 0, countOf(t, asOwner(t, temporary), "files"))
}

func TestConcludingWithAFileNeverSentOrThatDoesNotExistNamesItsFieldAndSavesNothing(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	started := startUpload(t, api, uploadBody(code, "user_photo", "image/png", 2048))
	require.Equal(t, http.StatusCreated, started.status, started.body)
	neverSent, _ := started.fields["file_id"].(string)
	doesNotExist, err := identifier.File.Encode(987654)
	require.NoError(t, err)
	owner := asOwner(t, temporary)
	before := everyRow(t, owner)
	body := setupBody(code)
	object(body, "consultancy")["logo_file_id"] = doesNotExist
	object(body, "administrator")["photo_file_id"] = neverSent

	answer := concludeSetup(t, api, body)

	require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
	traceID, _ := answer.fields["trace_id"].(string)
	assert.JSONEq(t, `{
		"title": "Unprocessable Entity",
		"status": 422,
		"detail": "validation failed",
		"code": "validation_failed",
		"params": {},
		"trace_id": "`+traceID+`",
		"errors": [
			{"location": "body.consultancy.logo_file_id", "code": "file_not_found", "params": {}, "message": "no file of this kind has this identifier"},
			{"location": "body.administrator.photo_file_id", "code": "file_not_uploaded", "params": {}, "message": "the file was never sent to the storage"}
		]
	}`, answer.body)
	assert.Equal(t, before, everyRow(t, owner), "a refused setup changed what is stored")
}

// An identifier leaves the backend with its prefix, and one with another
// prefix is refused before anything is decoded. One with the prefix of a file
// that stands for no number names no file.
func TestTheIdentifierOfAFileIsReadOnTheWayIn(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	logo := sentImage(t, api, code, "consultancy_logo")
	cases := map[string]struct {
		logo, code string
	}{
		"with the prefix of another type": {"prj_" + strings.TrimPrefix(logo, "fil_"), "invalid"},
		"with no prefix":                  {strings.TrimPrefix(logo, "fil_"), "invalid"},
		"too short":                       {"fil_abc", "invalid"},
		"that stands for no number":       {"fil_" + strings.Repeat("A", 10), "file_not_found"},
	}
	for name, sent := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body := setupBody(code)
			object(body, "consultancy")["logo_file_id"] = sent.logo

			answer := concludeSetup(t, api, body)

			require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
			fields, _ := answer.fields["errors"].([]any)
			require.Len(t, fields, 1, answer.body)
			named, _ := fields[0].(map[string]any)
			assert.Equal(t, "body.consultancy.logo_file_id", named["location"])
			assert.Equal(t, sent.code, named["code"])
		})
	}
}

// Every field that breaks a rule is named at once: an identifier that names
// no file too, beside the others.
func TestAnIdentifierOfAFileThatNamesNoFileIsNamedWithTheOtherRefusedFields(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	body := setupBody(code)
	object(body, "consultancy")["logo_file_id"] = "fil_" + strings.Repeat("A", 10)
	object(body, "consultancy")["country"] = "ZZ"

	answer := concludeSetup(t, api, body)

	assert.Equal(t, map[string]any{
		"body.consultancy.country":      "country_unknown",
		"body.consultancy.logo_file_id": "file_not_found",
	}, fieldsRefused(t, answer))
}

// storedPastTheURL stores bytes for the file with the key of the API, past
// its signed URL, as a sender who sends what no URL allowed.
func storedPastTheURL(t *testing.T, temporary databasetest.Database, fileID string, content []byte) string {
	t.Helper()
	id, err := identifier.File.Decode(fileID)
	require.NoError(t, err)
	var key string
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(),
		"SELECT object_key FROM files WHERE id = $1", id).Scan(&key))
	_, err = storageClient().PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("FILE_STORAGE_BUCKET")),
		Key:    aws.String(key),
		Body:   bytes.NewReader(content),
	})
	require.NoError(t, err)
	return key
}

// storageClient reaches the Garage of development with the key of the API.
func storageClient() *s3.Client {
	return s3.New(s3.Options{
		BaseEndpoint: aws.String(os.Getenv("FILE_STORAGE_URL")),
		UsePathStyle: true,
		Region:       os.Getenv("FILE_STORAGE_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("FILE_STORAGE_ACCESS_KEY_ID"), os.Getenv("FILE_STORAGE_SECRET_ACCESS_KEY"), ""),
	})
}

func TestAnImageThatBreaksARuleInTheStorageIsNamedAndRemovedWhenTheSetupIsConcluded(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	startedLogo := startUpload(t, api, uploadBody(code, "consultancy_logo", "image/png", 2048))
	require.Equal(t, http.StatusCreated, startedLogo.status, startedLogo.body)
	startedPhoto := startUpload(t, api, uploadBody(code, "user_photo", "image/png", 2048))
	require.Equal(t, http.StatusCreated, startedPhoto.status, startedPhoto.body)
	logo, _ := startedLogo.fields["file_id"].(string)
	photo, _ := startedPhoto.fields["file_id"].(string)
	tooLarge := storedPastTheURL(t, temporary, logo, aPNG(3*1024*1024))
	notAnImage := storedPastTheURL(t, temporary, photo, []byte("%PDF-1.7 a document, not an image"))
	body := setupBody(code)
	object(body, "consultancy")["logo_file_id"] = logo
	object(body, "administrator")["photo_file_id"] = photo

	answer := concludeSetup(t, api, body)

	require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
	assert.Contains(t, answer.body, `{"location":"body.consultancy.logo_file_id","code":"file_too_large","params":{"max_size":2097152}`)
	assert.Contains(t, answer.body, `{"location":"body.administrator.photo_file_id","code":"file_type_unsupported"`)
	for _, key := range []string{tooLarge, notAnImage} {
		_, err := storageClient().HeadObject(context.Background(), &s3.HeadObjectInput{
			Bucket: aws.String(os.Getenv("FILE_STORAGE_BUCKET")), Key: aws.String(key),
		})
		var notFound *types.NotFound
		assert.ErrorAs(t, err, &notFound, "the object %s is still in the storage", key)
	}
	assert.Equal(t, 0, countOf(t, asOwner(t, temporary), "consultancy"))
}
