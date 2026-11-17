// Package identifier writes the identifier of a record as it leaves the
// backend, and reads it back on the way in (ADR-0010): a prefix that names the
// type, then the Sqids encoding of the integer, at least 8 characters long,
// with an alphabet of the type's own. The domain, the database and the
// foreign keys keep the integer, which never leaves the backend as such.
//
// A prefix and an alphabet are constants and never change once an identifier
// was handed out: every link, bookmark and reference to it would break.
package identifier

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/sqids/sqids-go"
)

// minLength is the fewest characters of an encoding, after the prefix.
const minLength = 8

// File is the type of the identifier of a file.
var File = newType("fil", "9wg4cebEa5uNVtyY8G7rIxClZRoSXKT30pUhkWPiqOjJBnvF1zQfM2dLHA6smD")

// Type is a type of record whose identifier leaves the backend.
type Type struct {
	prefix string
	codec  *sqids.Sqids
}

func newType(prefix, alphabet string) Type {
	codec, err := sqids.New(sqids.Options{Alphabet: alphabet, MinLength: minLength})
	if err != nil {
		panic(fmt.Sprintf("the alphabet of the identifiers %s: %v", prefix, err))
	}
	return Type{prefix: prefix + "_", codec: codec}
}

// errNotOfThisType is why an identifier is refused: it has another prefix,
// or what follows the prefix is not the encoding of one number.
var errNotOfThisType = errors.New("the identifier is not one of this type")

// Encode writes the identifier of the record with this number. The numbers
// of the sequences start at 1, so only a positive number has an identifier.
func (t Type) Encode(number int64) (string, error) {
	if number <= 0 {
		return "", fmt.Errorf("encode the identifier %s of %d: only a positive number has one", t.prefix, number)
	}
	encoded, err := t.codec.Encode([]uint64{uint64(number)})
	if err != nil {
		return "", fmt.Errorf("encode the identifier %s: %w", t.prefix, err)
	}
	return t.prefix + encoded, nil
}

// Decode reads the number an identifier of this type stands for. It refuses
// one with another prefix before it decodes anything, and one the encoder
// would not have written, so that each record has a single identifier.
func (t Type) Decode(identifier string) (int64, error) {
	encoded, prefixed := strings.CutPrefix(identifier, t.prefix)
	if !prefixed {
		return 0, errNotOfThisType
	}
	numbers := t.codec.Decode(encoded)
	if len(numbers) != 1 || numbers[0] == 0 || numbers[0] > math.MaxInt64 {
		return 0, errNotOfThisType
	}
	again, err := t.codec.Encode(numbers)
	if err != nil || again != encoded {
		return 0, errNotOfThisType
	}
	return int64(numbers[0]), nil
}

// Pattern is the regular expression an identifier of this type follows, for
// the contract to declare: the prefix and at least 8 letters and digits. An
// identifier that follows it may still decode to no number.
func (t Type) Pattern() string {
	return fmt.Sprintf("^%s[0-9A-Za-z]{%d,}$", regexp.QuoteMeta(t.prefix), minLength)
}
