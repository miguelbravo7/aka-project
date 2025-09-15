-- name: CreateCharacter :one
INSERT INTO characters (id, name, status, species, type, gender, image, url, created, origin_id, location_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetMissingCharacterIDs :many
WITH input_ids AS (
    SELECT UNNEST($1::int[]) AS id
)
SELECT id::int4 AS id
FROM input_ids
WHERE id NOT IN (
    SELECT id FROM characters
);