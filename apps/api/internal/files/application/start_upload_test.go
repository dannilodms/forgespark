package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/application/applicationtest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

var noon = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

// twoMegabytes is the largest image, in bytes.
const twoMegabytes = 2 * 1024 * 1024

// countingBytes stands in for the random source: it gives 0, 1, 2 and so on,
// so the key of the first object is known.
type countingBytes struct {
	next byte
}

func (c *countingBytes) Read(buffer []byte) (int, error) {
	for i := range buffer {
		buffer[i] = c.next
		c.next++
	}
	return len(buffer), nil
}

// storage is the files area with nothing in it, built on the in-memory
// implementations, with the clock stopped at noon.
type storage struct {
	clock         *applicationtest.Clock
	files         *applicationtest.FileRepository
	objects       *applicationtest.Storage
	startUpload   *application.StartUpload
	confirmUpload *application.ConfirmUpload
}

func newStorage() storage {
	clock := applicationtest.NewClock(noon)
	files := applicationtest.NewFileRepository()
	objects := applicationtest.NewStorage()
	transactions := applicationtest.NewTransactions(files)
	return storage{
		clock:         clock,
		files:         files,
		objects:       objects,
		startUpload:   application.NewStartUpload(transactions, files, objects, clock, &countingBytes{}),
		confirmUpload: application.NewConfirmUpload(transactions, files, objects, clock),
	}
}

func aLogo(mediaType string, size int64) application.StartUploadInput {
	return application.StartUploadInput{Purpose: application.ConsultancyLogo, MediaType: mediaType, Size: size}
}

func TestStartingAnUploadSavesTheFileAsDeclaredAndNotConfirmed(t *testing.T) {
	stored := newStorage()

	started, err := stored.startUpload.Execute(context.Background(), aLogo("image/png", 1024))

	require.NoError(t, err)
	files := stored.files.All()
	require.Len(t, files, 1)
	assert.Equal(t, started.FileID, files[0].ID())
	assert.Equal(t, application.ConsultancyLogo, files[0].Purpose())
	assert.Equal(t, application.PNG, files[0].Declared().MediaType())
	assert.Equal(t, int64(1024), files[0].Declared().Size())
	assert.Equal(t, noon, files[0].StartedAt())
	_, confirmed := files[0].ConfirmedAt()
	assert.False(t, confirmed, "a file is confirmed before its object was checked")
}

func TestStartingAnUploadAnswersAURLSignedForTheObjectThatExpiresInFiveMinutes(t *testing.T) {
	stored := newStorage()

	started, err := stored.startUpload.Execute(context.Background(), aLogo("image/webp", 2048))

	require.NoError(t, err)
	key := stored.files.All()[0].ObjectKey()
	signed, found := stored.objects.SignedUpload(key)
	require.True(t, found, "no URL was signed for the object of the file")
	assert.Equal(t, signed.URL, started.URL)
	assert.Equal(t, application.WebP, signed.Declared.MediaType(), "the URL accepts another type than the one declared")
	assert.Equal(t, int64(2048), signed.Declared.Size(), "the URL accepts another size than the one declared")
	assert.Equal(t, 5*time.Minute, signed.Validity)
	assert.Equal(t, noon.Add(5*time.Minute), started.ExpiresAt)
}

// The key comes from the random source alone, so it holds neither the name of
// the file nor an identifier.
func TestTheObjectKeyOfEachFileIsMadeOfRandomBytesOnly(t *testing.T) {
	stored := newStorage()

	_, err := stored.startUpload.Execute(context.Background(), aLogo("image/png", 10))
	require.NoError(t, err)
	_, err = stored.startUpload.Execute(context.Background(), aLogo("image/png", 10))
	require.NoError(t, err)

	files := stored.files.All()
	require.Len(t, files, 2)
	assert.Equal(t, "000102030405060708090a0b0c0d0e0f", string(files[0].ObjectKey()))
	assert.Equal(t, "101112131415161718191a1b1c1d1e1f", string(files[1].ObjectKey()))
}

func TestEachOfTheThreeImageTypesMayBeDeclared(t *testing.T) {
	for _, mediaType := range []string{"image/png", "image/jpeg", "image/webp"} {
		t.Run(mediaType, func(t *testing.T) {
			stored := newStorage()

			_, err := stored.startUpload.Execute(context.Background(), aLogo(mediaType, 10))

			require.NoError(t, err)
			assert.Equal(t, mediaType, string(stored.files.All()[0].Declared().MediaType()))
		})
	}
}

// brokenFields returns, for each field the error names, the code of the rule
// it breaks.
func brokenFields(t *testing.T, err error) map[string]string {
	t.Helper()
	var invalid kernel.InvalidFieldsError
	require.ErrorAs(t, err, &invalid)
	broken := map[string]string{}
	for _, field := range invalid.Fields {
		broken[field.Field] = field.Rule.Code()
	}
	return broken
}

func TestADeclaredTypeOtherThanAPNGAJPEGOrAWebPStartsNoUpload(t *testing.T) {
	for _, mediaType := range []string{"image/gif", "image/svg+xml", "application/pdf", "IMAGE/PNG", "image/png; charset=utf-8", ""} {
		t.Run(mediaType, func(t *testing.T) {
			stored := newStorage()

			_, err := stored.startUpload.Execute(context.Background(), aLogo(mediaType, 10))

			assert.Equal(t, map[string]string{"media_type": "file_type_unsupported"}, brokenFields(t, err))
			assert.Empty(t, stored.files.All())
			assert.Empty(t, stored.objects.SignedKeys())
		})
	}
}

func TestADeclaredSizeOverTwoMegabytesStartsNoUpload(t *testing.T) {
	stored := newStorage()

	_, err := stored.startUpload.Execute(context.Background(), aLogo("image/jpeg", twoMegabytes+1))

	assert.Equal(t, map[string]string{"size": "file_too_large"}, brokenFields(t, err))
	var invalid kernel.InvalidFieldsError
	require.ErrorAs(t, err, &invalid)
	var tooLarge application.FileTooLargeError
	require.ErrorAs(t, invalid.Fields[0].Rule, &tooLarge)
	assert.Equal(t, map[string]any{"max_size": int64(twoMegabytes)}, tooLarge.Params())
	assert.Empty(t, stored.files.All())
	assert.Empty(t, stored.objects.SignedKeys())
}

func TestAnImageOfTwoMegabytesExactlyMayBeDeclared(t *testing.T) {
	stored := newStorage()

	_, err := stored.startUpload.Execute(context.Background(), aLogo("image/jpeg", twoMegabytes))

	require.NoError(t, err)
}

func TestAnEmptyFileStartsNoUpload(t *testing.T) {
	for _, size := range []int64{0, -1} {
		stored := newStorage()

		_, err := stored.startUpload.Execute(context.Background(), aLogo("image/png", size))

		assert.Equal(t, map[string]string{"size": "file_empty"}, brokenFields(t, err))
		assert.Empty(t, stored.files.All())
	}
}

func TestATypeAndASizeThatAreBothRefusedAreNamedTogether(t *testing.T) {
	stored := newStorage()

	_, err := stored.startUpload.Execute(context.Background(), aLogo("image/gif", twoMegabytes+1))

	assert.Equal(t, map[string]string{"media_type": "file_type_unsupported", "size": "file_too_large"}, brokenFields(t, err))
}

func TestAPhotoOfAUserMayBeUploadedToo(t *testing.T) {
	stored := newStorage()

	_, err := stored.startUpload.Execute(context.Background(), application.StartUploadInput{
		Purpose: application.UserPhoto, MediaType: "image/png", Size: 10,
	})

	require.NoError(t, err)
	assert.Equal(t, application.UserPhoto, stored.files.All()[0].Purpose())
}

func TestAFileForSomethingElseStartsNoUpload(t *testing.T) {
	stored := newStorage()

	_, err := stored.startUpload.Execute(context.Background(), application.StartUploadInput{
		Purpose: application.Purpose("invoice"), MediaType: "image/png", Size: 10,
	})

	require.Error(t, err)
	assert.Empty(t, stored.files.All())
}

func TestAFileTheDatabaseRefusesToSaveGetsNoSignedURL(t *testing.T) {
	stored := newStorage()
	refused := errors.New("the database refused the row")
	stored.files.FailWith(refused)

	_, err := stored.startUpload.Execute(context.Background(), aLogo("image/png", 10))

	require.ErrorIs(t, err, refused)
	assert.Empty(t, stored.objects.SignedKeys())
}
