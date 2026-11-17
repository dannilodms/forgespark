-- +goose Up

-- A file whose bytes are an object of the file storage (ADR-0017). Its
-- record is created when its upload starts, with what the sender declared,
-- and it is usable once its upload is confirmed. The object key is random and
-- holds neither the name of the file nor an identifier.
CREATE SEQUENCE files_id_seq AS bigint;

CREATE TABLE files (
    id bigint PRIMARY KEY,
    object_key text NOT NULL
        CONSTRAINT files_object_key_key UNIQUE,
    -- What the file is for.
    purpose text NOT NULL
        CONSTRAINT files_purpose_check CHECK (purpose IN ('consultancy_logo', 'user_photo')),
    -- The type and the size, in bytes, declared when the upload started.
    media_type text NOT NULL
        CONSTRAINT files_media_type_check CHECK (media_type IN ('image/png', 'image/jpeg', 'image/webp')),
    size bigint NOT NULL,
    started_at timestamptz NOT NULL,
    confirmed_at timestamptz
);

GRANT SELECT, INSERT, UPDATE ON files TO api;
GRANT USAGE ON SEQUENCE files_id_seq TO api;

-- The logo of the consultancy and the photo of a user, each a file.
ALTER TABLE consultancy ADD COLUMN logo_file_id bigint REFERENCES files;
ALTER TABLE users ADD COLUMN photo_file_id bigint REFERENCES files;
