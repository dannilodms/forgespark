-- +goose Up

-- The cleanup of the storage removes the record of an upload never confirmed
-- a day after it started (Files in CODING_STANDARDS.md). The API role may
-- remove a file; it removes only those, and a confirmed file is never
-- removed.
GRANT DELETE ON files TO api;

-- The cleanup looks for the uploads never confirmed every hour. They are few
-- beside the confirmed files, which this index leaves out.
CREATE INDEX files_started_at_unconfirmed_idx ON files (started_at) WHERE confirmed_at IS NULL;
