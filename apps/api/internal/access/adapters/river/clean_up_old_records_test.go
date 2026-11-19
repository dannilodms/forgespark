package river_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	accessriver "github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/river"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/jobs"
)

// The job of the daily cleanup names nothing: the River panel and the
// telemetry show the arguments of a job. What it removes is tested on the
// use case, and through the `worker` command, which runs it when it starts.
func TestTheJobOfTheDailyCleanupCarriesNoArgumentAndRunsInTheLightQueue(t *testing.T) {
	arguments, err := json.Marshal(accessriver.CleanUpOldRecordsArgs{})
	require.NoError(t, err)

	assert.JSONEq(t, `{}`, string(arguments))
	assert.Equal(t, "clean_up_old_records", accessriver.CleanUpOldRecordsArgs{}.Kind())
	assert.Equal(t, jobs.QueueLight, accessriver.CleanUpOldRecordsArgs{}.InsertOpts().Queue)
}
