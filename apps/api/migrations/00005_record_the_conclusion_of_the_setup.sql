-- +goose Up

-- Concluding the setup is a sensitive operation (ADR-0024). Its entry is
-- acted by the setup code used, and acts on the user created.
ALTER TABLE audit_log
    DROP CONSTRAINT audit_log_operation_check,
    ADD CONSTRAINT audit_log_operation_check
        CHECK (operation IN ('setup_code_issued', 'setup_concluded')),
    DROP CONSTRAINT audit_log_actor_kind_check,
    ADD CONSTRAINT audit_log_actor_kind_check
        CHECK (actor_kind IN ('installer', 'setup_code')),
    DROP CONSTRAINT audit_log_subject_kind_check,
    ADD CONSTRAINT audit_log_subject_kind_check
        CHECK (subject_kind IN ('setup_code', 'user'));
