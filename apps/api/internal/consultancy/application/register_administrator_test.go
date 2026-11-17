package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// concluded is an installation whose setup was concluded with the first code,
// ten minutes after noon, with the clock ten minutes later still. Its users
// are the first administrator, and no setup code works.
func concluded(t *testing.T) installation {
	t.Helper()
	installed := issued(t)
	require.NoError(t, installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode)))
	installed.clock.Advance(10 * time.Minute)
	return installed
}
