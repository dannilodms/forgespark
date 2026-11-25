package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// aDeveloper acts with no scope: reading the consultancy asks for none.
var aDeveloper = kernel.NewActor(2, nil)

// concludedWith is an installation whose setup was concluded with this logo,
// or with none when it is domain.FileID 0.
func concludedWith(t *testing.T, logo func(installation) domain.FileID) installation {
	t.Helper()
	installed := issued(t)
	input := typedSetup(firstCode)
	input.Consultancy.LogoFileID = logo(installed)
	require.NoError(t, installed.concludeSetup.Execute(context.Background(), input))
	return installed
}

// The shell of the app shows the name of the consultancy and its logo, which
// the browser reads from the storage through a URL signed for fifteen
// minutes.
func TestTheConsultancyIsItsNameAndAURLThatReadsItsLogo(t *testing.T) {
	var logo domain.FileID
	installed := concludedWith(t, func(installed installation) domain.FileID {
		logo = sentImage(t, installed, firstCode, application.ConsultancyLogo, aPNG(1024))
		return logo
	})

	consultancy, err := installed.tellConsultancy.Execute(context.Background(), aDeveloper)

	require.NoError(t, err)
	assert.Equal(t, "Fábrica de Software", consultancy.Name)
	signed, found := installed.objects.SignedDownload(installed.fileOf(t, logo).ObjectKey())
	require.True(t, found, "no download was signed for the logo")
	assert.Equal(t, signed.URL, consultancy.LogoURL)
	assert.Equal(t, 15*time.Minute, signed.Validity)
}

func TestAConsultancyWithNoLogoIsItsNameAlone(t *testing.T) {
	installed := concludedWith(t, func(installation) domain.FileID { return 0 })

	consultancy, err := installed.tellConsultancy.Execute(context.Background(), aDeveloper)

	require.NoError(t, err)
	assert.Equal(t, "Fábrica de Software", consultancy.Name)
	assert.Empty(t, consultancy.LogoURL)
	assert.Empty(t, installed.objects.SignedDownloadKeys())
}

// Nobody logs in before the setup is concluded, so a read of the consultancy
// that finds none is a failure of the system, not an answer.
func TestReadingTheConsultancyBeforeTheSetupIsConcludedIsAFailure(t *testing.T) {
	installed := issued(t)

	_, err := installed.tellConsultancy.Execute(context.Background(), aDeveloper)

	require.Error(t, err)
}

func TestAConsultancyWhoseLogoCannotBeSignedIsAFailure(t *testing.T) {
	installed := concludedWith(t, func(installed installation) domain.FileID {
		return sentImage(t, installed, firstCode, application.ConsultancyLogo, aPNG(1024))
	})
	unreachable := errors.New("the storage does not answer")
	installed.objects.FailWith(unreachable)

	_, err := installed.tellConsultancy.Execute(context.Background(), aDeveloper)

	require.ErrorIs(t, err, unreachable)
}
