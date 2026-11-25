package clock_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/clock"
)

func TestTheSystemClockTellsTheTimeOfTheMachineInUTCToTheMicrosecond(t *testing.T) {
	before := time.Now()

	now := clock.System{}.Now()

	assert.WithinRange(t, now, before.Add(-time.Millisecond), time.Now())
	assert.Equal(t, time.UTC, now.Location())
	assert.Zero(t, now.Nanosecond()%1000, "PostgreSQL keeps no more than microseconds")
}
