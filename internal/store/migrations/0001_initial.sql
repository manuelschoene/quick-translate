CREATE TABLE
    "providers" (
        id INTEGER PRIMARY KEY, -- The provider's enum value in Go.
        languages_updated_at INTEGER NOT NULL
    );

CREATE TABLE
    "languages" (tag TEXT PRIMARY KEY NOT NULL);

CREATE TABLE
    "provider_languages" (
        provider_id INTEGER NOT NULL REFERENCES providers (id) ON DELETE CASCADE,
        tag TEXT NOT NULL REFERENCES languages (tag),
        name TEXT NOT NULL,
        source BOOLEAN NOT NULL CHECK (source IN (0, 1)), -- BOOLEAN only for sqlc's bool mapping; SQLite stores 0 or 1.
        target BOOLEAN NOT NULL CHECK (target IN (0, 1)),
        stable BOOLEAN NOT NULL CHECK (stable IN (0, 1)),
        PRIMARY KEY (provider_id, tag)
    );

CREATE TABLE
    "translations" (
        id INTEGER PRIMARY KEY,
        text TEXT NOT NULL,
        translation TEXT NOT NULL,
        created_at INTEGER NOT NULL DEFAULT (unixepoch()),
        provider_id INTEGER NOT NULL REFERENCES providers (id),
        source_tag TEXT REFERENCES languages (tag), -- The detected language when detection is set; NULL if it could not be resolved.
        target_tag TEXT NOT NULL REFERENCES languages (tag),
        detection BOOLEAN NOT NULL CHECK (detection IN (0, 1)),
        CHECK (
            detection = 1
            OR source_tag IS NOT NULL
        )
    );