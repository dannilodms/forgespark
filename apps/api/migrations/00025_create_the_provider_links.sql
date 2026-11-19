-- +goose Up

-- The provider links (Login in CODING_STANDARDS.md, ADR-0011): each one
-- records that a user holds one account at a login provider, from their first
-- login through it, and later logins through that provider find the user by
-- it. The provider is named by its key, which configuration gives, so the
-- table checks the form of a key and lists none. The login never changes or
-- removes a link.
CREATE TABLE provider_links (
    provider_key text NOT NULL
        CONSTRAINT provider_links_provider_key_check
            CHECK (provider_key ~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$' AND length(provider_key) <= 64),
    -- The identifier the provider gives the account: the `sub` of its ID
    -- token.
    subject text NOT NULL
        CONSTRAINT provider_links_subject_check CHECK (length(subject) BETWEEN 1 AND 255),
    user_id bigint NOT NULL REFERENCES users,
    linked_at timestamptz NOT NULL,
    -- An account is linked to one user.
    PRIMARY KEY (provider_key, subject),
    -- A user has one account of each provider linked.
    CONSTRAINT provider_links_one_of_each_provider UNIQUE (provider_key, user_id)
);

GRANT SELECT, INSERT ON provider_links TO api;
