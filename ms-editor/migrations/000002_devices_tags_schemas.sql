-- +goose Up

-- ============================================================
-- 1. Fix existing public tables
-- ============================================================

ALTER TABLE public.figures
    ADD COLUMN IF NOT EXISTS type text NOT NULL DEFAULT 'rect';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'figures'
          AND column_name = 'scheme_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'figures'
          AND column_name = 'diagram_id'
    ) THEN
        ALTER TABLE public.figures RENAME COLUMN scheme_id TO diagram_id;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE schemaname = 'public'
          AND indexname = 'figures_scheme_id_idx'
    ) AND NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE schemaname = 'public'
          AND indexname = 'figures_diagram_id_idx'
    ) THEN
        ALTER INDEX public.figures_scheme_id_idx RENAME TO figures_diagram_id_idx;
    END IF;
END $$;

ALTER TABLE public.objects
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

ALTER TABLE public.mimic
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

ALTER TABLE public.figures
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

-- ============================================================
-- 2. Create schemas
-- ============================================================

CREATE SCHEMA IF NOT EXISTS devices;
CREATE SCHEMA IF NOT EXISTS tags;

-- ============================================================
-- 3. Reference tables
-- ============================================================

CREATE TABLE IF NOT EXISTS devices.device_type (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

INSERT INTO devices.device_type (name) VALUES
    ('modbus_rtu'),
    ('modbus_tcp'),
    ('snmp_v1'),
    ('snmp_v2c'),
    ('snmp_v3')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS tags.data_types (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

INSERT INTO tags.data_types (name) VALUES
    ('bool'),
    ('int8'),
    ('uint8'),
    ('int16'),
    ('uint16'),
    ('int32'),
    ('uint32'),
    ('float32'),
    ('float64'),
    ('string')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS tags.units (
    id serial PRIMARY KEY,
    name text NOT NULL,
    symbol text NOT NULL UNIQUE,
    category text NOT NULL
);

INSERT INTO tags.units (name, symbol, category) VALUES
    ('Volt', 'V', 'electrical'),
    ('Millivolt', 'mV', 'electrical'),
    ('Kilovolt', 'kV', 'electrical'),
    ('Ampere', 'A', 'electrical'),
    ('Milliampere', 'mA', 'electrical'),
    ('Watt', 'W', 'electrical'),
    ('Kilowatt', 'kW', 'electrical'),
    ('Megawatt', 'MW', 'electrical'),
    ('Kilowatt-hour', 'kWh', 'electrical'),
    ('Megawatt-hour', 'MWh', 'electrical'),
    ('Volt-Ampere', 'VA', 'electrical'),
    ('Kilovolt-Ampere', 'kVA', 'electrical'),
    ('Hertz', 'Hz', 'electrical'),
    ('Ohm', U&'\03A9', 'electrical'),
    ('Kilohm', U&'k\03A9', 'electrical'),
    ('Celsius', U&'\00B0C', 'temperature'),
    ('Fahrenheit', U&'\00B0F', 'temperature'),
    ('Kelvin', 'K', 'temperature'),
    ('Pascal', 'Pa', 'pressure'),
    ('Kilopascal', 'kPa', 'pressure'),
    ('Megapascal', 'MPa', 'pressure'),
    ('Bar', 'bar', 'pressure'),
    ('Millibar', 'mbar', 'pressure'),
    ('Relative humidity', '%RH', 'humidity'),
    ('Percent', '%', 'percentage'),
    ('Revolutions per minute', 'RPM', 'frequency'),
    ('Second', 's', 'time'),
    ('Millisecond', 'ms', 'time'),
    ('Minute', 'min', 'time'),
    ('Hour', 'h', 'time')
ON CONFLICT (symbol) DO NOTHING;

-- ============================================================
-- 4. Devices tables
-- ============================================================

CREATE TABLE IF NOT EXISTS devices.devices (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id uuid NULL REFERENCES public.objects (id),
    type_id int NOT NULL REFERENCES devices.device_type (id),
    name text NOT NULL,
    description text NULL,
    deleted_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS devices_object_id_idx ON devices.devices (object_id);
CREATE INDEX IF NOT EXISTS devices_type_id_idx ON devices.devices (type_id);
CREATE UNIQUE INDEX IF NOT EXISTS devices_object_id_name_uniq
    ON devices.devices (object_id, name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS devices.devices_params (
    device_id uuid PRIMARY KEY REFERENCES devices.devices (id) ON DELETE CASCADE,
    settings jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- ============================================================
-- 5. Tags tables
-- ============================================================

CREATE TABLE IF NOT EXISTS tags.tags (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id uuid NOT NULL REFERENCES devices.devices (id),
    name text NOT NULL,
    description text NULL,
    deleted_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tags_device_id_idx ON tags.tags (device_id);
CREATE UNIQUE INDEX IF NOT EXISTS tags_device_id_name_uniq
    ON tags.tags (device_id, name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS tags.tag_params (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_id uuid NOT NULL REFERENCES tags.tags (id) ON DELETE CASCADE,
    data_type_id int NOT NULL REFERENCES tags.data_types (id),
    unit_id int NULL REFERENCES tags.units (id),
    address jsonb NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tag_params_tag_id_idx ON tags.tag_params (tag_id);

CREATE TABLE IF NOT EXISTS tags.tag_setpoints (
    param_id uuid PRIMARY KEY REFERENCES tags.tag_params (id) ON DELETE CASCADE,
    lolo double precision NULL,
    lo double precision NULL,
    hi double precision NULL,
    hihi double precision NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tags.tag_scaling (
    param_id uuid PRIMARY KEY REFERENCES tags.tag_params (id) ON DELETE CASCADE,
    raw_min double precision NULL,
    raw_max double precision NULL,
    eng_min double precision NULL,
    eng_max double precision NULL,
    factor double precision NULL,
    "offset" double precision NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down

DROP TABLE IF EXISTS tags.tag_scaling;
DROP TABLE IF EXISTS tags.tag_setpoints;
DROP TABLE IF EXISTS tags.tag_params;
DROP TABLE IF EXISTS tags.tags;
DROP TABLE IF EXISTS devices.devices_params;
DROP TABLE IF EXISTS devices.devices;
DROP TABLE IF EXISTS tags.units;
DROP TABLE IF EXISTS tags.data_types;
DROP TABLE IF EXISTS devices.device_type;
DROP SCHEMA IF EXISTS tags;
DROP SCHEMA IF EXISTS devices;

ALTER TABLE public.figures
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE public.mimic
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE public.objects
    DROP COLUMN IF EXISTS deleted_at;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'figures'
          AND column_name = 'diagram_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'figures'
          AND column_name = 'scheme_id'
    ) THEN
        ALTER TABLE public.figures RENAME COLUMN diagram_id TO scheme_id;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE schemaname = 'public'
          AND indexname = 'figures_diagram_id_idx'
    ) AND NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE schemaname = 'public'
          AND indexname = 'figures_scheme_id_idx'
    ) THEN
        ALTER INDEX public.figures_diagram_id_idx RENAME TO figures_scheme_id_idx;
    END IF;
END $$;

ALTER TABLE public.figures
    DROP COLUMN IF EXISTS type;
