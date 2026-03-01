-- +goose Up
-- +goose StatementBegin
CREATE TABLE stops (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id      UUID        NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    location     TEXT,
    arrived_at   TIMESTAMPTZ NOT NULL,
    departed_at  TIMESTAMPTZ,
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stops;
-- +goose StatementEnd
