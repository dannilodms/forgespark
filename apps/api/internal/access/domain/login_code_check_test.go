package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
)

// rightCode is the code of the request asked with the identifier 37, the
// random value counting from 100 and the key counting from 0, computed apart
// (see TestTheCodeIsSixDigitsComputedFromTheKeyTheIdentifierAndTheRandomValue).
const rightCode = "025080"

func requireInvalid(t *testing.T, err error, attemptsLeft int) {
	t.Helper()
	var invalid domain.LoginCodeInvalidError
	require.ErrorAs(t, err, &invalid)
	assert.Equal(t, "login_code_invalid", invalid.Code())
	assert.Equal(t, map[string]any{"attempts_left": attemptsLeft}, invalid.Params())
}

func requireEnded(t *testing.T, err error) {
	t.Helper()
	var ended domain.LoginCodeEndedError
	require.ErrorAs(t, err, &ended)
	assert.Equal(t, "login_code_ended", ended.Code())
}

func TestTheRightCodeWithinTenMinutesEndsTheRequestAndLogsTheUserIn(t *testing.T) {
	request := asked(t, 37, countingFrom(100))
	at := noon.Add(10*time.Minute - time.Second)

	err := request.Check(rightCode, keyFrom(t, 0), at)

	require.NoError(t, err)
	endedAt, ended := request.EndedAt()
	assert.True(t, ended, "a code works once")
	assert.Equal(t, at, endedAt)
	assert.Equal(t, 0, request.Attempts())
}

func TestTheThirdWrongCodeEndsTheRequest(t *testing.T) {
	request := asked(t, 37, countingFrom(100))
	key := keyFrom(t, 0)

	requireInvalid(t, request.Check("111111", key, noon.Add(time.Minute)), 2)
	requireInvalid(t, request.Check("02508", key, noon.Add(2*time.Minute)), 1)
	_, endedEarly := request.EndedAt()
	thirdAt := noon.Add(3 * time.Minute)
	third := request.Check("025081", key, thirdAt)

	assert.False(t, endedEarly, "a request ended before the third wrong code")
	requireEnded(t, third)
	endedAt, ended := request.EndedAt()
	assert.True(t, ended)
	assert.Equal(t, thirdAt, endedAt)
	assert.Equal(t, 3, request.Attempts())
	requireEnded(t, request.Check(rightCode, key, noon.Add(4*time.Minute)))
	assert.Equal(t, 3, request.Attempts(), "the attempts of a request pass three")
}

// A request that does not work answers the right code as it answers a wrong
// one, so whoever types it learns nothing of why: expired, ended by a new
// request, or used.
func TestTheRightCodeOfARequestThatNoLongerWorksIsAWrongAttempt(t *testing.T) {
	cases := map[string]struct {
		prepare func(request *domain.LoginCodeRequest)
		at      time.Time
	}{
		"expired at ten minutes": {prepare: func(*domain.LoginCodeRequest) {}, at: noon.Add(10 * time.Minute)},
		"ended by a new request": {prepare: func(request *domain.LoginCodeRequest) { request.End(noon.Add(time.Minute)) }, at: noon.Add(2 * time.Minute)},
		"used already": {
			prepare: func(request *domain.LoginCodeRequest) { request.End(noon.Add(time.Minute)) },
			at:      noon.Add(time.Minute),
		},
	}
	for name, example := range cases {
		t.Run(name, func(t *testing.T) {
			request := asked(t, 37, countingFrom(100))
			example.prepare(request)

			first := request.Check(rightCode, keyFrom(t, 0), example.at)
			second := request.Check(rightCode, keyFrom(t, 0), example.at)
			third := request.Check(rightCode, keyFrom(t, 0), example.at)

			requireInvalid(t, first, 2)
			requireInvalid(t, second, 1)
			requireEnded(t, third)
			_, ended := request.EndedAt()
			assert.True(t, ended)
		})
	}
}

// A browser that holds no request, because it asked for an e-mail of no
// active user or asked for none, is answered as a request that exists: the
// answer tells nothing of the e-mail.
func TestABrowserWithNoRequestIsAnsweredAsARequestThatNoLongerWorks(t *testing.T) {
	standIn := domain.StandInLoginCodeRequest(0)
	key := keyFrom(t, 0)

	requireInvalid(t, standIn.Check(rightCode, key, noon), 2)
	requireInvalid(t, standIn.Check("111111", key, noon), 1)
	requireEnded(t, standIn.Check(rightCode, key, noon))
	assert.Equal(t, 3, standIn.Attempts())

	again := domain.StandInLoginCodeRequest(1)
	requireInvalid(t, again.Check(rightCode, key, noon), 1)
	assert.Equal(t, 2, again.Attempts())
	assert.Equal(t, domain.NoLoginCodeRequest, again.ID())
}

func TestTheRulesOfTheCheckSayWhatToDo(t *testing.T) {
	assert.Contains(t, domain.LoginCodeInvalidError{AttemptsLeft: 2}.Error(), "attempts_left")
	assert.Contains(t, domain.LoginCodeEndedError{}.Error(), "ask for a new one")
}
