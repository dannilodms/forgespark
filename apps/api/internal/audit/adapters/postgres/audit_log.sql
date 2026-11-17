-- name: AppendAuditLogEntry :exec
INSERT INTO audit_log (occurred_at, operation, actor_kind, actor_id, subject_kind, subject_id)
VALUES (@occurred_at, @operation, @actor_kind, @actor_id, @subject_kind, @subject_id);
