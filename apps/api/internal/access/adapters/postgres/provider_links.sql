-- name: FindProviderLink :one
SELECT provider_key, subject, user_id, linked_at
FROM provider_links
WHERE provider_key = @provider_key AND subject = @subject;

-- name: HasProviderLink :one
SELECT EXISTS (
    SELECT 1 FROM provider_links WHERE provider_key = @provider_key AND user_id = @user_id
);

-- name: InsertProviderLink :execrows
-- Inserts nothing when the account, or the user at the provider, is linked
-- already. A link another transaction is inserting meanwhile is waited for:
-- once it commits, this one inserts nothing, and stays usable, so it can read
-- that link.
INSERT INTO provider_links (provider_key, subject, user_id, linked_at)
VALUES (@provider_key, @subject, @user_id, @linked_at)
ON CONFLICT DO NOTHING;
