-- +goose Up

-- The consultancy, its users and its roles (ADR-0024, ADR-0026). The setup
-- saves them together. Every text is stored as typed; the rules of each
-- field are in the domain.

-- The software company the system is run for. There is exactly one: the
-- database refuses a second row.
CREATE SEQUENCE consultancy_id_seq AS bigint;

CREATE TABLE consultancy (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    legal_name text NOT NULL,
    -- ISO 3166-1 alpha-2.
    country text NOT NULL,
    tax_id text NOT NULL,
    -- A name of the IANA time zone database.
    time_zone text NOT NULL,
    -- ISO 4217.
    currency text NOT NULL,
    email text,
    -- A plus sign and digits, with no space.
    telephone text,
    website text,
    -- The address: every part may be empty, and an address with no part is
    -- none.
    address_line1 text,
    address_line2 text,
    address_city text,
    address_region text,
    address_postal_code text
);

CREATE UNIQUE INDEX consultancy_only_one ON consultancy ((true));

GRANT SELECT, INSERT, UPDATE ON consultancy TO api;
GRANT USAGE ON SEQUENCE consultancy_id_seq TO api;

-- A role is a named set of scopes. The two the setup creates carry a fixed
-- key, which the frontend translates. No scope exists yet, so no table links
-- a role to its scopes: the administrator role grants all of them by rule.
CREATE SEQUENCE roles_id_seq AS bigint;

CREATE TABLE roles (
    id bigint PRIMARY KEY,
    key text NOT NULL
        CONSTRAINT roles_key_key UNIQUE
        CONSTRAINT roles_key_check CHECK (key IN ('administrator', 'developer'))
);

GRANT SELECT, INSERT, UPDATE ON roles TO api;
GRANT USAGE ON SEQUENCE roles_id_seq TO api;

-- A person registered in the system, who logs in with their e-mail. The
-- e-mail is stored in lowercase and belongs to one user. The preferred
-- language has no CHECK: adding a language never changes the schema
-- (ADR-0005).
CREATE SEQUENCE users_id_seq AS bigint;

CREATE TABLE users (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL
        CONSTRAINT users_email_key UNIQUE,
    country text NOT NULL,
    national_id text NOT NULL,
    telephone text,
    preferred_language text NOT NULL,
    deactivated boolean NOT NULL
);

GRANT SELECT, INSERT, UPDATE ON users TO api;
GRANT USAGE ON SEQUENCE users_id_seq TO api;

-- Which user has which role. A user has at least one; the domain keeps that.
CREATE TABLE user_roles (
    user_id bigint NOT NULL REFERENCES users,
    role_id bigint NOT NULL REFERENCES roles,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX user_roles_role_id ON user_roles (role_id);

GRANT SELECT, INSERT ON user_roles TO api;
