-- name: NextUserID :one
SELECT nextval('users_id_seq')::bigint;

-- name: EmailInUse :one
-- A deactivated user keeps their e-mail. The e-mail is stored in lowercase,
-- and the parameter is too.
SELECT EXISTS (SELECT FROM users WHERE email = @email);

-- name: SaveUser :exec
-- The e-mail belongs to one user: another user's e-mail breaks the unique
-- constraint users_email_key.
INSERT INTO users (id, name, email, country, national_id, telephone, preferred_language, photo_file_id, deactivated)
VALUES (@id, @name, @email, @country, @national_id, @telephone, @preferred_language, @photo_file_id, @deactivated)
ON CONFLICT (id) DO UPDATE
SET name = excluded.name,
    email = excluded.email,
    country = excluded.country,
    national_id = excluded.national_id,
    telephone = excluded.telephone,
    preferred_language = excluded.preferred_language,
    photo_file_id = excluded.photo_file_id,
    deactivated = excluded.deactivated;

-- name: GiveUserRoles :exec
-- A role the user already has is left as it is. Taking a role away arrives
-- with the use case that does it.
INSERT INTO user_roles (user_id, role_id)
SELECT @user_id, unnest(@role_ids::bigint[])
ON CONFLICT DO NOTHING;
