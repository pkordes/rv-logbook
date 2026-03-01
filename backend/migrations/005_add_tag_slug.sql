-- +goose Up
-- +goose StatementBegin

-- Add slug as the canonical identity for tags.
-- Slug is always lowercase and hyphenated (e.g. "rocky-mountains").
-- It is the unique key used for upserts — case-insensitive deduplication
-- is achieved by normalizing all input to a slug before touching the DB.
--
-- The existing UNIQUE on name is dropped: slug is the identity, not name.
-- Name preserves the original display text from whoever created the tag first.
ALTER TABLE tags
    ADD COLUMN slug TEXT NOT NULL DEFAULT '',
    ADD CONSTRAINT tags_slug_key UNIQUE (slug);

ALTER TABLE tags
    DROP CONSTRAINT tags_name_key;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tags
    DROP CONSTRAINT tags_slug_key,
    DROP COLUMN slug;

ALTER TABLE tags
    ADD CONSTRAINT tags_name_key UNIQUE (name);
-- +goose StatementEnd
