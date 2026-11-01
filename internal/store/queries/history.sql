-- name: AppendTranslation :one
-- Fails if the provider has no row yet, which UpsertProvider creates when its languages are stored.
INSERT INTO
    translations (
        text,
        translation,
        provider_id,
        source_tag,
        target_tag,
        detection
    )
VALUES
    (
        sqlc.arg (text),
        sqlc.arg (translation),
        sqlc.arg (provider_id),
        sqlc.narg (source_tag),
        sqlc.arg (target_tag),
        sqlc.arg (detection)
    ) RETURNING id,
    created_at;

-- name: LatestTranslation :one
SELECT
    *
FROM
    translations
ORDER BY
    id DESC
LIMIT
    1;

-- name: PreviousTranslation :one
-- Only the newest window_size entries are reachable.
SELECT
    *
FROM
    translations h
WHERE
    h.id < sqlc.arg (id)
    AND h.id > (
        SELECT
            MAX(t.id)
        FROM
            translations t
    ) - sqlc.arg (window_size)
ORDER BY
    h.id DESC
LIMIT
    1;

-- name: NextTranslation :one
SELECT
    *
FROM
    translations
WHERE
    id > sqlc.arg (id)
ORDER BY
    id ASC
LIMIT
    1;

-- name: HasPreviousTranslation :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            translations t
        WHERE
            t.id < sqlc.arg (id)
            AND t.id > (
                SELECT
                    MAX(m.id)
                FROM
                    translations m
            ) - sqlc.arg (window_size)
    );

-- name: HasNextTranslation :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            translations
        WHERE
            id > sqlc.arg (id)
    );