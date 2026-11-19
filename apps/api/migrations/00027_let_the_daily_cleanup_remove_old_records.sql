-- +goose Up

-- The daily cleanup of the worker (Jobs in CODING_STANDARDS.md) removes what
-- the login and the e-mail keep only for a while: the login code requests
-- ended or expired more than a day ago, and the bounce notices received more
-- than 30 days ago. It also removes the asks for a login code older than an
-- hour and the sessions that ended, which the API role could remove already.
-- The worker uses the API role.
GRANT DELETE ON login_code_requests TO api;
GRANT DELETE ON bounce_notices TO api;
