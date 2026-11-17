-- +goose Up

-- Registering an administrator with a setup code, the way back in, is a
-- sensitive operation (ADR-0024). Its entry is acted by the setup code used,
-- and acts on the user created: kinds the audit log already takes.
ALTER TABLE audit_log
    DROP CONSTRAINT audit_log_operation_check,
    ADD CONSTRAINT audit_log_operation_check
        CHECK (operation IN ('setup_code_issued', 'setup_concluded', 'administrator_registered'));
