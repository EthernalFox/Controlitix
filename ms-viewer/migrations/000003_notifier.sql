-- +goose Up

CREATE TABLE IF NOT EXISTS alarms.notifications (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    alarm_event_id uuid NOT NULL,
    chat_id        bigint NOT NULL,
    status         smallint NOT NULL,
    attempt        smallint NOT NULL DEFAULT 0,
    last_error     text NULL,
    sent_at        timestamptz NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (alarm_event_id, chat_id)
);

CREATE INDEX IF NOT EXISTS notifications_pending_idx
    ON alarms.notifications (status, attempt)
    WHERE status = 0;

CREATE TABLE IF NOT EXISTS alarms.escalations (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    alarm_event_id uuid NOT NULL,
    tag_id         uuid NOT NULL,
    state          smallint NOT NULL,
    fire_at        timestamptz NOT NULL,
    processed      boolean NOT NULL DEFAULT false,
    skipped        boolean NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS escalations_due_idx
    ON alarms.escalations (fire_at)
    WHERE processed = false;

-- +goose Down

DROP TABLE IF EXISTS alarms.escalations;
DROP TABLE IF EXISTS alarms.notifications;
