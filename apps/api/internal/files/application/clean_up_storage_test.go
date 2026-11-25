package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/files/application"
	"github.com/dannilodms/forgespark/apps/api/internal/files/domain"
)

// A key no file names, as a failed removal or a confirmation that did not
// commit leaves behind (ADR-0017).
const (
	leftBehind      domain.ObjectKey = "0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f"
	otherLeftBehind domain.ObjectKey = "1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e"
)

func cleanUp(t *testing.T, stored storage) {
	t.Helper()
	require.NoError(t, stored.cleanUpStorage.Execute(context.Background()))
}

func TestTheCleanupRemovesAnObjectNoFileNamesOnceItIsMoreThanADayOld(t *testing.T) {
	stored := newStorage()
	stored.objects.Put(leftBehind, anImage(pngStart, 64))
	stored.clock.Advance(24*time.Hour + time.Second)

	cleanUp(t, stored)

	assert.False(t, stored.objects.Holds(leftBehind))
}

func TestTheCleanupKeepsAnObjectNoFileNamesForADay(t *testing.T) {
	stored := newStorage()
	stored.objects.Put(leftBehind, anImage(pngStart, 64))
	stored.clock.Advance(24 * time.Hour)

	cleanUp(t, stored)

	assert.True(t, stored.objects.Holds(leftBehind))
}

// confirmed is a logo uploaded at noon and confirmed a minute later.
func confirmed(t *testing.T, stored storage) application.FileID {
	t.Helper()
	id := uploaded(t, stored, anImage(pngStart, 64))
	require.NoError(t, confirmLogo(stored, id))
	return id
}

// The object of a confirmed file is named by it, however old it is.
func TestTheCleanupNeverTouchesAConfirmedFileOrItsObject(t *testing.T) {
	stored := newStorage()
	id := confirmed(t, stored)
	before, _ := stored.files.Find(id)
	stored.clock.Advance(30 * 24 * time.Hour)

	cleanUp(t, stored)

	after, found := stored.files.Find(id)
	require.True(t, found, "the confirmed file was removed")
	assert.Equal(t, before, after)
	content, found := stored.objects.Content(after.ObjectKey())
	require.True(t, found, "the object of the confirmed file was removed")
	assert.Equal(t, anImage(pngStart, 64), content)
}

func TestTheCleanupRemovesAnUploadNeverConfirmedMoreThanADayAfterItStartedWithItsObject(t *testing.T) {
	stored := newStorage()
	uploaded(t, stored, anImage(pngStart, 64))
	uploadKey := stored.files.All()[0].ObjectKey()
	// The upload started at noon; the clock is a minute past noon.
	stored.clock.Advance(24*time.Hour - time.Minute + time.Second)

	cleanUp(t, stored)

	assert.Empty(t, stored.files.All(), "the file of the abandoned upload is kept")
	assert.False(t, stored.objects.Holds(uploadKey), "the object of the abandoned upload is kept")
}

func TestTheCleanupKeepsAnUploadNeverConfirmedForADayAfterItStarted(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 64))
	uploadKey := stored.files.All()[0].ObjectKey()
	stored.clock.Advance(24*time.Hour - time.Minute)

	cleanUp(t, stored)

	_, found := stored.files.Find(id)
	assert.True(t, found, "the file of an upload a day old was removed")
	assert.True(t, stored.objects.Holds(uploadKey), "the object of an upload a day old was removed")
	// It can still be confirmed.
	require.NoError(t, confirmLogo(stored, id))
}

// The bytes of an upload reach the storage after it started, while its URL
// works. The file goes a day after the start, and the object a day after it
// was written, at the next run.
func TestTheObjectOfAnAbandonedUploadIsRemovedOnceItIsADayOldItself(t *testing.T) {
	stored := newStorage()
	_, err := stored.startUpload.Execute(context.Background(), aLogo("image/png", 64))
	require.NoError(t, err)
	uploadKey := stored.files.All()[0].ObjectKey()
	stored.clock.Advance(4 * time.Minute)
	stored.objects.Put(uploadKey, anImage(pngStart, 64))
	stored.clock.Advance(24*time.Hour - 3*time.Minute)

	cleanUp(t, stored)

	assert.Empty(t, stored.files.All(), "the file of the abandoned upload is kept")
	assert.True(t, stored.objects.Holds(uploadKey), "an object younger than a day was removed")
	stored.clock.Advance(time.Hour)
	cleanUp(t, stored)
	assert.False(t, stored.objects.Holds(uploadKey), "the object of the abandoned upload is kept")
}

// When the object an upload was sent to cannot be removed after the
// confirmation, it is left behind: the file names its copy (ADR-0017).
func TestTheObjectOfAnUploadLeftBehindAfterItsConfirmationIsRemovedOnceItIsADayOld(t *testing.T) {
	stored := newStorage()
	id := uploaded(t, stored, anImage(pngStart, 64))
	uploadKey := stored.files.All()[0].ObjectKey()
	stored.objects.FailRemovingWith(errors.New("the storage did not answer the removal"))
	require.NoError(t, confirmLogo(stored, id))
	stored.objects.FailRemovingWith(nil)
	kept, _ := stored.files.Find(id)
	stored.clock.Advance(24 * time.Hour)

	cleanUp(t, stored)

	assert.False(t, stored.objects.Holds(uploadKey), "the object left behind is kept")
	assert.True(t, stored.objects.Holds(kept.ObjectKey()), "the object of the confirmed file was removed")
	_, found := stored.files.Find(id)
	assert.True(t, found, "the confirmed file was removed")
}

// aStorageWithEverything holds, a day and an hour after noon: a confirmed
// file, an abandoned upload, an upload started an hour ago, an object no file
// names written at noon and one written an hour ago. More objects than a page
// of the listing holds.
func aStorageWithEverything(t *testing.T) (stored storage, confirmedFile, youngUpload application.FileID) {
	t.Helper()
	stored = newStorage()
	confirmedFile = confirmed(t, stored)
	sentNow(t, stored)
	stored.objects.Put(leftBehind, anImage(pngStart, 64))
	stored.clock.Advance(24 * time.Hour)
	youngUpload = sentNow(t, stored)
	stored.objects.Put(otherLeftBehind, anImage(pngStart, 64))
	stored.clock.Advance(time.Hour - time.Minute)
	return stored, confirmedFile, youngUpload
}

// sentNow starts the upload of a logo and sends its bytes at once.
func sentNow(t *testing.T, stored storage) application.FileID {
	t.Helper()
	started, err := stored.startUpload.Execute(context.Background(), aLogo("image/png", 64))
	require.NoError(t, err)
	file, _ := stored.files.Find(started.FileID)
	stored.objects.Put(file.ObjectKey(), anImage(pngStart, 64))
	return started.FileID
}

func TestTheCleanupRemovesTheOldObjectsNoConfirmedFileNamesAndTheAbandonedUploadsOnly(t *testing.T) {
	stored, confirmedFile, youngUpload := aStorageWithEverything(t)
	require.Len(t, stored.objects.Keys(), 5)

	cleanUp(t, stored)

	files := stored.files.All()
	require.Len(t, files, 2)
	assert.Equal(t, confirmedFile, files[0].ID())
	assert.Equal(t, youngUpload, files[1].ID())
	assert.ElementsMatch(t, []domain.ObjectKey{files[0].ObjectKey(), files[1].ObjectKey(), otherLeftBehind}, stored.objects.Keys())
}

func TestASecondCleanupOverTheSameStateChangesNothing(t *testing.T) {
	stored, _, _ := aStorageWithEverything(t)
	cleanUp(t, stored)
	files, keys := stored.files.All(), stored.objects.Keys()

	cleanUp(t, stored)

	assert.Equal(t, files, stored.files.All())
	assert.Equal(t, keys, stored.objects.Keys())
}

// The files go first: a storage that cannot be listed leaves the abandoned
// uploads removed from the database, and their objects for the next run.
func TestAStorageThatCannotBeListedFailsTheCleanupAfterTheAbandonedFilesWereRemoved(t *testing.T) {
	stored, _, _ := aStorageWithEverything(t)
	unreachable := errors.New("the storage does not answer")
	stored.objects.FailListingWith(unreachable)

	err := stored.cleanUpStorage.Execute(context.Background())

	require.ErrorIs(t, err, unreachable)
	assert.Len(t, stored.files.All(), 2)
	assert.Len(t, stored.objects.Keys(), 5)
	stored.objects.FailListingWith(nil)
	cleanUp(t, stored)
	assert.Len(t, stored.objects.Keys(), 3)
}

// Without the database, the cleanup cannot tell which objects a confirmed
// file names, so it removes none.
func TestADatabaseThatDoesNotAnswerFailsTheCleanupAndNothingIsRemoved(t *testing.T) {
	stored, _, _ := aStorageWithEverything(t)
	unreachable := errors.New("the database does not answer")
	stored.files.FailWith(unreachable)

	err := stored.cleanUpStorage.Execute(context.Background())

	require.ErrorIs(t, err, unreachable)
	assert.Len(t, stored.files.All(), 3)
	assert.Len(t, stored.objects.Keys(), 5)
}

func TestARemovalThatFailsFailsTheCleanupAndTheNextRunRemovesTheObject(t *testing.T) {
	stored := newStorage()
	stored.objects.Put(leftBehind, anImage(pngStart, 64))
	stored.clock.Advance(25 * time.Hour)
	refused := errors.New("the storage did not answer the removal")
	stored.objects.FailRemovingWith(refused)

	err := stored.cleanUpStorage.Execute(context.Background())

	require.ErrorIs(t, err, refused)
	assert.True(t, stored.objects.Holds(leftBehind))
	stored.objects.FailRemovingWith(nil)
	cleanUp(t, stored)
	assert.False(t, stored.objects.Holds(leftBehind))
}
