-- +goose Up
CREATE SCHEMA IF NOT EXISTS history;

CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE history.tag_values_raw (
    tag_id uuid NOT NULL,
    ts timestamptz NOT NULL,
    v double precision NULL,
    q smallint NOT NULL,
    PRIMARY KEY (tag_id, ts)
);

SELECT create_hypertable('history.tag_values_raw', 'ts', chunk_time_interval => INTERVAL '1 day');

CREATE INDEX tag_values_raw_tag_ts_idx
    ON history.tag_values_raw (tag_id, ts DESC);

ALTER TABLE history.tag_values_raw SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'tag_id',
    timescaledb.compress_orderby = 'ts DESC'
);

SELECT add_compression_policy('history.tag_values_raw', INTERVAL '7 days');
SELECT add_retention_policy('history.tag_values_raw', INTERVAL '30 days');

CREATE MATERIALIZED VIEW history.tag_values_1m
WITH (timescaledb.continuous) AS
SELECT
    tag_id,
    time_bucket(INTERVAL '1 minute', ts) AS bucket,
    avg(v) AS avg,
    min(v) AS min,
    max(v) AS max,
    count(v) AS count,
    last(v, ts) AS last,
    last(q, ts) AS q
FROM history.tag_values_raw
GROUP BY tag_id, bucket
WITH NO DATA;

SELECT add_continuous_aggregate_policy('history.tag_values_1m',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

SELECT add_retention_policy('history.tag_values_1m', INTERVAL '365 days');

-- +goose Down
SELECT remove_retention_policy('history.tag_values_1m', if_exists => true);
SELECT remove_continuous_aggregate_policy('history.tag_values_1m', if_exists => true);
DROP MATERIALIZED VIEW IF EXISTS history.tag_values_1m;
SELECT remove_retention_policy('history.tag_values_raw', if_exists => true);
SELECT remove_compression_policy('history.tag_values_raw', if_exists => true);
DROP TABLE IF EXISTS history.tag_values_raw;
DROP SCHEMA IF EXISTS history;
