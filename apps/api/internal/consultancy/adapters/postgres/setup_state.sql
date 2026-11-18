-- name: ReadSetupState :one
-- The setup is concluded once the consultancy exists. A setup code is waiting
-- while one was not ended and has not expired at this moment.
SELECT
    EXISTS (SELECT FROM consultancy) AS concluded,
    EXISTS (
        SELECT FROM setup_codes
        WHERE ended_at IS NULL AND expires_at > @at
    ) AS setup_code_waiting;
