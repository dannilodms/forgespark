-- name: NextConsultancyID :one
SELECT nextval('consultancy_id_seq')::bigint;

-- name: ConsultancyExists :one
SELECT EXISTS (SELECT FROM consultancy);

-- name: SaveConsultancy :exec
-- The table holds one row at most: a second consultancy breaks the unique
-- index consultancy_only_one.
INSERT INTO consultancy (
    id, name, legal_name, country, tax_id, time_zone, currency, email, telephone, website,
    address_line1, address_line2, address_city, address_region, address_postal_code
)
VALUES (
    @id, @name, @legal_name, @country, @tax_id, @time_zone, @currency, @email, @telephone, @website,
    @address_line1, @address_line2, @address_city, @address_region, @address_postal_code
)
ON CONFLICT (id) DO UPDATE
SET name = excluded.name,
    legal_name = excluded.legal_name,
    country = excluded.country,
    tax_id = excluded.tax_id,
    time_zone = excluded.time_zone,
    currency = excluded.currency,
    email = excluded.email,
    telephone = excluded.telephone,
    website = excluded.website,
    address_line1 = excluded.address_line1,
    address_line2 = excluded.address_line2,
    address_city = excluded.address_city,
    address_region = excluded.address_region,
    address_postal_code = excluded.address_postal_code;
