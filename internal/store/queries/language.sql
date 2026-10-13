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
-- Tags are shared by every provider. DeleteOrphanLanguages removes those nothing references any more.
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

-- name: DeleteOrphanLanguages :execrows
-- Deletes languages that neither a provider nor the history references.
-- source_tag is filtered for NULL, because a NULL inside NOT IN matches nothing.
DELETE FROM languages
WHERE
    tag NOT IN (
        SELECT
            tag
        FROM
            provider_languages
    )
    AND tag NOT IN (
        SELECT
            target_tag
        FROM
            translations
    )
    AND tag NOT IN (
        SELECT
            source_tag
        FROM
            translations
        WHERE
            source_tag IS NOT NULL
    );

-- name: UpsertProvider :exec
-- Records when the provider's languages were fetched and creates its row on the first fetch.
INSERT INTO
    providers (id, languages_updated_at)
VALUES
    (sqlc.arg (id), sqlc.arg (languages_updated_at)) ON CONFLICT (id) DO
UPDATE
SET
    languages_updated_at = excluded.languages_updated_at;

-- name: ListLanguagesUpdatedAt :many
-- A provider without a row has never been fetched and is left out.
SELECT
    id,
    languages_updated_at
FROM
    providers
WHERE
    id IN (sqlc.slice (ids))
ORDER BY
    id;
