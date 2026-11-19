package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application/applicationtest"
	emailapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
)

// The expected codes were computed apart from this code, with Python's hmac
// module, from the key of these tests and the random values countingBytes
// gives the first and the second request.
func TestTheDeliveryOfTheEmailGivesTheCodeOfItsRequestAndHowLongItWorks(t *testing.T) {
	s := newSystem(t)
	s.registered(t, "ana.souza@fabrica.example", "pt-BR")
	require.NoError(t, s.ask("ana.souza@fabrica.example"))
	s.clock.Advance(time.Minute)
	require.NoError(t, s.ask("ana.souza@fabrica.example"))

	first, err := s.loginCodes.ContentOf(context.Background(), emailapplication.LoginCodeRequest(1))
	require.NoError(t, err)
	second, err := s.loginCodes.ContentOf(context.Background(), emailapplication.LoginCodeRequest(2))
	require.NoError(t, err)

	assert.Equal(t, emailapplication.LoginCodeContent{Code: "774421", Lifetime: 10 * time.Minute}, first)
	assert.Equal(t, emailapplication.LoginCodeContent{Code: "865246", Lifetime: 10 * time.Minute}, second)
}

func TestTheContentNeedsAKeyOfAtLeast32Bytes(t *testing.T) {
	_, err := application.NewLoginCodeContents(applicationtest.NewLoginCodeRequestRepository(), make([]byte, 31))

	assert.Error(t, err)
}

func TestTheDeliveryGetsNoContentForARequestThatDoesNotExist(t *testing.T) {
	s := newSystem(t)

	_, err := s.loginCodes.ContentOf(context.Background(), emailapplication.LoginCodeRequest(37))

	assert.Error(t, err)
}
