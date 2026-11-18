package application_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
)

// The first bytes of each type of image, as a file of that type starts.
var (
	pngStart  = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\x0dIHDR")
	jpegStart = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01}
	webpStart = []byte("RIFF\x24\x00\x00\x00WEBPVP8 ")
	pdfStart  = []byte("%PDF-1.7\n%\xe2\xe3\xcf\xd3")
)

// anImage is the content of a file of this size that starts as given.
func anImage(start []byte, size int) []byte {
	content := bytes.Repeat([]byte{0x42}, size)
	copy(content, start)
	return content
}

// uploaded is a logo whose upload started as a PNG of a kilobyte, and whose
// bytes were then sent to the storage: whatever the sender chose to send.
func uploaded(t *testing.T, stored storage, content []byte) application.FileID {
	t.Helper()
	started, err := stored.startUpload.Execute(context.Background(), aLogo("image/png", 1024))
	require.NoError(t, err)
	stored.objects.Put(stored.files.All()[0].ObjectKey(), content)
	stored.clock.Advance(time.Minute)
	return started.FileID
}

func confirmLogo(stored storage, id application.FileID) error {
	return stored.confirmUpload.Execute(context.Background(), application.ConfirmUploadInput{
		FileID: id, Purpose: application.ConsultancyLogo,
	})
}

func TestConfirmingAnImageThatWasSentMakesTheFileConfirmed(t *testing.T) {
	for name, start := range map[string][]byte{"a PNG": pngStart, "a JPEG": jpegStart, "a WebP": webpStart} {
		t.Run(name, func(t *testing.T) {
			stored := newStorage()
			id := uploaded(t, stored, anImage(start, 4096))

			err := confirmLogo(stored, id)

			require.NoError(t, err)
			confirmedAt, confirmed := stored.files.All()[0].ConfirmedAt()
			assert.True(t, confirmed)
			assert.Equal(t, noon.Add(time.Minute), confirmedAt)
			assert.True(t, stored.objects.Holds(stored.files.All()[0].ObjectKey()), "the object of a confirmed file was removed")
		})
	}
}

// The bytes of a confirmed file are kept under a key of their own, from the
// random source, which no URL was signed for, and the object the upload was
// sent to is removed.
func TestAConfirmedImageIsKeptUnderANewKeyAndTheObjectOfTheUploadIsRemoved(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 64))
	uploadKey := stored.files.All()[0].ObjectKey()

	require.NoError(t, confirmLogo(stored, id))

	kept := stored.files.All()[0].ObjectKey()
	// The second key of the counting source: the first went to the upload.
	assert.Equal(t, "101112131415161718191a1b1c1d1e1f", string(kept))
	content, found := stored.objects.Content(kept)
	require.True(t, found, "nothing is stored under the key of the confirmed file")
	assert.Equal(t, anImage(pngStart, 64), content)
	_, signed := stored.objects.SignedUpload(kept)
	assert.False(t, signed, "a URL was signed for the key of the confirmed file")
	assert.False(t, stored.objects.Holds(uploadKey), "the object the upload was sent to is still stored")
}

// Whoever holds the signed URL of the upload can send to it until it
// expires. What they send after the confirmation changes nothing of the file.
func TestSendingToTheSignedURLAfterTheConfirmationDoesNotChangeTheConfirmedImage(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 64))
	uploadKey := stored.files.All()[0].ObjectKey()
	require.NoError(t, confirmLogo(stored, id))

	stored.objects.Put(uploadKey, anImage(jpegStart, 64))

	content, found := stored.objects.Content(stored.files.All()[0].ObjectKey())
	require.True(t, found, "the confirmed file has no object")
	assert.Equal(t, anImage(pngStart, 64), content, "a PUT to the signed URL changed the confirmed image")
}

// The object the upload was sent to is removed after the commit. When that
// fails, the file is confirmed all the same, under its new key: what is left
// behind is an object no file names (ADR-0017).
func TestAConfirmationStandsWhenTheObjectOfTheUploadCannotBeRemoved(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 64))
	uploadKey := stored.files.All()[0].ObjectKey()
	stored.objects.FailRemovingWith(errors.New("the storage did not answer the removal"))

	err := confirmLogo(stored, id)

	require.NoError(t, err)
	_, confirmed := stored.files.All()[0].ConfirmedAt()
	assert.True(t, confirmed)
	kept := stored.files.All()[0].ObjectKey()
	assert.NotEqual(t, uploadKey, kept)
	content, _ := stored.objects.Content(kept)
	assert.Equal(t, anImage(pngStart, 64), content)
	assert.True(t, stored.objects.Holds(uploadKey), "the object of the upload is left behind")
}

// A copy that fails confirms nothing: the file stays under the key of its
// upload, whose object is kept, and a later confirmation can try again.
func TestAnObjectThatCannotBeCopiedConfirmsNothing(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 64))
	uploadKey := stored.files.All()[0].ObjectKey()
	failed := errors.New("the storage did not answer the copy")
	stored.objects.FailCopyingWith(failed)

	err := confirmLogo(stored, id)

	require.ErrorIs(t, err, failed)
	_, confirmed := stored.files.All()[0].ConfirmedAt()
	assert.False(t, confirmed)
	assert.Equal(t, uploadKey, stored.files.All()[0].ObjectKey())
	assert.True(t, stored.objects.Holds(uploadKey), "the object of the upload was removed")
	stored.objects.FailCopyingWith(nil)
	require.NoError(t, confirmLogo(stored, id))
}

// A transaction that does not commit leaves the file as it was, under the key
// of its upload, with its object: the copy it made is left behind.
func TestAConfirmationThatIsNotCommittedLeavesTheFileUnderTheKeyOfItsUpload(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 64))
	uploadKey := stored.files.All()[0].ObjectKey()
	lost := errors.New("the connection was lost at the commit")
	stored.transactions.FailCommitWith(lost)

	err := confirmLogo(stored, id)

	require.ErrorIs(t, err, lost)
	_, confirmed := stored.files.All()[0].ConfirmedAt()
	assert.False(t, confirmed)
	assert.Equal(t, uploadKey, stored.files.All()[0].ObjectKey())
	assert.True(t, stored.objects.Holds(uploadKey), "the object of the upload was removed before the commit")
}

func TestAnImageOfTwoMegabytesExactlyIsConfirmed(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, twoMegabytes))

	require.NoError(t, confirmLogo(stored, id))
}

// A signed URL cannot bound what is sent to it, so the size declared proves
// nothing: the storage tells the real one.
func TestAnObjectOverTheLimitIsRefusedAndRemovedFromTheStorage(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, twoMegabytes+1))

	err := confirmLogo(stored, id)

	require.ErrorAs(t, err, &application.FileTooLargeError{})
	assert.False(t, stored.objects.Holds(stored.files.All()[0].ObjectKey()), "the object over the limit is still stored")
	_, confirmed := stored.files.All()[0].ConfirmedAt()
	assert.False(t, confirmed)
}

func TestAnObjectThatIsNotAPNGAJPEGOrAWebPIsRefusedAndRemovedFromTheStorage(t *testing.T) {
	notImages := map[string][]byte{
		"a PDF":                       anImage(pdfStart, 2048),
		"text":                        []byte("<svg xmlns='http://www.w3.org/2000/svg'/>"),
		"a RIFF file that is no WebP": anImage([]byte("RIFF\x24\x00\x00\x00WAVEfmt "), 2048),
		"an empty object":             {},
		"the start of a PNG only":     pngStart[:4],
	}
	for name, content := range notImages {
		t.Run(name, func(t *testing.T) {
			stored := newStorage()
			id := uploaded(t, stored, content)

			err := confirmLogo(stored, id)

			require.ErrorAs(t, err, &application.FileTypeUnsupportedError{})
			assert.False(t, stored.objects.Holds(stored.files.All()[0].ObjectKey()), "the object that is not an image is still stored")
			_, confirmed := stored.files.All()[0].ConfirmedAt()
			assert.False(t, confirmed)
		})
	}
}

// Removing the object is an effect on the storage, so it waits for the commit
// of the transaction that checked the file (Writes in CODING_STANDARDS.md): a
// transaction that does not commit leaves the object where it was.
func TestARefusedObjectIsRemovedOnlyOnceTheTransactionThatCheckedItIsCommitted(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, twoMegabytes+1))
	lost := errors.New("the connection was lost at the commit")
	stored.transactions.FailCommitWith(lost)

	err := confirmLogo(stored, id)

	require.ErrorIs(t, err, lost)
	assert.True(t, stored.objects.Holds(stored.files.All()[0].ObjectKey()), "the object was removed before the commit")
	stored.transactions.FailCommitWith(nil)
	require.ErrorAs(t, confirmLogo(stored, id), &application.FileTooLargeError{})
	assert.False(t, stored.objects.Holds(stored.files.All()[0].ObjectKey()), "the object over the limit is still stored")
}

func TestAFileWhoseBytesWereNeverSentIsNotUploaded(t *testing.T) {
	stored := newStorage()
	started, err := stored.startUpload.Execute(context.Background(), aLogo("image/png", 10))
	require.NoError(t, err)

	err = confirmLogo(stored, started.FileID)

	require.ErrorAs(t, err, &application.FileNotUploadedError{})
	_, confirmed := stored.files.All()[0].ConfirmedAt()
	assert.False(t, confirmed)
}

func TestAFileThatDoesNotExistIsNotFound(t *testing.T) {
	stored := newStorage()
	uploaded(t, stored, anImage(pngStart, 10))

	for _, id := range []application.FileID{99, 0, -1} {
		err := confirmLogo(stored, id)

		require.ErrorAs(t, err, &application.FileNotFoundError{}, "file %d", id)
	}
}

// The identifier of the photo of a user does not stand for the logo of the
// consultancy.
func TestAFileForSomethingElseIsNotFound(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 10))

	err := stored.confirmUpload.Execute(context.Background(), application.ConfirmUploadInput{
		FileID: id, Purpose: application.UserPhoto,
	})

	require.ErrorAs(t, err, &application.FileNotFoundError{})
	_, confirmed := stored.files.All()[0].ConfirmedAt()
	assert.False(t, confirmed)
	assert.True(t, stored.objects.Holds(stored.files.All()[0].ObjectKey()))
}

// Concluding the setup can fail after its files were confirmed, and be tried
// again: the files are confirmed again, and keep their first confirmation,
// its time and its key. Nothing is copied again.
func TestAFileConfirmedTwiceKeepsItsFirstConfirmation(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 10))
	require.NoError(t, confirmLogo(stored, id))
	kept := stored.files.All()[0].ObjectKey()
	stored.clock.Advance(time.Hour)

	err := confirmLogo(stored, id)

	require.NoError(t, err)
	confirmedAt, _ := stored.files.All()[0].ConfirmedAt()
	assert.Equal(t, noon.Add(time.Minute), confirmedAt)
	assert.Equal(t, kept, stored.files.All()[0].ObjectKey())
	content, found := stored.objects.Content(kept)
	require.True(t, found, "the object of the confirmed file was removed")
	assert.Equal(t, anImage(pngStart, 10), content)
	assert.Len(t, stored.objects.Keys(), 1, "the object was copied again")
}

// No signed URL reaches the object of a confirmed file, so it cannot change:
// confirming the file again does not read the storage.
func TestAFileConfirmedBeforeIsConfirmedAgainWhileTheStorageDoesNotAnswer(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 10))
	require.NoError(t, confirmLogo(stored, id))
	stored.objects.FailWith(errors.New("the storage does not answer"))

	err := confirmLogo(stored, id)

	require.NoError(t, err)
}

func TestAStorageThatDoesNotAnswerConfirmsNothing(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 10))
	unreachable := errors.New("the storage does not answer")
	stored.objects.FailWith(unreachable)

	err := confirmLogo(stored, id)

	require.ErrorIs(t, err, unreachable)
	_, confirmed := stored.files.All()[0].ConfirmedAt()
	assert.False(t, confirmed)
}
