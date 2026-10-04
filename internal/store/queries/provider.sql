-- name: GetProvider :one
SELECT
    id,
    languages_updated_at
FROM
    providers
WHERE
    id = sqlc.arg (id);

-- name: UpsertProvider :exec
-- Records when the provider's languages were fetched and creates its row on the first fetch.
INSERT INTO
    providers (id, languages_updated_at)
VALUES
    (sqlc.arg (id), sqlc.arg (languages_updated_at)) ON CONFLICT (id) DO
UPDATE
SET
    languages_updated_at = excluded.languages_updated_at;