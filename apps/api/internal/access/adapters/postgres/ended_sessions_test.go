package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/postgres"
)

// The store of the sessions removes none that ended: the daily cleanup
// removes the ones whose expiry passed, and keeps the others.
func TestTheSessionsThatEndedByAMomentAreRemoved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	for token, expiry := range map[string]time.Time{"ended": noon.Add(-time.Second), "now": noon, "open": noon.Add(7 * 24 * time.Hour)} {
		_, err := temporary.pool.Exec(context.Background(), "INSERT INTO sessions (token, data, expiry) VALUES ($1, $2, $3)", token, []byte{0}, expiry)
		require.NoError(t, err)
	}

	err := postgres.NewEndedSessions(temporary.pool).Remove(context.Background(), noon)

	require.NoError(t, err)
	rows, err := temporary.pool.Query(context.Background(), "SELECT token FROM sessions ORDER BY token")
	require.NoError(t, err)
	var kept []string
	for rows.Next() {
		var token string
		require.NoError(t, rows.Scan(&token))
		kept = append(kept, token)
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, []string{"now", "open"}, kept)
}
