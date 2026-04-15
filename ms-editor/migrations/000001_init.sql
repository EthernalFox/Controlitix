-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS public.objects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    description text NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS objects_created_at_idx ON public.objects (created_at);

CREATE TABLE IF NOT EXISTS public.mimic (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id uuid NOT NULL REFERENCES public.objects (id),
    name text NULL,
    description text NULL,
    published_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS mimic_object_id_idx ON public.mimic (object_id);

CREATE TABLE IF NOT EXISTS public.figures (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id uuid NOT NULL REFERENCES public.mimic (id),
    tag_id uuid NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS figures_scheme_id_idx ON public.figures (scheme_id);
CREATE INDEX IF NOT EXISTS figures_tag_id_idx ON public.figures (tag_id);

CREATE TABLE IF NOT EXISTS public.figure_params (
    figure_id uuid PRIMARY KEY REFERENCES public.figures (id),
    params jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS figure_params_params_gin_idx
    ON public.figure_params USING gin (params);

-- +goose Down
DROP TABLE IF EXISTS public.figure_params;
DROP TABLE IF EXISTS public.figures;
DROP TABLE IF EXISTS public.mimic;
DROP TABLE IF EXISTS public.objects;
