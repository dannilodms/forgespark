package application_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
)

// aPNG is the content of a PNG of this size: its signature, then filler.
func aPNG(size int) []byte {
	content := bytes.Repeat([]byte{0x42}, size)
	copy(content, "\x89PNG\r\n\x1a\n\x00\x00\x00\x0dIHDR")
	return content
}

// sentImage starts the upload of an image with the code, and sends its bytes
// to the storage as the browser does. It returns the identifier of the file.
func sentImage(t *testing.T, installed installation, code string, purpose application.FilePurpose, content []byte) domain.FileID {
	t.Helper()
	id := startedImage(t, installed, code, purpose)
	installed.sendTheBytesOf(t, id, content)
	return id
}

// startedImage starts the upload of an image of a kilobyte, and sends nothing.
func startedImage(t *testing.T, installed installation, code string, purpose application.FilePurpose) domain.FileID {
	t.Helper()
	started, err := installed.startSetupUpload.Execute(context.Background(), application.StartSetupUploadInput{
		Code: code, Purpose: purpose, MediaType: "image/png", Size: 1024,
	})
	require.NoError(t, err)
	return started.FileID
}

// withImages is a full setup with a logo and a photo.
func withImages(code string, logo, photo domain.FileID) application.ConcludeSetupInput {
	input := typedSetup(code)
	input.Consultancy.LogoFileID = logo
	input.Administrator.PhotoFileID = photo
	return input
}

func TestConcludingWithALogoAndAPhotoSavesTheConsultancyAndTheAdministratorWithThem(t *testing.T) {
	installed := issued(t)
	logo := sentImage(t, installed, firstCode, application.ConsultancyLogo, aPNG(1024))
	photo := sentImage(t, installed, firstCode, application.UserPhoto, aPNG(1024))

	err := installed.concludeSetup.Execute(context.Background(), withImages(firstCode, logo, photo))

	require.NoError(t, err)
	savedLogo, hasLogo := installed.consultancies.All()[0].Details().LogoFileID()
	assert.True(t, hasLogo)
	assert.Equal(t, logo, savedLogo)
	savedPhoto, hasPhoto := installed.users.All()[0].Details().PhotoFileID()
	assert.True(t, hasPhoto)
	assert.Equal(t, photo, savedPhoto)
	for _, id := range []domain.FileID{logo, photo} {
		confirmedAt, confirmed := installed.fileOf(t, id).ConfirmedAt()
		assert.True(t, confirmed, "file %d", id)
		assert.Equal(t, noon.Add(10*time.Minute), confirmedAt)
	}
}

func TestConcludingWithoutALogoOrAPhotoSavesNeither(t *testing.T) {
	installed := issued(t)

	err := installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode))

	require.NoError(t, err)
	_, hasLogo := installed.consultancies.All()[0].Details().LogoFileID()
	assert.False(t, hasLogo)
	_, hasPhoto := installed.users.All()[0].Details().PhotoFileID()
	assert.False(t, hasPhoto)
}

func TestConcludingWithAFileThatWasNeverSentOrDoesNotExistNamesItsFieldAndSavesNothing(t *testing.T) {
	installed := issued(t)
	neverSent := startedImage(t, installed, firstCode, application.UserPhoto)

	err := installed.concludeSetup.Execute(context.Background(), withImages(firstCode, 404, neverSent))

	assert.Equal(t, map[string]string{
		"consultancy.logo_file_id":    "file_not_found",
		"administrator.photo_file_id": "file_not_uploaded",
	}, brokenFields(t, err))
	requireNothingSaved(t, installed)
}

func TestThePhotoOfTheAdministratorIsNotTakenForTheLogo(t *testing.T) {
	installed := issued(t)
	photo := sentImage(t, installed, firstCode, application.UserPhoto, aPNG(1024))

	err := installed.concludeSetup.Execute(context.Background(), withImages(firstCode, photo, 0))

	assert.Equal(t, map[string]string{"consultancy.logo_file_id": "file_not_found"}, brokenFields(t, err))
	requireNothingSaved(t, installed)
}

func TestAnImageThatBreaksARuleOnceSentIsNamedRemovedAndNothingIsSaved(t *testing.T) {
	installed := issued(t)
	tooLarge := sentImage(t, installed, firstCode, application.ConsultancyLogo, aPNG(3*1024*1024))
	notAnImage := sentImage(t, installed, firstCode, application.UserPhoto, []byte("%PDF-1.7 not an image"))

	err := installed.concludeSetup.Execute(context.Background(), withImages(firstCode, tooLarge, notAnImage))

	assert.Equal(t, map[string]string{
		"consultancy.logo_file_id":    "file_too_large",
		"administrator.photo_file_id": "file_type_unsupported",
	}, brokenFields(t, err))
	requireNothingSaved(t, installed)
	assert.False(t, installed.holdsTheBytesOf(t, tooLarge), "the image over the limit is still stored")
	assert.False(t, installed.holdsTheBytesOf(t, notAnImage), "the file that is no image is still stored")
}

// An hour is not a deadline for typing: an image sent under a code that
// expired since is still accepted with a new one.
func TestAnImageSentUnderACodeThatExpiredIsAcceptedWithANewCode(t *testing.T) {
	installed := issued(t)
	logo := sentImage(t, installed, firstCode, application.ConsultancyLogo, aPNG(1024))
	installed.clock.Advance(time.Hour)
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)

	err = installed.concludeSetup.Execute(context.Background(), withImages(secondCode, logo, 0))

	require.NoError(t, err)
	savedLogo, _ := installed.consultancies.All()[0].Details().LogoFileID()
	assert.Equal(t, logo, savedLogo)
}

// A setup that failed after its images were confirmed is concluded later with
// the same images.
func TestImagesConfirmedByASetupThatFailedAreAcceptedWhenItIsConcluded(t *testing.T) {
	installed := issued(t)
	logo := sentImage(t, installed, firstCode, application.ConsultancyLogo, aPNG(1024))
	failing := errors.New("the database refuses the write")
	installed.users.FailWith(failing)
	require.ErrorIs(t, installed.concludeSetup.Execute(context.Background(), withImages(firstCode, logo, 0)), failing)
	_, confirmed := installed.fileOf(t, logo).ConfirmedAt()
	require.True(t, confirmed, "the image was not confirmed before the setup failed")
	installed.users.FailWith(nil)

	err := installed.concludeSetup.Execute(context.Background(), withImages(firstCode, logo, 0))

	require.NoError(t, err)
	savedLogo, _ := installed.consultancies.All()[0].Details().LogoFileID()
	assert.Equal(t, logo, savedLogo)
}

// Without a code that works nothing reaches the storage: no image is
// confirmed and none is removed.
func TestAWrongCodeConfirmsNoImage(t *testing.T) {
	installed := issued(t)
	logo := sentImage(t, installed, firstCode, application.ConsultancyLogo, aPNG(1024))
	tooLarge := sentImage(t, installed, firstCode, application.UserPhoto, aPNG(3*1024*1024))

	err := installed.concludeSetup.Execute(context.Background(), withImages("ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ", logo, tooLarge))

	requireSetupCodeInvalid(t, err)
	_, confirmed := installed.fileOf(t, logo).ConfirmedAt()
	assert.False(t, confirmed, "a wrong code confirmed an image")
	assert.True(t, installed.holdsTheBytesOf(t, tooLarge), "a wrong code removed an image")
}

// An identifier that names no file, such as one an inbound adapter could not
// decode, is answered with the other fields that break a rule, all at once.
func TestAnIdentifierThatNamesNoFileIsNamedWithTheOtherFieldsThatBreakARule(t *testing.T) {
	installed := issued(t)
	input := withImages(firstCode, domain.NoSuchFile, domain.NoSuchFile)
	input.Consultancy.Country = "ZZ"
	input.Administrator.Name = ""

	err := installed.concludeSetup.Execute(context.Background(), input)

	assert.Equal(t, map[string]string{
		"consultancy.country":         "country_unknown",
		"consultancy.logo_file_id":    "file_not_found",
		"administrator.name":          "required",
		"administrator.photo_file_id": "file_not_found",
	}, brokenFields(t, err))
	requireNothingSaved(t, installed)
}

// A setup code that opens the registration of a new administrator concludes
// nothing, and reaches no image of the storage on the way.
func TestACodeThatOpensTheRegistrationOfAnAdministratorConfirmsNoImageWhenItConcludes(t *testing.T) {
	installed := backIn(t)
	tooLarge := sentImage(t, installed, secondCode, application.UserPhoto, aPNG(3*1024*1024))

	err := installed.concludeSetup.Execute(context.Background(), withImages(secondCode, 0, tooLarge))

	var alreadyConcluded domain.SetupAlreadyConcludedError
	require.ErrorAs(t, err, &alreadyConcluded)
	assert.True(t, installed.holdsTheBytesOf(t, tooLarge), "an image was removed")
}

// The fields are checked first: a setup refused for one of them confirms no
// image.
func TestAFieldThatBreaksARuleIsAnsweredBeforeAnyImageIsConfirmed(t *testing.T) {
	installed := issued(t)
	logo := sentImage(t, installed, firstCode, application.ConsultancyLogo, aPNG(1024))
	input := withImages(firstCode, logo, 0)
	input.Consultancy.Country = "ZZ"

	err := installed.concludeSetup.Execute(context.Background(), input)

	assert.Equal(t, map[string]string{"consultancy.country": "country_unknown"}, brokenFields(t, err))
	_, confirmed := installed.fileOf(t, logo).ConfirmedAt()
	assert.False(t, confirmed)
}

func (i installation) fileOf(t *testing.T, id domain.FileID) *filesapplication.File {
	t.Helper()
	file, found := i.files.Find(filesapplication.FileID(id))
	require.True(t, found, "no file %d", id)
	return &file
}

// holdsTheBytesOf tells whether the storage holds an object for the file.
func (i installation) holdsTheBytesOf(t *testing.T, id domain.FileID) bool {
	t.Helper()
	return i.objects.Holds(i.fileOf(t, id).ObjectKey())
}

// sendTheBytesOf stores bytes for the file, as the browser does with its
// signed URL.
func (i installation) sendTheBytesOf(t *testing.T, id domain.FileID, content []byte) {
	t.Helper()
	i.objects.Put(i.fileOf(t, id).ObjectKey(), content)
}
