-- +goose Up
-- +goose StatementBegin
CREATE TABLE stop_tags (
    stop_id UUID NOT NULL REFERENCES stops(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tags(id)  ON DELETE CASCADE,
    PRIMARY KEY (stop_id, tag_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stop_tags;
-- +goose StatementEnd
