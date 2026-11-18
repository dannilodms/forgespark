package identifier_test

import (
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/identifier"
)

func TestTheIdentifierOfAFileIsItsPrefixAndAtLeastEightCharactersThatDecodeBack(t *testing.T) {
	t.Parallel()
	for _, number := range []int64{1, 2, 63, 1000, 1 << 40, math.MaxInt64} {
		encoded, err := identifier.File.Encode(number)
		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(encoded, "fil_"), encoded)
		assert.GreaterOrEqual(t, len(strings.TrimPrefix(encoded, "fil_")), 8, encoded)
		decoded, err := identifier.File.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, number, decoded)
	}
}

// A prefix and an alphabet never change once an identifier was handed out
// (ADR-0010): every link to a file would break.
func TestTheIdentifierOfAFileNeverChanges(t *testing.T) {
	t.Parallel()
	first, err := identifier.File.Encode(1)
	require.NoError(t, err)
	thousandth, err := identifier.File.Encode(1000)
	require.NoError(t, err)

	// Written down the first time they were encoded.
	assert.Equal(t, "fil_sn2TPfBW", first)
	assert.Equal(t, "fil_KBa6y1qO", thousandth)
}

func TestTwoFilesHaveTwoIdentifiers(t *testing.T) {
	t.Parallel()
	first, err := identifier.File.Encode(1)
	require.NoError(t, err)
	second, err := identifier.File.Encode(2)
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

func TestOnlyAPositiveNumberIsEncoded(t *testing.T) {
	t.Parallel()
	for _, number := range []int64{0, -1, math.MinInt64} {
		_, err := identifier.File.Encode(number)

		assert.Error(t, err, "%d was encoded", number)
	}
}

func TestAnIdentifierThatIsNotOfAFileIsRefused(t *testing.T) {
	t.Parallel()
	encoded, err := identifier.File.Encode(1)
	require.NoError(t, err)
	body := strings.TrimPrefix(encoded, "fil_")

	refused := map[string]string{
		"with no prefix":                         body,
		"with the prefix of another":             "prj_" + body,
		"with the prefix in upper case":          "FIL_" + body,
		"with nothing after the prefix":          "fil_",
		"empty":                                  "",
		"with a character outside the alphabet":  "fil_" + body[:7] + "-",
		"that the encoder would write otherwise": "fil_" + strings.Repeat(body[:1], 8),
		// What Sqids writes for the numbers 1 and 2 with the alphabet of a
		// file: a text the encoder of one number never writes.
		"of two numbers": "fil_Q0WBEBTq",
	}
	for name, typed := range refused {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := identifier.File.Decode(typed)

			assert.Error(t, err)
		})
	}
}

func TestThePatternOfAFileIdentifierMatchesWhatTheEncoderWrites(t *testing.T) {
	t.Parallel()
	pattern := regexp.MustCompile(identifier.File.Pattern())
	encoded, err := identifier.File.Encode(42)
	require.NoError(t, err)

	assert.Regexp(t, pattern, encoded)
	assert.NotRegexp(t, pattern, "prj_"+strings.TrimPrefix(encoded, "fil_"))
	assert.NotRegexp(t, pattern, "fil_short")
	assert.NotRegexp(t, pattern, encoded+" ")
}
