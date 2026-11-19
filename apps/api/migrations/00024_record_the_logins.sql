-- +goose Up

-- Logging in is a sensitive operation (ADR-0012). Its entry is acted by the
-- user who logged in, on themselves, and keeps the way in: `login_code`, or
-- the key of the login provider the user came through. A provider is
-- configured, so its key is not listed: a login has a way in, and any other
-- operation has none.
ALTER TABLE audit_log
    ADD COLUMN way_in text,
    DROP CONSTRAINT audit_log_operation_check,
    ADD CONSTRAINT audit_log_operation_check
        CHECK (operation IN ('setup_code_issued', 'setup_concluded', 'administrator_registered', 'logged_in')),
    DROP CONSTRAINT audit_log_actor_kind_check,
    ADD CONSTRAINT audit_log_actor_kind_check
        CHECK (actor_kind IN ('installer', 'setup_code', 'user')),
    ADD CONSTRAINT audit_log_way_in_check
        CHECK ((operation = 'logged_in') = (way_in IS NOT NULL));
