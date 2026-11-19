-- name: FindProviderLink :one
SELECT provider_key, subject, user_id, linked_at
FROM provider_links
WHERE provider_key = @provider_key AND subject = @subject;

-- name: HasProviderLink :one
SELECT EXISTS (
    SELECT 1 FROM provider_links WHERE provider_key = @provider_key AND user_id = @user_id
);

-- name: InsertProviderLink :exec
INSERT INTO provider_links (provider_key, subject, user_id, linked_at)
VALUES (@provider_key, @subject, @user_id, @linked_at);
