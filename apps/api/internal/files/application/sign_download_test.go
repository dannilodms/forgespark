package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A confirmed file is read by the browser from the storage, through a URL
// signed for the key its bytes are kept under, which expires in fifteen
// minutes.
func TestTheDownloadOfAConfirmedFileIsSignedForItsObjectForFifteenMinutes(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 1024))
	require.NoError(t, confirmLogo(stored, id))
	kept := stored.files.All()[0].ObjectKey()

	url, found, err := stored.signDownload.Execute(context.Background(), id)

	require.NoError(t, err)
	require.True(t, found)
	signed, signedFound := stored.objects.SignedDownload(kept)
	require.True(t, signedFound, "no download was signed for the object the file is kept under")
	assert.Equal(t, signed.URL, url)
	assert.Equal(t, 15*time.Minute, signed.Validity)
}

// The object of a file that was never confirmed may be anything its sender
// chose to send, so nothing reads it.
func TestNoDownloadIsSignedForAFileNeverConfirmedOrThatDoesNotExist(t *testing.T) {
	stored := newStorage()
	unconfirmed := uploaded(t, stored, anImage(pngStart, 1024))

	_, unconfirmedFound, err := stored.signDownload.Execute(context.Background(), unconfirmed)
	require.NoError(t, err)
	_, unknownFound, err := stored.signDownload.Execute(context.Background(), unconfirmed+1)
	require.NoError(t, err)

	assert.False(t, unconfirmedFound, "a download was signed for a file never confirmed")
	assert.False(t, unknownFound, "a download was signed for a file that does not exist")
	assert.Empty(t, stored.objects.SignedDownloadKeys())
}

func TestADownloadTheStorageCannotSignIsAFailure(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 1024))
	require.NoError(t, confirmLogo(stored, id))
	unreachable := errors.New("the storage does not answer")
	stored.objects.FailWith(unreachable)

	_, _, err := stored.signDownload.Execute(context.Background(), id)

	require.ErrorIs(t, err, unreachable)
}

func TestADownloadOfAFileTheDatabaseCannotReadIsAFailure(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 1024))
	require.NoError(t, confirmLogo(stored, id))
	unreachable := errors.New("the database does not answer")
	stored.files.FailWith(unreachable)

	_, _, err := stored.signDownload.Execute(context.Background(), id)

	require.ErrorIs(t, err, unreachable)
	assert.Empty(t, stored.objects.SignedDownloadKeys())
}
