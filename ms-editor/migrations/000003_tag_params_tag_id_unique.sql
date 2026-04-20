-- +goose Up

DROP INDEX IF EXISTS tags.tag_params_tag_id_idx;

CREATE UNIQUE INDEX IF NOT EXISTS tag_params_tag_id_uniq
    ON tags.tag_params (tag_id);

-- +goose Down

DROP INDEX IF EXISTS tags.tag_params_tag_id_uniq;

CREATE INDEX IF NOT EXISTS tag_params_tag_id_idx
    ON tags.tag_params (tag_id);

