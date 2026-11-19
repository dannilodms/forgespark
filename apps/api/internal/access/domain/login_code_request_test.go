package domain_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
)

var noon = time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)

// countingFrom stands in for the random source: it gives first, first+1 and
// so on.
func countingFrom(first byte) *bytes.Reader {
	drawn := make([]byte, 64)
	for i := range drawn {
		drawn[i] = first + byte(i)
	}
	return bytes.NewReader(drawn)
}

// keyFrom is a key of 32 bytes: first, first+1 and so on.
func keyFrom(t *testing.T, first byte) domain.LoginCodeKey {
	t.Helper()
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = first + byte(i)
	}
	key, err := domain.NewLoginCodeKey(raw)
	require.NoError(t, err)
	return key
}

func asked(t *testing.T, id int64, random *bytes.Reader) *domain.LoginCodeRequest {
	t.Helper()
	request, err := domain.AskForLoginCode(domain.LoginCodeRequestID(id), domain.UserID(7), noon, random)
	require.NoError(t, err)
	return request
}

func TestARequestWorksForTenMinutesFromWhenItWasAskedFor(t *testing.T) {
	request := asked(t, 37, countingFrom(100))

	assert.Equal(t, noon, request.AskedAt())
	assert.Equal(t, noon.Add(10*time.Minute), request.ExpiresAt())
	assert.Equal(t, 0, request.Attempts())
	_, ended := request.EndedAt()
	assert.False(t, ended)
}

// The expected codes were computed apart from this code, with Python's hmac
// module: HMAC-SHA256 under the key, over the identifier as eight bytes,
// big-endian, followed by the random value, reduced to six digits by the
// dynamic truncation of RFC 4226.
func TestTheCodeIsSixDigitsComputedFromTheKeyTheIdentifierAndTheRandomValue(t *testing.T) {
	cases := map[string]struct {
		key    byte
		id     int64
		random byte
		code   string
	}{
		"a code with a leading zero": {key: 0, id: 37, random: 100, code: "025080"},
		"another identifier":         {key: 0, id: 38, random: 100, code: "294646"},
		"another key":                {key: 1, id: 37, random: 100, code: "737659"},
		"another random value":       {key: 0, id: 37, random: 101, code: "969712"},
	}
	for name, example := range cases {
		t.Run(name, func(t *testing.T) {
			request := asked(t, example.id, countingFrom(example.random))

			code := request.Code(keyFrom(t, example.key))

			assert.Equal(t, example.code, code.Reveal())
		})
	}
}

func TestTheCodeOfARequestRestoredFromWhatWasStoredIsTheSame(t *testing.T) {
	request := asked(t, 37, countingFrom(100))

	restored := domain.RestoreLoginCodeRequest(domain.StoredLoginCodeRequest{
		ID:          request.ID(),
		User:        request.User(),
		RandomValue: request.RandomValue(),
		AskedAt:     request.AskedAt(),
		ExpiresAt:   request.ExpiresAt(),
		Attempts:    request.Attempts(),
		EndedAt:     nil,
	})

	assert.Equal(t, request, restored)
	assert.Equal(t, "025080", restored.Code(keyFrom(t, 0)).Reveal())
}

func TestTheCodePrintsAsAPlaceholderEverywhereButReveal(t *testing.T) {
	code := asked(t, 37, countingFrom(100)).Code(keyFrom(t, 0))
	encoded, err := json.Marshal(map[string]any{"code": code})
	require.NoError(t, err)

	for _, printed := range []string{fmt.Sprint(code), fmt.Sprintf("%v %+v %#v %s", code, code, code, code), string(encoded)} {
		assert.NotContains(t, printed, "025080")
		assert.Contains(t, printed, "[login code]")
	}
}

func TestAKeyShorterThan32BytesIsRefused(t *testing.T) {
	_, err := domain.NewLoginCodeKey(make([]byte, 31))

	assert.Error(t, err)
}

func TestAnEndedRequestKeepsTheMomentItWasFirstEndedAt(t *testing.T) {
	request := asked(t, 37, countingFrom(100))

	request.End(noon.Add(time.Minute))
	request.End(noon.Add(2 * time.Minute))

	endedAt, ended := request.EndedAt()
	assert.True(t, ended)
	assert.Equal(t, noon.Add(time.Minute), endedAt)
}

func TestARandomSourceThatFailsAsksForNoCode(t *testing.T) {
	_, err := domain.AskForLoginCode(domain.LoginCodeRequestID(1), domain.UserID(7), noon, bytes.NewReader(make([]byte, 3)))

	assert.Error(t, err)
}
