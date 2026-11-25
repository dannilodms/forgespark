package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
)

// aLogoOf is the start of the upload of a logo of this size, with a code.
func aLogoOf(code string, size int64) application.StartSetupUploadInput {
	return application.StartSetupUploadInput{Code: code, Purpose: application.ConsultancyLogo, MediaType: "image/png", Size: size}
}

func TestWithAValidCodeStartingAnUploadAnswersTheFileAndItsSignedURL(t *testing.T) {
	installed := issued(t)

	started, err := installed.startSetupUpload.Execute(context.Background(), aLogoOf(firstCode, 1024))

	require.NoError(t, err)
	files := installed.files.All()
	require.Len(t, files, 1)
	assert.Equal(t, domain.FileID(files[0].ID()), started.FileID)
	assert.Equal(t, application.ConsultancyLogo, files[0].Purpose())
	signed, found := installed.objects.SignedUpload(files[0].ObjectKey())
	require.True(t, found)
	assert.Equal(t, signed.URL, started.URL)
	assert.Equal(t, noon.Add(10*time.Minute+5*time.Minute), started.ExpiresAt, "the URL expires five minutes after the upload started")
}

func TestThePhotoOfTheAdministratorMayBeUploadedToo(t *testing.T) {
	installed := issued(t)

	_, err := installed.startSetupUpload.Execute(context.Background(), application.StartSetupUploadInput{
		Code: firstCode, Purpose: application.UserPhoto, MediaType: "image/jpeg", Size: 2048,
	})

	require.NoError(t, err)
	assert.Equal(t, application.UserPhoto, installed.files.All()[0].Purpose())
}

// The way back in registers an administrator and leaves the consultancy as it
// is, so its setup code starts the upload of a photo and never of a logo.
func TestACodeThatOpensTheRegistrationOfAnAdministratorStartsNoUploadOfTheLogo(t *testing.T) {
	installed := backIn(t)

	_, err := installed.startSetupUpload.Execute(context.Background(), aLogoOf(secondCode, 1024))

	var alreadyConcluded domain.SetupAlreadyConcludedError
	require.ErrorAs(t, err, &alreadyConcluded)
	assert.Equal(t, "setup_already_concluded", alreadyConcluded.Code())
	assert.Empty(t, installed.files.All(), "a file was saved")
	assert.Empty(t, installed.objects.SignedKeys(), "a URL was signed")
	_, err = installed.startSetupUpload.Execute(context.Background(), application.StartSetupUploadInput{
		Code: secondCode, Purpose: application.UserPhoto, MediaType: "image/png", Size: 1024,
	})
	require.NoError(t, err, "the photo of the new administrator")
}

func TestTheCodeMayBeTypedInAnyFormItIsAcceptedIn(t *testing.T) {
	installed := issued(t)

	_, err := installed.startSetupUpload.Execute(context.Background(), aLogoOf("0l234 56789 abcde fghjk", 1024))

	require.NoError(t, err)
}

func TestACodeThatIsNotValidStartsNoUpload(t *testing.T) {
	cases := map[string]func(t *testing.T, installed installation) (typed string){
		"wrong":            func(*testing.T, installation) string { return "ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ" },
		"not a setup code": func(*testing.T, installation) string { return "open sesame" },
		"expired": func(_ *testing.T, installed installation) string {
			installed.clock.Advance(50 * time.Minute)
			return firstCode
		},
		"ended by a new one": func(t *testing.T, installed installation) string {
			_, err := installed.issueSetupCode.Execute(context.Background())
			require.NoError(t, err)
			return firstCode
		},
	}
	for name, arrange := range cases {
		t.Run(name, func(t *testing.T) {
			installed := issued(t)
			typed := arrange(t, installed)

			_, err := installed.startSetupUpload.Execute(context.Background(), aLogoOf(typed, 1024))

			requireSetupCodeInvalid(t, err)
			assert.Empty(t, installed.files.All(), "a file was saved")
			assert.Empty(t, installed.objects.SignedKeys(), "a URL was signed")
		})
	}
}

func TestTheRulesOfAnImageAreCheckedWhenTheUploadStarts(t *testing.T) {
	installed := issued(t)

	_, err := installed.startSetupUpload.Execute(context.Background(), application.StartSetupUploadInput{
		Code: firstCode, Purpose: application.ConsultancyLogo, MediaType: "image/gif", Size: 3 * 1024 * 1024,
	})

	assert.Equal(t, map[string]string{"media_type": "file_type_unsupported", "size": "file_too_large"}, brokenFields(t, err))
	assert.Empty(t, installed.files.All())
}

// A wrong code is answered as such before anything else, so that a caller
// without a code learns nothing of the rules.
func TestAWrongCodeIsAnsweredBeforeTheRulesOfAnImage(t *testing.T) {
	installed := issued(t)

	_, err := installed.startSetupUpload.Execute(context.Background(), application.StartSetupUploadInput{
		Code: "ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ", Purpose: application.ConsultancyLogo, MediaType: "image/gif", Size: 10,
	})

	requireSetupCodeInvalid(t, err)
}
