-- name: ListProviderLanguages :many
SELECT
    pl.tag,
    pl.name,
    pl.source,
    pl.target,
    pl.stable
FROM
    provider_languages pl
WHERE
    pl.provider_id = sqlc.arg (provider_id)
ORDER BY
    pl.tag;

-- name: InsertLanguage :exec
-- Tags are shared by every provider and never removed, because the history references them.
INSERT INTO
    languages (tag)
VALUES
    (sqlc.arg (tag)) ON CONFLICT DO NOTHING;

-- name: DeleteProviderLanguages :exec
DELETE FROM provider_languages
WHERE
    provider_id = sqlc.arg (provider_id);

-- name: InsertProviderLanguage :exec
-- Keeps the first of two provider codes that canonicalize to the same tag.
INSERT INTO
    provider_languages (provider_id, tag, name, source, target, stable)
VALUES
    (
        sqlc.arg (provider_id),
        sqlc.arg (tag),
        sqlc.arg (name),
        sqlc.arg (source),
        sqlc.arg (target),
        sqlc.arg (stable)
    ) ON CONFLICT DO NOTHING;

-- name: DeleteExpiredLanguages :execrows
-- Keeps the provider rows, which the history references.
DELETE FROM provider_languages
WHERE
    provider_id IN (
        SELECT
            id
        FROM
            providers
        WHERE
            languages_updated_at < sqlc.arg (expired_before)
    );