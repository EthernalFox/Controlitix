-- +goose Up

CREATE SCHEMA IF NOT EXISTS alarms;

CREATE TABLE IF NOT EXISTS alarms.states (
    tag_id       uuid PRIMARY KEY,
    state        smallint NOT NULL,
    last_value   double precision NULL,
    last_quality smallint NOT NULL,
    entered_at   timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    suppressed   boolean NOT NULL DEFAULT false,
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS states_state_idx ON alarms.states (state) WHERE state > 0;
CREATE INDEX IF NOT EXISTS states_last_seen_idx ON alarms.states (last_seen_at) WHERE state > 0;

CREATE TABLE IF NOT EXISTS alarms.events (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_id      uuid NOT NULL,
    event_type  smallint NOT NULL,
    state_from  smallint NOT NULL,
    state_to    smallint NOT NULL,
    value       double precision NULL,
    quality     smallint NOT NULL,
    ts          timestamptz NOT NULL,
    actor_id    text NULL,
    note        text NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS events_tag_ts_idx ON alarms.events (tag_id, ts DESC);
CREATE INDEX IF NOT EXISTS events_ts_idx ON alarms.events (ts DESC);

CREATE TABLE IF NOT EXISTS alarms.acks (
    tag_id    uuid PRIMARY KEY REFERENCES alarms.states(tag_id) ON DELETE CASCADE,
    state     smallint NOT NULL,
    actor_id  text NOT NULL,
    note      text NULL,
    acked_at  timestamptz NOT NULL DEFAULT now()
);

-- +goose Down

DROP TABLE IF EXISTS alarms.acks;
DROP TABLE IF EXISTS alarms.events;
DROP TABLE IF EXISTS alarms.states;
DROP SCHEMA IF EXISTS alarms;
